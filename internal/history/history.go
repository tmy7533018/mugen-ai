package history

import (
	"sync"

	"github.com/tmy7533018/mugen-ai/internal/ollama"
)

type History struct {
	mu       sync.Mutex
	messages []ollama.Message
	system   string
}

func New(systemPrompt string) *History {
	return &History{system: systemPrompt}
}

func (h *History) Add(role, content string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = append(h.messages, ollama.Message{Role: role, Content: content})
}

// Messages returns all messages prefixed with the system prompt.
func (h *History) Messages() []ollama.Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	result := make([]ollama.Message, 0, len(h.messages)+1)
	if h.system != "" {
		result = append(result, ollama.Message{Role: "system", Content: h.system})
	}
	return append(result, h.messages...)
}

func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = nil
}
