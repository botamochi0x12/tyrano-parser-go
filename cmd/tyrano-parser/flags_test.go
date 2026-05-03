package main

import "testing"

func TestParseGlobalFlags_Defaults(t *testing.T) {
	g, rest, err := parseGlobalFlags([]string{"file.ks"})
	if err != nil {
		t.Fatal(err)
	}
	if g.format != "json" || g.strict || g.quiet || g.projectRoot != "" {
		t.Errorf("unexpected defaults: %+v", g)
	}
	if len(rest) != 1 || rest[0] != "file.ks" {
		t.Errorf("rest = %v", rest)
	}
}

func TestParseGlobalFlags_AllSet(t *testing.T) {
	g, rest, err := parseGlobalFlags([]string{"--project-root", "/tmp/proj", "--format", "report", "--strict", "--quiet", "file.ks"})
	if err != nil {
		t.Fatal(err)
	}
	if g.projectRoot != "/tmp/proj" || g.format != "report" || !g.strict || !g.quiet {
		t.Errorf("flags not set: %+v", g)
	}
	if len(rest) != 1 || rest[0] != "file.ks" {
		t.Errorf("rest = %v", rest)
	}
}

func TestParseGlobalFlags_BadFormat(t *testing.T) {
	_, _, err := parseGlobalFlags([]string{"--format", "xml"})
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseGlobalFlags_FlagsAfterPositional(t *testing.T) {
	g, rest, err := parseGlobalFlags([]string{"file.ks", "--format", "report", "--strict"})
	if err != nil {
		t.Fatal(err)
	}
	if g.format != "report" || !g.strict {
		t.Errorf("flags not parsed after positional: %+v", g)
	}
	if len(rest) != 1 || rest[0] != "file.ks" {
		t.Errorf("rest = %v", rest)
	}
}
