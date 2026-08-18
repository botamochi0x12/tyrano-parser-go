package main

import (
	"errors"
	"fmt"
	"regexp"
)

// devVersion labels artifacts built outside a tagged release.
const devVersion = "dev"

// semverTagPattern is the pattern recommended by semver.org, extended with the
// leading "v" that Git release tags carry.
var semverTagPattern = regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
	`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// releaseTagPattern tells a release tag apart from an ordinary ref such as
// "main" or a "version-bump" branch, so only real tags are held to semver.
var releaseTagPattern = regexp.MustCompile(`^v\d`)

func validateSemverTag(tag string) error {
	if !semverTagPattern.MatchString(tag) {
		return fmt.Errorf("%q is not a semantic version tag (want vMAJOR.MINOR.PATCH)", tag)
	}
	return nil
}

// resolveVersion picks the version to stamp into the binaries: an explicit
// --version flag first, then the tag CI is building, and a dev snapshot
// otherwise. Anything that looks like a release tag must be valid semver.
func resolveVersion(args []string, lookupEnv func(string) string) (string, error) {
	for i := 0; i < len(args); i++ {
		if args[i] != "--version" && args[i] != "-version" {
			continue
		}
		if i+1 >= len(args) {
			return "", errors.New("--version requires a value, e.g. --version v1.2.3")
		}
		return checkedVersion(args[i+1])
	}
	ref := lookupEnv("GITHUB_REF_NAME")
	if !releaseTagPattern.MatchString(ref) {
		return devVersion, nil
	}
	return checkedVersion(ref)
}

func checkedVersion(value string) (string, error) {
	if value == devVersion {
		return devVersion, nil
	}
	if err := validateSemverTag(value); err != nil {
		return "", err
	}
	return value, nil
}

func goBuildArgs(output, version string) []string {
	return []string{
		"build",
		"-trimpath",
		"-ldflags", "-s -w -X main.version=" + version,
		"-o", output,
		"./cmd/tyrano-parser",
	}
}
