package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const DefaultAgentsMD = `# AIDU Agent Context

## Project Overview
This is a general purpose AIDU session. The agent should work on the task provided
with best practices and clean, maintainable code.

## Coding Standards
- Follow language-specific best practices
- Use clear, descriptive variable names
- Handle errors appropriately
- Add comments only where necessary for clarity

## Testing Requirements
- Write tests for new functionality
- Ensure existing tests pass
- Document test coverage

## Success Criteria
- Code compiles/runs without errors
- Tests pass
- Code is clean and maintainable
`

type AgentsConfig struct {
	Content string
	Path    string
}

func LoadAgentsMD(path string) (*AgentsConfig, error) {
	if path == "" {
		path = "./AGENTS.md"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AgentsConfig{
				Content: DefaultAgentsMD,
				Path:    absPath,
			}, nil
		}
		return nil, fmt.Errorf("reading AGENTS.md: %w", err)
	}

	return &AgentsConfig{
		Content: string(content),
		Path:    absPath,
	}, nil
}
