package main

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestCICommandsMatchVerificationSuite(t *testing.T) {
	got := ciCommands()
	want := []commandSpec{
		{name: "unit tests", exe: "go", args: []string{"test", "./..."}},
		{name: "race tests", exe: "go", args: []string{"test", "-race", "./..."}},
		{name: "coverage tests", exe: "go", args: []string{"test", "-cover", "./..."}},
		{name: "vet", exe: "go", args: []string{"vet", "./..."}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ciCommands() = %#v, want %#v", got, want)
	}
}

func TestRunCommandsStopsOnFirstFailure(t *testing.T) {
	commands := []commandSpec{
		{name: "first", exe: "go", args: []string{"test", "./..."}},
		{name: "second", exe: "go", args: []string{"vet", "./..."}},
	}
	var ran []string
	fail := errors.New("boom")
	err := runCommands(commands, bytes.NewBuffer(nil), func(cmd commandSpec) error {
		ran = append(ran, cmd.name)
		if cmd.name == "first" {
			return fail
		}
		return nil
	})
	if !errors.Is(err, fail) {
		t.Fatalf("runCommands() error = %v, want %v", err, fail)
	}
	if !reflect.DeepEqual(ran, []string{"first"}) {
		t.Fatalf("ran = %#v, want first command only", ran)
	}
}

func TestRunCommandsPrintsStepHeader(t *testing.T) {
	commands := []commandSpec{
		{name: "unit tests", exe: "go", args: []string{"test", "./..."}},
	}
	var out bytes.Buffer
	err := runCommands(commands, &out, func(commandSpec) error { return nil })
	if err != nil {
		t.Fatalf("runCommands() error = %v", err)
	}
	want := "==> unit tests: go test ./...\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}
