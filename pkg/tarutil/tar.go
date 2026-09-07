package tarutil

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"log/slog"
	"os/exec"
)

//TODO Use Golang way of archiving?

func Extract(tgz string, destDir string) error {
	slog.Info("Extracting", "tgz", tgz, "destDir", destDir)
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("tar xvzf %s -C %s --strip 1", tgz, destDir))
	err := cmdexec.NewCommandExecutor().Execute(cmd)
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func CompressDir(sourceDir string, destDir, tgz string) error {
	slog.Info("Archiving", "sourceDir", sourceDir, "destDir", destDir, "tgz", tgz)
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar cvzf %s/%s .",
		sourceDir, destDir, tgz))
	err := cmdexec.NewCommandExecutor().Execute(cmd)
	if err != nil {
		return err
	}
	return cmd.Wait()
}
