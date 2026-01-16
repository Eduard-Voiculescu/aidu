package analyzer

import (
	"strings"

	"github.com/eduardvoiculescu/aidu/pkg/types"
)

type Analyzer struct{}

func New() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) AnalyzeTask(description string) types.TaskComplexity {
	lower := strings.ToLower(description)

	if a.isSimpleQuestion(lower) {
		return types.ComplexitySimple
	}

	if a.isDataFetch(lower) {
		return types.ComplexityDataFetch
	}

	if a.isImplementation(lower) {
		return types.ComplexityImplement
	}

	return types.ComplexityComplex
}

func (a *Analyzer) isSimpleQuestion(text string) bool {
	simplePatterns := []string{
		"what is",
		"how much is",
		"calculate",
		"convert",
	}

	questionWords := []string{"what", "when", "where", "who", "how"}
	hasQuestion := false
	for _, qw := range questionWords {
		if strings.HasPrefix(text, qw) {
			hasQuestion = true
			break
		}
	}

	if !hasQuestion {
		return false
	}

	for _, pattern := range simplePatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}

	if len(strings.Split(text, " ")) < 15 && strings.Contains(text, "?") {
		return true
	}

	return false
}

func (a *Analyzer) isDataFetch(text string) bool {
	dataFetchKeywords := []string{
		"fetch",
		"retrieve",
		"get data",
		"download",
		"pull data",
		"query",
		"stock",
		"api",
	}

	for _, keyword := range dataFetchKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

func (a *Analyzer) isImplementation(text string) bool {
	implementKeywords := []string{
		"implement",
		"create",
		"build",
		"add",
		"develop",
		"write",
		"code",
		"function",
		"feature",
		"endpoint",
		"component",
		"service",
		"authentication",
		"authorization",
	}

	for _, keyword := range implementKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}
