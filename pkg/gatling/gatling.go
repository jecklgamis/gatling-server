package gatling

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Ops interface {
	RunSimulation(commandOps cmdexec.CommandExecutionOps, task *Task) (*exec.Cmd, error)
}

type RunSimulationFunc func(commandOps cmdexec.CommandExecutionOps, task *Task) (*exec.Cmd, error)

func (r RunSimulationFunc) RunSimulation(commandOps cmdexec.CommandExecutionOps, task *Task) (*exec.Cmd, error) {
	return r(commandOps, task)
}

type Gatling struct {
	ScriptsDir string
}

func NewGatling(scriptsDir string) *Gatling {
	if !filepath.IsAbs(scriptsDir) {
		slog.Error("Scripts dir is not absolute", "scriptsDir", scriptsDir)
		return nil
	}
	slog.Info("Using scripts dir", "scriptsDir", scriptsDir)
	return &Gatling{scriptsDir}
}

type Task struct {
	Id           string
	UserFilesDir *workspace.UserFilesDir
	Simulation   string
	JavaOpts     string
	Tags         map[string]string
	FileType     string
}

type Result struct {
	Ok bool
}

func NewTask(id string, simulation string, javaOpts string, userFilesDir *workspace.UserFilesDir) *Task {
	return &Task{Id: id, Simulation: simulation, JavaOpts: javaOpts, UserFilesDir: userFilesDir}
}

func (g *Gatling) RunSimulation(commandOps cmdexec.CommandExecutionOps, task *Task) (*exec.Cmd, error) {
	slog.Info("Running simulations from", "baseDir", task.UserFilesDir.BaseDir)
	userFilesDir := task.UserFilesDir
	gatlingSh := fmt.Sprintf("%s/gatling-jar-runner.sh", g.ScriptsDir)
	cmd := exec.Command(gatlingSh, "-s", task.Simulation,
		"--results-folder", userFilesDir.Results)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_OPTS=%s", task.JavaOpts))
	cmd.Env = append(cmd.Env, fmt.Sprintf("JAR_FILE=%s/*", userFilesDir.Simulations))
	slog.Info("JAVA_OPTS", "javaOpts", task.JavaOpts)
	slog.Info("About to execute command", "cmd", cmd)
	err := commandOps.ExecuteAndLog(cmd, filepath.Join(task.UserFilesDir.BaseDir, "console.log"), task.Id)
	return cmd, err
}
