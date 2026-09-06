package cmdexec

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

// maxLogFileSize bounds how much output ExecuteAndLog will write to the log
// file for a single command, preventing a runaway or malicious process from
// filling up disk with unbounded log output.
const maxLogFileSize = 20 << 20 // 20MB

type CommandExecutor struct {
}

func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// this command does not call cmd.Wait(), the client is expected to call it to get the result
type CommandExecutionOps interface {
	ExecuteAndLog(cmd *exec.Cmd, filename string, taskId string) error
}

func (c *CommandExecutor) ExecuteAndLog(cmd *exec.Cmd, filename string, taskId string) error {
	f, err := os.Create(filename)
	if err != nil {
		log.Println("Unable to open file for writing :", err)
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		f.Close()
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		f.Close()
		return err
	}
	if err := cmd.Start(); err != nil {
		f.Close()
		return err
	}
	fileWriter := newPrefixWriter(newLimitedWriter(f, maxLogFileSize), taskId)
	stdoutWriter := newPrefixWriter(os.Stdout, taskId)
	writer := io.MultiWriter(fileWriter, stdoutWriter)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); readPipe(stdout, writer) }()
	go func() { defer wg.Done(); readPipe(stderr, writer) }()
	go func() { wg.Wait(); f.Close() }()
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
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			writer.Write([]byte(line))
		}
		if err != nil {
			return
		}
	}
}

// limitedWriter forwards writes to an underlying writer up to a fixed byte
// budget, discarding anything beyond it. It is safe for concurrent use since
// stdout and stderr are read into the same destination on separate goroutines.
type limitedWriter struct {
	mu        sync.Mutex
	w         io.Writer
	remaining int64
}

func newLimitedWriter(w io.Writer, limit int64) *limitedWriter {
	return &limitedWriter{w: w, remaining: limit}
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.remaining <= 0 {
		return len(p), nil
	}
	toWrite := p
	if int64(len(toWrite)) > l.remaining {
		toWrite = toWrite[:l.remaining]
	}
	n, err := l.w.Write(toWrite)
	l.remaining -= int64(n)
	return len(p), err
}

// prefixWriter writes each incoming chunk to the underlying writer with a
// fixed prefix prepended. It is safe for concurrent use since stdout and
// stderr are read into the same destination on separate goroutines.
type prefixWriter struct {
	mu     sync.Mutex
	w      io.Writer
	prefix string
}

func newPrefixWriter(w io.Writer, taskId string) *prefixWriter {
	return &prefixWriter{w: w, prefix: fmt.Sprintf("[%s] ", taskId)}
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := io.WriteString(p.w, p.prefix); err != nil {
		return 0, err
	}
	if _, err := p.w.Write(b); err != nil {
		return 0, err
	}
	return len(b), nil
}
