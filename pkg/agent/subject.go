package agent

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	invalidServiceNameCharactersRegex = regexp.MustCompile("[:.]")

	subjectPrefix = "dbus"
)

func NormalizeObjectPath(path string) string {
	// ensure there is a leading /
	if path[0] != '/' {
		path = "/" + path
	}
	return strings.ReplaceAll(path, "/", ".")
}

func NormalizeDestination(name string) string {
	return invalidServiceNameCharactersRegex.ReplaceAllString(name, "_")
}

func busSubjects(dest string) []string {
	return []string{
		// this subject allows for targeted requests against a specific agent
		fmt.Sprintf("%s.agent.%s.%s", subjectPrefix, nkey, NormalizeDestination(dest)),
		// this subject allows for broadcast requests against all agents
		fmt.Sprintf("%s.broadcast.%s", subjectPrefix, NormalizeDestination(dest)),
	}
}
