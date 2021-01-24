package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jecklgamis/gatling-server/pkg/accesslog"
	"github.com/jecklgamis/gatling-server/pkg/env"
	"github.com/jecklgamis/gatling-server/pkg/event"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/handler"
	"github.com/jecklgamis/gatling-server/pkg/heartbeat"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	"github.com/jecklgamis/gatling-server/pkg/uploader"
	"github.com/jecklgamis/gatling-server/pkg/version"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func printRoutes(router *mux.Router) {
	log.Println("Available endpoints:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		template, err := route.GetPathTemplate()
		if err == nil {
			log.Println(template)
		}
		return nil
	})
}

func Start() {
	appEnv := env.GetOrElse("APP_ENVIRONMENT", "dev")
	config := ReadConfig(appEnv)

	router := mux.NewRouter()
	router.HandleFunc("/buildInfo", handler.BuildInfoHandler)
	router.HandleFunc("/probe/ready", handler.ReadinessProbeHandler)
	router.HandleFunc("/probe/live", handler.LivenessProbeHandler)

	eventBus := event.NewEventBus()
	if config.Heartbeat.Enabled {
		heartbeat.New(config.Heartbeat.Frequency, func() { eventBus.EventC <- event.NewHeartbeatEvent() })
	}
	configureEventNotifiers(eventBus, config.EventNotifiers)
	uploaders := configureUploaders(config.Uploaders)

	gatlingDir, _ := filepath.Abs(config.GatlingDir)
	gatling := gatling.NewGatling(gatlingDir)
	taskManager := taskmanager.NewTaskManager(gatling, eventBus.EventC, uploaders)
	workspaceDir, _ := filepath.Abs(config.WorkspaceDir)
	workspace := workspace.NewWorkspace(workspaceDir)
	if workspace == nil {
		panic("unable to create workspace")
	}
	uploadDir, _ := filepath.Abs(config.UploadDir)
	log.Println("Using upload dir", uploadDir)
	httpUploadHandler := handler.NewHttpUploadHandler(workspace, taskManager, uploadDir)
	router.HandleFunc("/task/upload/http", httpUploadHandler.Handle)

	s3Config, found := config.Downloaders["s3"]
	if found && s3Config.Enabled {
		region, found := s3Config.ConfigMap["region"]
		if found {
			s3ops := s3.NewS3Manager(region)
			s3DownloadHandler := handler.NewS3DownloadHandler(workspace, taskManager, s3ops)
			router.HandleFunc("/task/download/s3", s3DownloadHandler.Handle)
		} else {
			log.Println("S3 downloader missing region config")
		}
	}
	taskHandler := handler.NewTaskHandler(workspace, taskManager)
	router.HandleFunc("/task/{taskId}", taskHandler.TaskContextHandler)
	router.HandleFunc("/task/metadata/{taskId}", taskHandler.MetadataHandler)
	router.HandleFunc("/task/console/{taskId}", taskHandler.ConsoleLogHandler)
	router.HandleFunc("/task/results/{taskId}", taskHandler.ResultsHandler)
	router.HandleFunc("/task/simulationLog/{taskId}", taskHandler.SimulationLogHandler)
	router.HandleFunc("/task/abort/{taskId}", taskHandler.AbortTaskHandler)
	router.HandleFunc("/blackhole", handler.BlackholeHandler)

	fs := http.FileServer(http.Dir(workspace.BaseDir() + "/"))
	router.PathPrefix("/").Handler(fs)
	printRoutes(router)
	router.Use(accesslog.AccessLoggerMiddleware)

	log.Printf("Version: %s\n", version.BuildVersion)
	if config.Server.Https.KeyFile != "" && config.Server.Https.CertFile != "" {
		go func() {
			addr := fmt.Sprintf("0.0.0.0:%d", config.Server.Https.Port)
			log.Printf("Starting HTTPS server on %s\n", addr)
			srv := &http.Server{
				Handler:      router,
				Addr:         addr,
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  15 * time.Second,
			}
			log.Fatal(srv.ListenAndServeTLS(config.Server.Https.CertFile, config.Server.Https.KeyFile))
		}()
	}
	if config.Server.Http.Port > 0 {
		go func() {
			addr := fmt.Sprintf("0.0.0.0:%d", config.Server.Http.Port)
			log.Printf("Starting HTTP server on %s\n", addr)
			srv := &http.Server{
				Handler:      router,
				Addr:         addr,
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  15 * time.Second,
			}
			log.Fatal(srv.ListenAndServe())
		}()
	}
	for {
		time.Sleep(time.Second)
	}
}

func configureEventNotifiers(eventBus *event.Bus, configs []EventNotifierConfig) {
	for _, config := range configs {
		if !config.Enabled {
			continue
		}
		switch config.Type {
		case "http":
			listener := event.NewHTTPNotifier(config.ConfigMap)
			if listener != nil {
				eventBus.RegisterListener(listener)
			}
		case "sns":
			sns := event.CreateSNS(config.ConfigMap["region"])
			if sns != nil {
				listener := event.NewSNSEventNotifier(sns, config.ConfigMap)
				if listener != nil {
					eventBus.RegisterListener(listener)
				}
			}
		default:
			log.Println("Unsupported event notifier type", config.Type)
		}
	}
}

func configureUploaders(configs []UploaderConfig) []uploader.GatlingArtifactUploader {
	var uploaders = make([]uploader.GatlingArtifactUploader, 0)
	for _, config := range configs {
		if !config.Enabled {
			continue
		}
		switch config.Type {
		case "s3":
			uploader := uploader.NewS3Uploader(s3.NewS3Manager(config.ConfigMap["region"]), config.ConfigMap)
			if uploader != nil {
				uploaders = append(uploaders, uploader)
			}
		default:
			log.Println("Unsupported uploader type", config.Type)
		}
	}
	return uploaders
}
