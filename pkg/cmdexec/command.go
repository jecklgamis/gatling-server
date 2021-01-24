package cmdexec

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
)

type CommandExecutor struct {
}

func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// these commands do not call cmd.Wait(), the client is expected to call it to get the result
type CommandExecutionOps interface {
	Execute(cmd *exec.Cmd) error
	ExecuteAndLog(cmd *exec.Cmd, filename string) error
}

func (c *CommandExecutor) ExecuteAndLog(cmd *exec.Cmd, filename string) error {
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go readPipeToFile(pipe, filename)
	return nil
}

func (c *CommandExecutor) Execute(cmd *exec.Cmd) error {
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go readPipe(pipe, os.Stdout)
	return nil
}

func readPipe(readCloser io.ReadCloser, writer io.Writer) {
	reader := bufio.NewReader(readCloser)
	line, err := reader.ReadString('\n')
	for err == nil {
		writer.Write([]byte(line))
		line, err = reader.ReadString('\n')
	}
}

func readPipeToFile(pipe io.ReadCloser, file string) error {
	f, err := os.Create(file)
	if err != nil {
		log.Println("Unable to open file for writing :", err)
		return err
	}
	readPipe(pipe, f)
	defer f.Close()
	return nil
}
