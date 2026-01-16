package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eduardvoiculescu/aidu/pkg/types"
)

type Validator struct {
	acceptanceThreshold float64
}

func New(threshold float64) *Validator {
	if threshold <= 0 {
		threshold = 0.75
	}
	return &Validator{
		acceptanceThreshold: threshold,
	}
}

func (v *Validator) ValidateOutput(outputPath string, task *types.Task) types.ValidationResult {
	result := types.ValidationResult{
		Valid:    true,
		Score:    1.0,
		Errors:   []string{},
		Warnings: []string{},
	}

	if err := v.checkOutputExists(outputPath); err != nil {
		result.Valid = false
		result.Score = 0.0
		result.Errors = append(result.Errors, err.Error())
		result.Feedback = "No output was produced. Please ensure the task completes and writes output."
		result.ShouldRetry = true
		return result
	}

	logs, _ := os.ReadFile(filepath.Join(outputPath, "logs.txt"))
	exitCodeBytes, _ := os.ReadFile(filepath.Join(outputPath, "exit_code"))

	errorChecks := v.checkForErrors(string(logs))
	if len(errorChecks) > 0 {
		result.Errors = append(result.Errors, errorChecks...)
		result.Score -= 0.3
	}

	if string(exitCodeBytes) != "0\n" && string(exitCodeBytes) != "" {
		result.Errors = append(result.Errors, fmt.Sprintf("Worker exited with non-zero code: %s", strings.TrimSpace(string(exitCodeBytes))))
		result.Score -= 0.2
	}

	completenessScore := v.checkCompleteness(outputPath, task)
	result.Score = (result.Score + completenessScore) / 2

	if result.Score < v.acceptanceThreshold {
		result.Valid = false
		result.ShouldRetry = true
		result.Feedback = v.generateFeedback(result)
	}

	return result
}

func (v *Validator) checkOutputExists(outputPath string) error {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist")
	}

	entries, err := os.ReadDir(outputPath)
	if err != nil {
		return fmt.Errorf("cannot read output directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("output directory is empty")
	}

	return nil
}

func (v *Validator) checkForErrors(logs string) []string {
	errors := []string{}
	lower := strings.ToLower(logs)

	errorPatterns := []string{
		"error:",
		"fatal:",
		"panic:",
		"failed:",
		"exception:",
	}

	for _, pattern := range errorPatterns {
		if strings.Contains(lower, pattern) {
			errors = append(errors, fmt.Sprintf("Found error pattern in logs: %s", pattern))
		}
	}

	return errors
}

func (v *Validator) checkCompleteness(outputPath string, task *types.Task) float64 {
	score := 1.0

	entries, err := os.ReadDir(outputPath)
	if err != nil {
		return 0.0
	}

	hasCode := false
	hasTests := false

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".py") ||
			strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".ts") {
			hasCode = true
		}
		if strings.Contains(name, "test") || strings.Contains(name, "_test") {
			hasTests = true
		}
	}

	if task.Complexity == types.ComplexityImplement || task.Complexity == types.ComplexityComplex {
		if !hasCode {
			score -= 0.4
		}
		if !hasTests {
			score -= 0.2
		}
	}

	return score
}

func (v *Validator) generateFeedback(result types.ValidationResult) string {
	feedback := "The output needs improvement:\n\n"

	if len(result.Errors) > 0 {
		feedback += "Errors found:\n"
		for _, err := range result.Errors {
			feedback += fmt.Sprintf("- %s\n", err)
		}
		feedback += "\n"
	}

	if len(result.MissingFiles) > 0 {
		feedback += "Missing expected files:\n"
		for _, file := range result.MissingFiles {
			feedback += fmt.Sprintf("- %s\n", file)
		}
		feedback += "\n"
	}

	feedback += fmt.Sprintf("Quality score: %.2f (threshold: %.2f)\n", result.Score, v.acceptanceThreshold)
	feedback += "\nPlease address these issues and provide a complete, working solution."

	return feedback
}
