package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type streamEvent struct {
	Type    string       `json:"type"`
	Subtype string       `json:"subtype"`
	Message *eventMessage `json:"message"`
	Result  string       `json:"result"`
}

type eventMessage struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Name  string `json:"name"`
	Input any    `json:"input"`
}

// parseStreamJSON reads a stream-json log line by line, extracts the human-readable
// text from assistant messages, and writes it to the output writer.
// It returns the full accumulated text output.
func parseStreamJSON(r io.Reader, w io.Writer) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var fullText strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event streamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Type {
		case "assistant":
			if event.Message == nil {
				continue
			}
			for _, block := range event.Message.Content {
				switch block.Type {
				case "text":
					fmt.Fprint(w, block.Text)
					fullText.WriteString(block.Text)
				case "tool_use":
					msg := fmt.Sprintf("\n[tool: %s]\n", block.Name)
					fmt.Fprint(w, msg)
					fullText.WriteString(msg)
				}
			}

		case "result":
			if event.Subtype != "success" {
				msg := fmt.Sprintf("\n[result: %s]\n", event.Subtype)
				fmt.Fprint(w, msg)
				fullText.WriteString(msg)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fullText.String(), fmt.Errorf("reading stream: %w", err)
	}

	return fullText.String(), nil
}
