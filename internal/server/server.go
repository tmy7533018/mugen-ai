package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/tmy7533018/mugen-ai/internal/history"
	"github.com/tmy7533018/mugen-ai/internal/ollama"
)

const maxRequestBody = 64 * 1024 // 64KB

type Server struct {
	mu      sync.RWMutex
	client  *ollama.Client
	history *history.History
	model   string
}

func New(client *ollama.Client, hist *history.History, model string) *Server {
	return &Server{client: client, history: hist, model: model}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat", s.handleChat)
	mux.HandleFunc("DELETE /history", s.handleClearHistory)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /models", s.handleModels)
	mux.HandleFunc("PUT /model", s.handleSwitchModel)
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type chatRequest struct {
	Message string `json:"message"`
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	s.history.Add("user", req.Message)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	var fullResponse string

	err := s.client.Chat(r.Context(), s.history.Messages(), func(chunk ollama.ChatChunk) error {
		fullResponse += chunk.Message.Content
		data, _ := json.Marshal(map[string]any{
			"content": chunk.Message.Content,
			"done":    chunk.Done,
		})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return nil
	})

	if err != nil {
		// Remove the dangling user message on failure
		s.history.RemoveLast()
		data, _ := json.Marshal(map[string]any{"error": err.Error(), "done": true})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return
	}

	if fullResponse != "" {
		s.history.Add("assistant", fullResponse)
	}
}

func (s *Server) handleClearHistory(w http.ResponseWriter, _ *http.Request) {
	s.history.Clear()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	model := s.model
	s.mu.RUnlock()

	ollamaOk := s.client.Ping(r.Context())

	status := "ok"
	code := http.StatusOK
	if !ollamaOk {
		status = "ollama_unavailable"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"model":  model,
	})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	models, err := s.client.Models(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"models": models})
}

type switchModelRequest struct {
	Model string `json:"model"`
}

func (s *Server) handleSwitchModel(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

	var req switchModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Model == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.model = req.Model
	s.mu.Unlock()

	s.client.SetModel(req.Model)
	s.history.Clear()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"model": req.Model})
}
