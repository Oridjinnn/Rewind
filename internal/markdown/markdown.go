package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

func ExportMarkdown(
	session types.Session,
) error {

	err := os.MkdirAll(
		"exports",
		os.ModePerm,
	)

	if err != nil {
		return err
	}

	filePath := filepath.Join(
		"exports",
		session.ID+".md",
	)

	file, err := os.Create(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	builder := strings.Builder{}

	builder.WriteString(
		"# REWIND SESSION\n\n",
	)

	if session.Title != "" {

		builder.WriteString(
			fmt.Sprintf(
				"**Title:** %s\n\n",
				session.Title,
			),
		)
	}

	if session.Model != "" {

		builder.WriteString(
			fmt.Sprintf(
				"**Model:** %s\n\n",
				session.Model,
			),
		)
	}

	if session.Mood != "" {

		builder.WriteString(
			fmt.Sprintf(
				"**Mood:** %s\n\n",
				session.Mood,
			),
		)
	}

	if len(session.Tags) > 0 {

		builder.WriteString(
			"**Tags:** ",
		)

		for i, tag := range session.Tags {

			builder.WriteString(tag)

			if i < len(session.Tags)-1 {
				builder.WriteString(", ")
			}
		}

		builder.WriteString("\n\n")
	}

	builder.WriteString(
		fmt.Sprintf(
			"**Started:** %s\n\n",
			session.StartedAt,
		),
	)

	if session.Summary != "" {

		builder.WriteString(
			"## SUMMARY\n\n",
		)

		builder.WriteString(
			session.Summary + "\n\n",
		)
	}

	builder.WriteString("---\n\n")

	startTime, _ := time.Parse(
		time.RFC3339,
		session.StartedAt,
	)

	for _, event := range session.Events {

		eventTime, _ := time.Parse(
			time.RFC3339,
			event.Timestamp,
		)

		offset := eventTime.Sub(startTime)

		minutes := int(offset.Minutes())
		seconds := int(offset.Seconds()) % 60

		timestamp := fmt.Sprintf(
			"%02d:%02d",
			minutes,
			seconds,
		)

		switch event.Type {

		case "user_message":

			builder.WriteString(
				fmt.Sprintf(
					"## %s YOU\n\n%s\n\n",
					timestamp,
					event.Content,
				),
			)

		case "assistant_message":

			builder.WriteString(
				fmt.Sprintf(
					"## %s ASSISTANT\n\n%s\n\n",
					timestamp,
					event.Content,
				),
			)

		default:

			builder.WriteString(
				fmt.Sprintf(
					"## %s %s\n\n%s\n\n",
					timestamp,
					event.Type,
					event.Content,
				),
			)
		}
	}

	_, err = file.WriteString(
		builder.String(),
	)

	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("markdown exported:")
	fmt.Println(filePath)
	fmt.Println("")

	return nil
}
