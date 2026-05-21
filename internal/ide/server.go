package ide

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server runs an HTTP server that receives IDE events.
type Server struct {
	recorder *SQLiteRecorder
	port     int
	server   *http.Server
	started  time.Time
}

// NewServer creates a new IDE event HTTP server.
func NewServer(recorder *SQLiteRecorder, port int) *Server {
	if port == 0 {
		port = ServerPort
	}
	return &Server{recorder: recorder, port: port}
}

// Start begins listening for IDE events.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleEvent)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/batch", s.handleBatch)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	s.recorder.SetServerState(true, s.port)
	s.started = time.Now()

	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on :%d: %w", s.port, err)
	}

	fmt.Printf(" Rewind IDE server listening on http://localhost:%d\n", s.port)
	fmt.Println("   Waiting for IDE extensions to connect...")
	fmt.Println("")
	fmt.Println("   To enable recording:  rewind ide permissions <ide-name> on")
	fmt.Println("   To check status:      rewind ide status")
	fmt.Println("")

	// Graceful shutdown on SIGINT/SIGTERM
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\n Shutting down IDE server...")
		s.Stop()
	}()

	return s.server.Serve(ln)
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() error {
	s.recorder.SetServerState(false, s.port)
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Port returns the server's port.
func (s *Server) Port() int {
	return s.port
}

// StartedAt returns when the server was started.
func (s *Server) StartedAt() time.Time {
	return s.started
}

// handleEvent processes single IDE events (POST /).
func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	event, err := ParseEvent(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid event: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.recorder.RecordEvent(event); err != nil {
		if err.Error()[:19] == "recording disabled" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, fmt.Sprintf("failed to record: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleBatch processes multiple IDE events (POST /batch).
func (s *Server) handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var events []json.RawMessage
	if err := json.Unmarshal(body, &events); err != nil {
		http.Error(w, "invalid batch JSON", http.StatusBadRequest)
		return
	}

	accepted := 0
	rejected := 0
	var errors []string

	for _, raw := range events {
		event, err := ParseEvent(raw)
		if err != nil {
			rejected++
			errors = append(errors, err.Error())
			continue
		}

		if err := s.recorder.RecordEvent(event); err != nil {
			rejected++
			errors = append(errors, err.Error())
			continue
		}
		accepted++
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"accepted": accepted,
		"rejected": rejected,
		"errors":   errors,
	})
}

// handleHealth returns server health (GET /health).
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.started).String()
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"uptime":  uptime,
		"version": ProtocolVersion,
	})
}

// handleStatus returns current IDE recording status (GET /status).
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.recorder.GetStatus()
	if err != nil {
		log.Printf("status error: %v", err)
		http.Error(w, "failed to get status", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(status)
}