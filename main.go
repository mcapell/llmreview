package main

import (
	"log/slog"

	"github.com/mcapell/llmreview/task"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	runner := task.NewRunner("./llmreview.yaml")

	if err := runner.Run(); err != nil {
		slog.Error("error running the task runner", "error", err)
	}

	slog.Info("done")
}
