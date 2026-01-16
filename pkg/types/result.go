package types

type ValidationResult struct {
	Valid          bool
	Score          float64
	Errors         []string
	Warnings       []string
	MissingFiles   []string
	Feedback       string
	ShouldRetry    bool
}

type TaskResult struct {
	TaskID         string
	WorkerID       string
	Success        bool
	OutputPath     string
	Artifacts      []string
	Validation     ValidationResult
	ExecutionTime  float64
}
