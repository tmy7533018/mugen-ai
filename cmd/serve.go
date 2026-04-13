package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
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
	serveCmd.Flags().StringVarP(&serveModel, "model", "m", "gemma3:4b", "Ollama model to use")
	serveCmd.Flags().StringVar(&serveOllama, "ollama-host", "http://localhost:11434", "Ollama host URL")
	serveCmd.Flags().StringVar(&serveSystem, "system", "You are a helpful desktop assistant. Be concise.", "System prompt")
}

func runServe(_ *cobra.Command, _ []string) error {
	client := ollama.New(serveOllama, serveModel)
	hist := history.New(serveSystem)
	srv := server.New(client, hist, serveModel)

	addr := fmt.Sprintf("127.0.0.1:%d", servePort)
	fmt.Fprintf(os.Stdout, "mugen-ai listening on %s (model: %s)\n", addr, serveModel)
	return http.ListenAndServe(addr, srv.Routes())
}
