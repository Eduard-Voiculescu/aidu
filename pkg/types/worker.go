package types

import "time"

type WorkerState string

const (
	WorkerStatePending   WorkerState = "PENDING"
	WorkerStateRunning   WorkerState = "RUNNING"
	WorkerStateCompleted WorkerState = "COMPLETED"
	WorkerStateFailed    WorkerState = "FAILED"
)

type Worker struct {
	ID        string
	TaskID    string
	PodName   string
	State     WorkerState
	StartedAt time.Time
	EndedAt   time.Time
	ExitCode  int
	Logs      string
}
