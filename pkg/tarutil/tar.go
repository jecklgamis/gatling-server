package tarutil

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"log"
	"os/exec"
)

//TODO Use Golang way of archiving?

func Extract(tgz string, destDir string) error {
	log.Printf("Extracting %s to %s", tgz, destDir)
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("tar xvzf %s -C %s --strip 1", tgz, destDir))
	err := cmdexec.NewCommandExecutor().Execute(cmd)
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func CompressDir(sourceDir string, destDir, tgz string) error {
	log.Printf("Archiving %s to %s/%s", sourceDir, destDir, tgz)
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar cvzf %s/%s .",
		sourceDir, destDir, tgz))
	err := cmdexec.NewCommandExecutor().Execute(cmd)
	if err != nil {
		return err
	}
	return cmd.Wait()
}
