package gatling

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"log"
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
	BaseDir string
}

func NewGatling(baseDir string) *Gatling {
	if !filepath.IsAbs(baseDir) {
		log.Println("Base dir is not absolute", baseDir)
		return nil
	}
	log.Println("Using gatling distribution", baseDir)
	return &Gatling{baseDir}
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
	log.Println("Running simulations from", task.UserFilesDir.BaseDir)
	userFilesDir := task.UserFilesDir
	var cmd *exec.Cmd = nil
	if task.FileType == "jar" {
		gatlingSh := fmt.Sprintf("%s/bin/gatling-jar-runner.sh", g.BaseDir)
		cmd = exec.Command(gatlingSh, "-s", task.Simulation,
			"--results-folder", userFilesDir.Results)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_OPTS=%s", task.JavaOpts))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER_JAR_FILE=%s/*", userFilesDir.Simulations))
	} else {
		gatlingSh := fmt.Sprintf("%s/bin/gatling-runner.sh", g.BaseDir)
		cmd = exec.Command(gatlingSh,
			"-s", task.Simulation,
			"--simulations-folder", userFilesDir.Simulations,
			"--resources-folder", userFilesDir.Resources,
			"--results-folder", userFilesDir.Results,
			"--binaries-folder", userFilesDir.Binaries)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_OPTS=%s", task.JavaOpts))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER_LIB_DIR=%s/*", userFilesDir.Libraries))
	}
	log.Printf("About to execute command [%v]\n", cmd)
	err := commandOps.ExecuteAndLog(cmd, filepath.Join(task.UserFilesDir.BaseDir, "console.log"))
	return cmd, err
}
