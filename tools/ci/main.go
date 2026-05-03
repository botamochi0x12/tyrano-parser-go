package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type commandSpec struct {
	name string
	exe  string
	args []string
}

type commandRunner func(commandSpec) error

func ciCommands() []commandSpec {
	return []commandSpec{
		{name: "unit tests", exe: "go", args: []string{"test", "./..."}},
		{name: "race tests", exe: "go", args: []string{"test", "-race", "./..."}},
		{name: "coverage tests", exe: "go", args: []string{"test", "-cover", "./..."}},
		{name: "vet", exe: "go", args: []string{"vet", "./..."}},
	}
}

func runCommands(commands []commandSpec, out io.Writer, runner commandRunner) error {
	for _, cmd := range commands {
		fmt.Fprintf(out, "==> %s: %s %s\n", cmd.name, cmd.exe, strings.Join(cmd.args, " "))
		if err := runner(cmd); err != nil {
			return err
		}
	}
	return nil
}

func execCommand(cmd commandSpec) error {
	command := exec.Command(cmd.exe, cmd.args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func main() {
	if err := runCommands(ciCommands(), os.Stdout, execCommand); err != nil {
		fmt.Fprintf(os.Stderr, "ci failed: %v\n", err)
		os.Exit(1)
	}
}
