package config

import (
	"fmt"

	"github.com/eduardvoiculescu/aidu/pkg/types"
)

type TaskContext struct {
	Task     *types.Task
	AgentsMD string
	Prompt   string
}

func BuildTaskContext(task *types.Task, agentsMD string) *TaskContext {
	ctx := &TaskContext{
		Task:     task,
		AgentsMD: agentsMD,
	}

	ctx.Prompt = ctx.generatePrompt()
	return ctx
}

func (tc *TaskContext) generatePrompt() string {
	prompt := fmt.Sprintf(`# Task: %s

## Description
%s

## Complexity Level
%s

`, tc.Task.ID, tc.Task.Description, tc.Task.Complexity)

	if tc.Task.RetryCount > 0 {
		prompt += fmt.Sprintf(`## Previous Attempts
This is retry #%d. Previous attempt had issues:
%s

Please address the feedback above and provide a better solution.

`, tc.Task.RetryCount, tc.Task.Feedback)
	}

	prompt += fmt.Sprintf(`## Agent Context
%s

## Instructions
Please complete the task described above. Your output should be:
- Complete and functional
- Well-tested
- Clean and maintainable
- Following the guidelines in the agent context

Write all code, tests, and documentation needed to complete this task.
`, tc.AgentsMD)

	return prompt
}

func (tc *TaskContext) ToConfigMapData() map[string]string {
	return map[string]string{
		"task.md":    tc.Prompt,
		"AGENTS.md":  tc.AgentsMD,
		"task_id":    tc.Task.ID,
		"complexity": string(tc.Task.Complexity),
	}
}
