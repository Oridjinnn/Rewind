package memory

import (
	"strings"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func BuildMemoryContext(
	sessions []types.Session,
	query string,
) string {

	query = strings.ToLower(query)

	var memories []string

	for _, session := range sessions {

		score := 0

		// title match
		if strings.Contains(
			strings.ToLower(session.Title),
			query,
		) {
			score += 3
		}

		// summary match
		if strings.Contains(
			strings.ToLower(session.Summary),
			query,
		) {
			score += 2
		}

		// tag match
		for _, tag := range session.Tags {

			if strings.Contains(
				strings.ToLower(tag),
				query,
			) {
				score += 2
			}
		}

		// event match
		for _, event := range session.Events {

			if strings.Contains(
				strings.ToLower(event.Content),
				query,
			) {
				score += 1
				break
			}
		}

		if score > 0 {

			memoryText :=
				"Title: " + session.Title + "\n" +
					"Summary: " + session.Summary + "\n"

			memories = append(
				memories,
				memoryText,
			)
		}

		// HARD LIMIT
		if len(memories) >= 3 {
			break
		}
	}

	return strings.Join(
		memories,
		"\n---\n",
	)
}
