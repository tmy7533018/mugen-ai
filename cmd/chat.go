package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
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
	chatCmd.Flags().StringVarP(&chatModel, "model", "m", "gemma3:4b", "Ollama model to use")
	chatCmd.Flags().StringVar(&chatOllama, "ollama-host", "http://localhost:11434", "Ollama host URL")
	chatCmd.Flags().StringVar(&chatSystem, "system", "You are a helpful desktop assistant. Be concise.", "System prompt")
}

func runChat(_ *cobra.Command, _ []string) error {
	client := ollama.New(chatOllama, chatModel)
	hist := history.New(chatSystem)
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
