package main

import (
	"strings"
	"testing"
)

func TestValidateSemverTagAcceptsSemanticVersions(t *testing.T) {
	tags := []string{
		"v0.1.0",
		"v1.2.3",
		"v10.20.30",
		"v1.2.3-rc.1",
		"v1.2.3+build.5",
		"v1.0.0-alpha.1+001",
	}
	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			if err := validateSemverTag(tag); err != nil {
				t.Fatalf("validateSemverTag(%q) error = %v, want nil", tag, err)
			}
		})
	}
}

func TestValidateSemverTagRejectsMalformedVersions(t *testing.T) {
	tags := []struct {
		name string
		tag  string
	}{
		{name: "empty", tag: ""},
		{name: "missing v prefix", tag: "1.2.3"},
		{name: "missing patch", tag: "v1.2"},
		{name: "extra segment", tag: "v1.2.3.4"},
		{name: "non numeric", tag: "va.b.c"},
		{name: "leading zero", tag: "v01.2.3"},
		{name: "dangling prerelease", tag: "v1.2.3-"},
		{name: "branch name", tag: "main"},
	}
	for _, tt := range tags {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSemverTag(tt.tag); err == nil {
				t.Fatalf("validateSemverTag(%q) error = nil, want a rejection", tt.tag)
			}
		})
	}
}

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		want    string
		wantErr bool
	}{
		{
			name: "defaults to a dev snapshot",
			want: devVersion,
		},
		{
			name: "takes the explicit flag",
			args: []string{"--version", "v1.2.3"},
			want: "v1.2.3",
		},
		{
			name: "takes the tag pushed in CI",
			env:  map[string]string{"GITHUB_REF_NAME": "v2.0.0"},
			want: "v2.0.0",
		},
		{
			name: "prefers the flag over the CI tag",
			args: []string{"--version", "v1.2.3"},
			env:  map[string]string{"GITHUB_REF_NAME": "v2.0.0"},
			want: "v1.2.3",
		},
		{
			name: "keeps a dev snapshot on branch builds",
			env:  map[string]string{"GITHUB_REF_NAME": "main"},
			want: devVersion,
		},
		{
			name:    "rejects a malformed explicit version",
			args:    []string{"--version", "1.2.3"},
			wantErr: true,
		},
		{
			name:    "rejects a malformed release tag",
			env:     map[string]string{"GITHUB_REF_NAME": "v1.2"},
			wantErr: true,
		},
		{
			name:    "rejects a flag without a value",
			args:    []string{"--version"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveVersion(tt.args, func(key string) string { return tt.env[key] })
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveVersion(%#v) error = nil, want a rejection", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveVersion(%#v) error = %v", tt.args, err)
			}
			if got != tt.want {
				t.Fatalf("resolveVersion(%#v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestGoBuildArgsStampVersion(t *testing.T) {
	args := goBuildArgs("dist/tyrano-parser", "v1.2.3")
	joined := strings.Join(args, " ")
	for _, want := range []string{"-X main.version=v1.2.3", "-trimpath", "./cmd/tyrano-parser"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("goBuildArgs() = %#v, want it to carry %q", args, want)
		}
	}
}
