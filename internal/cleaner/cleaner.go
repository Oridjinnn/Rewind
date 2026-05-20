package cleaner

import (
	"regexp"
	"strings"
)

var ansiRegex = regexp.MustCompile(
	`\x1b\[[0-9;?]*[a-zA-Z]`,
)

var spinnerRegex = regexp.MustCompile(
	`[⠁-⣿]+`,
)

func CleanANSI(
	text string,
) string {

	cleaned := ansiRegex.ReplaceAllString(
		text,
		"",
	)

	cleaned = spinnerRegex.ReplaceAllString(
		cleaned,
		"",
	)

	cleaned = strings.ReplaceAll(
		cleaned,
		"\r",
		"",
	)

	cleaned = strings.TrimSpace(
		cleaned,
	)

	return cleaned
}
