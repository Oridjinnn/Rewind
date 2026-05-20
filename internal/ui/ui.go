package ui

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"

	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

func Header(title string) string {
	return Cyan +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n" +
		title + "\n" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━" +
		Reset
}

func Success(text string) string {
	return Green + text + Reset
}

func Error(text string) string {
	return Red + text + Reset
}

func Muted(text string) string {
	return Gray + text + Reset
}

func Highlight(text string) string {
	return Yellow + text + Reset
}
