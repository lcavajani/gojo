package execute

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
)

type ExecTask struct {
	// Log is a logger
	Log zerolog.Logger

	Command string
	Args    []string
	// Shell   bool
	// Env    []string
	Cwd    string
	DryRun bool

	// Stdin connects a reader to stdin for the command
	// being executed.
	Stdin io.Reader

	// StreamStdio prints stdout and stderr directly to os.Stdout/err as
	// the command runs.
	StreamStdio bool

	// PrintCommand prints the command before executing
	PrintCommand bool
}

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (et *ExecTask) AddArgs(args ...string) {
	et.Args = append(et.Args, args...)
}

func (et *ExecTask) String() string {
	args := ""
	if len(et.Args) > 0 {
		args = strings.Join(et.Args, " ")
	}

	return fmt.Sprintf("%s %s", et.Command, args)
}

func (et *ExecTask) Print() {
	et.Log.Info().Str("cmd", et.String()).
		Msg("execute command")
}

func (et *ExecTask) Execute() (ExecResult, error) {
	et.Print()
	if et.DryRun {
		return ExecResult{}, nil
	}

	var cmd *exec.Cmd
	if strings.Index(et.Command, " ") > 0 {
		parts := strings.Split(et.Command, " ")
		command := parts[0]
		args := parts[1:]
		cmd = exec.Command(command, args...)
	} else {
		cmd = exec.Command(et.Command, et.Args...)
	}

	cmd.Dir = et.Cwd

	if et.Stdin != nil {
		cmd.Stdin = et.Stdin
	}

	stdoutBuff := bytes.Buffer{}
	stderrBuff := bytes.Buffer{}

	var stdoutWriters io.Writer
	var stderrWriters io.Writer

	if et.StreamStdio {
		stdoutWriters = io.MultiWriter(os.Stdout, &stdoutBuff)
		stderrWriters = io.MultiWriter(os.Stderr, &stderrBuff)
	} else {
		stdoutWriters = &stdoutBuff
		stderrWriters = &stderrBuff
	}

	cmd.Stdout = stdoutWriters
	cmd.Stderr = stderrWriters

	if err := cmd.Start(); err != nil {
		return ExecResult{}, err
	}

	exitCode := 0
	err := cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	execResult := ExecResult{
		Stdout:   stdoutBuff.String(),
		Stderr:   stderrBuff.String(),
		ExitCode: exitCode,
	}

	if !et.StreamStdio {
		et.Log.Info().Int("exit-code", execResult.ExitCode).
			Str("stdout", execResult.Stdout).
			Str("stderr", execResult.Stderr).
			Msg("cmd result")
	}

	return execResult, err
}
