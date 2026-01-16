package types

import "time"

type TaskComplexity string

const (
	ComplexitySimple     TaskComplexity = "SIMPLE"
	ComplexityDataFetch  TaskComplexity = "DATA_FETCH"
	ComplexityImplement  TaskComplexity = "IMPLEMENT"
	ComplexityComplex    TaskComplexity = "COMPLEX"
)

type Task struct {
	ID          string
	Description string
	Complexity  TaskComplexity
	Context     string
	AgentsMD    string
	CreatedAt   time.Time
	RetryCount  int
	Feedback    string
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusRetrying  TaskStatus = "RETRYING"
)
