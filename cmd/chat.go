package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmy7533018/mugen-ai/internal/config"
	ctxinfo "github.com/tmy7533018/mugen-ai/internal/context"
	"github.com/tmy7533018/mugen-ai/internal/history"
	"github.com/tmy7533018/mugen-ai/internal/ollama"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Interactive chat in the terminal",
	RunE:  runChat,
}

var (
	chatModel  string
	chatOllama string
	chatSystem string
)

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().StringVarP(&chatModel, "model", "m", "", "Ollama model to use (overrides config)")
	chatCmd.Flags().StringVar(&chatOllama, "ollama-host", "http://localhost:11434", "Ollama host URL")
	chatCmd.Flags().StringVar(&chatSystem, "system", "", "System prompt (overrides config)")
}

func runChat(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: config load failed, using defaults: %v\n", err)
		cfg = config.Default()
	}

	model := chatModel
	if model == "" {
		model = "gemma3:4b"
	}
	system := chatSystem
	if system == "" {
		system = cfg.Personality.SystemPrompt
	}

	client := ollama.New(chatOllama, model)
	hist := history.New(system)
	hist.ContextFunc = func() string { return ctxinfo.Build(cfg.Context) }
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Chat with %s  (commands: exit, clear)\n\n", chatModel)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		switch input {
		case "exit":
			return nil
		case "clear":
			hist.Clear()
			fmt.Println("History cleared.")
			continue
		}

		hist.Add("user", input)

		var fullResponse string
		err := client.Chat(context.Background(), hist.Messages(), func(chunk ollama.ChatChunk) error {
			fmt.Print(chunk.Message.Content)
			fullResponse += chunk.Message.Content
			return nil
		})
		fmt.Println()

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			hist.Clear()
			continue
		}

		hist.Add("assistant", fullResponse)
	}

	return nil
}
