package repo

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	sshPattern   = regexp.MustCompile(`git@github\.com:([^/]+)/([^/.]+?)(?:\.git)?$`)
	httpsPattern = regexp.MustCompile(`https?://github\.com/([^/]+)/([^/.]+?)(?:\.git)?$`)
)

func Detect() (owner, name string, err error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", "", fmt.Errorf("not a git repository or no origin remote: %w", err)
	}

	url := strings.TrimSpace(string(out))

	if m := sshPattern.FindStringSubmatch(url); m != nil {
		return m[1], m[2], nil
	}
	if m := httpsPattern.FindStringSubmatch(url); m != nil {
		return m[1], m[2], nil
	}

	return "", "", fmt.Errorf("could not parse GitHub remote URL: %s", url)
}
