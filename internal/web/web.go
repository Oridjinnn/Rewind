package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Oridjinnn/Rewind/internal/recall"
	"github.com/Oridjinnn/Rewind/internal/storage"
	"github.com/Oridjinnn/Rewind/pkg/types"
)

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Rewind Web UI</title>
  <style>
    body { font-family: Inter, -apple-system, BlinkMacSystemFont, sans-serif; margin: 0; padding: 0; background: #f5f7fb; color: #111827; }
    header { background: #0f172a; color: white; padding: 24px; }
    a { color: #2563eb; text-decoration: none; }
    a:hover { text-decoration: underline; }
    main { padding: 24px; max-width: 1100px; margin: auto; }
    .hero { display: flex; flex-wrap: wrap; gap: 16px; align-items: center; }
    .hero h1 { margin: 0; font-size: 2rem; }
    .pill { display: inline-block; background: #e0e7ff; color: #3730a3; padding: 4px 10px; border-radius: 999px; margin-top: 8px; }
    form input { width: 100%; max-width: 420px; padding: 10px 12px; border-radius: 10px; border: 1px solid #cbd5e1; }
    .card { background: white; border: 1px solid #e2e8f0; border-radius: 16px; padding: 18px; margin-bottom: 16px; box-shadow: 0 4px 18px rgba(15,23,42,0.06); }
    .grid { display: grid; gap: 16px; }
    .list-item { padding: 16px 0; border-bottom: 1px solid #e2e8f0; }
    .list-item:last-child { border-bottom: none; }
    .label { color: #64748b; font-size: 0.9rem; }
    pre { background: #0f172a; color: #f8fafc; padding: 14px; border-radius: 12px; overflow-x: auto; }
    .meta { display: grid; grid-template-columns: auto auto; gap: 12px; margin-top: 12px; }
    .small { font-size: 0.95rem; color: #475569; }
    .footer { margin-top: 40px; font-size: 0.95rem; color: #64748b; }
  </style>
</head>
<body>
<header>
  <div class="hero">
    <div>
      <h1>Rewind Web UI</h1>
      <p class="small">Browse and search your recorded terminal sessions in the browser.</p>
    </div>
    <span class="pill">{{if .Query}}Search: {{.Query}}{{else}}All sessions{{end}}</span>
  </div>
</header>
<main>
  <form action="/" method="get">
    <input type="search" name="q" placeholder="Search sessions, commands, outputs..." value="{{.Query}}">
  </form>
  {{if .Error}}
  <div class="card" style="border-color:#fca5a5;background:#fef2f2;color:#991b1b;">{{.Error}}</div>
  {{end}}
  {{if .SelectedSession}}
    <div class="card">
      <h2>Session {{.SelectedSession.ID}}</h2>
      <div class="meta">
        <div><strong>Command</strong><div class="small">{{.SelectedSession.Command}}</div></div>
        <div><strong>Model</strong><div class="small">{{.SelectedSession.Model}}</div></div>
        <div><strong>Summary</strong><div class="small">{{if .SelectedSession.Summary}}{{.SelectedSession.Summary}}{{else}}—{{end}}</div></div>
        <div><strong>Started</strong><div class="small">{{.SelectedSession.StartedAt}}</div></div>
        <div><strong>Exit</strong><div class="small">{{.SelectedSession.ExitCode}}</div></div>
      </div>
      <h3>Events</h3>
      {{range .SelectedSession.Events}}
      <div class="card" style="background:#f8fafc;">
        <div class="small"><strong>{{.Type}}</strong> • {{.Timestamp}}</div>
        <pre>{{.Content}}</pre>
      </div>
      {{else}}
      <div class="small">No events recorded.</div>
      {{end}}
      <div class="footer"><a href="/">Back to session list</a></div>
    </div>
  {{else}}
    <div class="grid">
      {{range .Sessions}}
      <div class="card">
        <div class="list-item">
          <h2><a href="/session?id={{.ID}}">{{.Command}}</a></h2>
          <div class="small">id: {{.ID}}</div>
          {{if .Summary}}<p>{{.Summary}}</p>{{end}}
          <div class="meta">
            <div><span class="label">Started</span><br>{{.StartedAt}}</div>
            <div><span class="label">Exit</span><br>{{.ExitCode}}</div>
            <div><span class="label">Events</span><br>{{len .Events}}</div>
          </div>
        </div>
      </div>
      {{else}}
      <div class="card">No sessions found.</div>
      {{end}}
    </div>
  {{end}}
</main>
</body>
</html>`))

type pageData struct {
	Sessions        []types.Session
	SelectedSession *types.Session
	Query           string
	Error           string
}

func Serve(port int) error {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/session", sessionHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	sessions, err := storage.LoadAllSessions()
	if err != nil {
		render(w, pageData{Error: fmt.Sprintf("failed to load sessions: %v", err)})
		return
	}

	if query != "" {
		sessions = filterSessions(sessions, query)
	}

	sortSessions(sessions)
	render(w, pageData{Sessions: sessions, Query: query})
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		render(w, pageData{Error: "missing session id"})
		return
	}

	sessionPath := filepath.Join("sessions", id+".json")
	session, err := storage.LoadSession(sessionPath)
	if err != nil {
		render(w, pageData{Error: fmt.Sprintf("failed to load session: %v", err)})
		return
	}

	render(w, pageData{SelectedSession: &session})
}

func filterSessions(sessions []types.Session, query string) []types.Session {
	var filtered []types.Session
	for _, session := range sessions {
		if recall.MatchSession(session, strings.ToLower(query)) {
			filtered = append(filtered, session)
		}
	}
	return filtered
}

func sortSessions(sessions []types.Session) {
	sort.SliceStable(sessions, func(i, j int) bool {
		// RFC3339 strings are lexicographically sortable. 
		// Comparing strings is significantly faster than parsing time in a loop.
		if sessions[i].StartedAt != sessions[j].StartedAt {
			return sessions[i].StartedAt > sessions[j].StartedAt
		}
		return sessions[i].ID > sessions[j].ID
	})
}

func render(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplate.Execute(w, data); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
	}
}
