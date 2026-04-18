package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmy7533018/mugen-ai/internal/config"
	ctxinfo "github.com/tmy7533018/mugen-ai/internal/context"
	"github.com/tmy7533018/mugen-ai/internal/history"
	"github.com/tmy7533018/mugen-ai/internal/ollama"
	"github.com/tmy7533018/mugen-ai/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the mugen-ai HTTP server for mugen-shell integration",
	RunE:  runServe,
}

var (
	servePort   int
	serveModel  string
	serveOllama string
	serveSystem string
)

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 11435, "port to listen on")
	serveCmd.Flags().StringVarP(&serveModel, "model", "m", "", "Ollama model to use (overrides config)")
	serveCmd.Flags().StringVar(&serveOllama, "ollama-host", "http://localhost:11434", "Ollama host URL")
	serveCmd.Flags().StringVar(&serveSystem, "system", "", "System prompt (overrides config)")
}

func runServe(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: config load failed, using defaults: %v\n", err)
		cfg = config.Default()
	}

	model := serveModel
	if model == "" {
		model = "gemma3:4b"
	}
	system := serveSystem
	if system == "" {
		system = cfg.Personality.SystemPrompt
	}

	client := ollama.New(serveOllama, model)
	hist := history.New(system)
	hist.ContextFunc = func() string { return ctxinfo.Build(cfg.Context) }
	srv := server.New(client, hist, model)

	addr := fmt.Sprintf("127.0.0.1:%d", servePort)
	httpSrv := &http.Server{Addr: addr, Handler: srv.Routes()}

	done := make(chan error, 1)
	go func() { done <- httpSrv.ListenAndServe() }()

	fmt.Fprintf(os.Stdout, "mugen-ai listening on %s (model: %s)\n", addr, serveModel)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		fmt.Fprintf(os.Stdout, "\nreceived %s, shutting down...\n", sig)
	case err := <-done:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return httpSrv.Shutdown(ctx)
}
