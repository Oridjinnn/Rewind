package recall

import (
	"fmt"
	"strings"

	"github.com/Oridjinnn/Rewind/internal/storage"
	"github.com/Oridjinnn/Rewind/pkg/types"
)

func Recall(
	query string,
) {

	sessions, err := storage.LoadAllSessions()

	if err != nil {
		fmt.Println("load failed:", err)
		return
	}

	query = strings.ToLower(
		query,
	)

	fmt.Println("")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("REWIND MEMORY RECALL")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("QUERY:", query)
	fmt.Println("")

	found := false

	for _, session := range sessions {

		if MatchSession(
			session,
			query,
		) {

			found = true

			PrintMatch(
				session,
			)
		}
	}

	if !found {

		fmt.Println(
			"No matching memories found.",
		)
	}
}

func MatchSession(
	session types.Session,
	query string,
) bool {

	query = strings.ToLower(
		query,
	)

	if strings.Contains(
		strings.ToLower(session.ID),
		query,
	) {
		return true
	}

	if strings.Contains(
		strings.ToLower(session.Model),
		query,
	) {
		return true
	}

	if strings.Contains(
		strings.ToLower(session.Title),
		query,
	) {
		return true
	}

	if strings.Contains(
		strings.ToLower(session.Summary),
		query,
	) {
		return true
	}

	if strings.Contains(
		strings.ToLower(session.Mood),
		query,
	) {
		return true
	}

	for _, tag := range session.Tags {

		if strings.Contains(
			strings.ToLower(tag),
			query,
		) {
			return true
		}
	}

	for _, event := range session.Events {

		if strings.Contains(
			strings.ToLower(event.Type),
			query,
		) {
			return true
		}

		if strings.Contains(
			strings.ToLower(event.Content),
			query,
		) {
			return true
		}
	}

	return false

}

func PrintMatch(
	session types.Session,
) {

	fmt.Println("SESSION:", session.ID)

	if session.Title != "" {
		fmt.Println("TITLE:", session.Title)
	}

	if session.Summary != "" {
		fmt.Println("SUMMARY:", session.Summary)
	}

	fmt.Println("MODEL:", session.Model)

	fmt.Println("")
}
