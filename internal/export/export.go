package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

func ExportHTML(session types.Session) (string, error) {

	err := os.MkdirAll("reports", 0755)

	if err != nil {
		return "", err
	}

	filename := session.ID + ".html"

	outputPath := filepath.Join(
		"reports",
		filename,
	)

	var builder strings.Builder

	builder.WriteString(`
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Rewind Report</title>

<style>

body {
	font-family: monospace;
	background: #0f1117;
	color: #e6edf3;
	padding: 40px;
	line-height: 1.6;
}

h1 {
	color: #58a6ff;
}

.section {
	margin-top: 30px;
	padding: 20px;
	border: 1px solid #30363d;
	border-radius: 8px;
	background: #161b22;
}

.event {
	padding: 10px;
	margin-bottom: 10px;
	border-left: 4px solid #58a6ff;
	background: #0d1117;
}

.stdout {
	border-color: #3fb950;
}

.stderr {
	border-color: #f85149;
}

.meta {
	color: #8b949e;
}

pre {
	white-space: pre-wrap;
	word-wrap: break-word;
}

</style>
</head>
<body>
`)

	builder.WriteString(
		fmt.Sprintf(
			"<h1>REWIND EXECUTION REPORT</h1>",
		),
	)

	builder.WriteString(`
<div class="section">
`)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Session ID:</strong> %s</p>",
			session.ID,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Command:</strong> %s</p>",
			session.Command,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Started:</strong> %s</p>",
			session.StartedAt,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Ended:</strong> %s</p>",
			session.EndedAt,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Exit Code:</strong> %d</p>",
			session.ExitCode,
		),
	)

	builder.WriteString(`
</div>
`)

	builder.WriteString(`
<div class="section">
<h2>EVENT TIMELINE</h2>
`)

	for _, event := range session.Events {

		className := "event"

		if event.Type == "stdout" {
			className += " stdout"
		}

		if event.Type == "stderr" {
			className += " stderr"
		}

		builder.WriteString(
			fmt.Sprintf(
				`
<div class="%s">
<div class="meta">%s — %s</div>
<pre>%s</pre>
</div>
`,
				className,
				event.Timestamp,
				event.Type,
				event.Content,
			),
		)
	}

	builder.WriteString(`
</div>
`)

	builder.WriteString(`
</body>
</html>
`)

	err = os.WriteFile(
		outputPath,
		[]byte(builder.String()),
		0644,
	)

	if err != nil {
		return "", err
	}

	return outputPath, nil
}
