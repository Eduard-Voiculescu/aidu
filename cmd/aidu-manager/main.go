package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eduardvoiculescu/aidu/pkg/manager"
)

func main() {
	var (
		namespace       = flag.String("namespace", "aidu-sessions", "Kubernetes namespace")
		maxRetries      = flag.Int("max-retries", 3, "Maximum number of retry attempts")
		timeout         = flag.Duration("timeout", 10*time.Minute, "Worker execution timeout")
		validationScore = flag.Float64("validation-score", 0.75, "Minimum validation score threshold")
		workerImage     = flag.String("worker-image", "localhost:5050/aidu-claude:latest", "Worker container image")
		agentsMDPath    = flag.String("agents-file", "./AGENTS.md", "Path to AGENTS.md file")
		outputDir       = flag.String("output-dir", "./output", "Output directory for results")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <task description>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "AIDU Manager - Orchestrates worker pods to complete tasks\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s \"implement user authentication\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --max-retries 5 \"build a REST API\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --agents-file ./my-agents.md \"add tests\"\n", os.Args[0])
	}

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: task description required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	taskDescription := flag.Arg(0)

	cfg := &manager.ManagerConfig{
		Namespace:       *namespace,
		MaxRetries:      *maxRetries,
		Timeout:         *timeout,
		ValidationScore: *validationScore,
		WorkerImage:     *workerImage,
		AgentsMDPath:    *agentsMDPath,
		OutputDir:       *outputDir,
	}

	mgr, err := manager.NewManager(cfg)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	log.Printf("Starting task execution...")
	log.Printf("Task: %s", taskDescription)
	log.Printf("Max retries: %d", *maxRetries)
	log.Printf("Timeout: %s", *timeout)
	log.Printf("Validation threshold: %.2f", *validationScore)
	log.Println("---")

	result, err := mgr.ExecuteTask(ctx, taskDescription)
	if err != nil {
		log.Fatalf("Task execution failed: %v", err)
	}

	log.Println("---")
	log.Println("Task completed successfully!")
	log.Printf("Output path: %s", result.OutputPath)
	log.Printf("Execution time: %.2f seconds", result.ExecutionTime)
	log.Printf("Validation score: %.2f", result.Validation.Score)

	if len(result.Validation.Warnings) > 0 {
		log.Println("\nWarnings:")
		for _, warning := range result.Validation.Warnings {
			log.Printf("  - %s", warning)
		}
	}

	os.Exit(0)
}
