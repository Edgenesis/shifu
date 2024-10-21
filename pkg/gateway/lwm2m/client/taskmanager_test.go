package client

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskManager := NewTaskManager(ctx)

	var mu sync.Mutex
	var executionCount int

	fn := func() {
		mu.Lock()
		defer mu.Unlock()
		executionCount++
	}

	taskManager.AddTask("task1", 100*time.Millisecond, fn)

	// Let the task run for a while
	time.Sleep(350 * time.Millisecond)

	mu.Lock()
	count := executionCount
	mu.Unlock()

	assert.GreaterOrEqual(t, count, 3, "Task should be executed at least 3 times")

	// Ensure the task gets canceled properly
	taskManager.CancelTask("task1")
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	newCount := executionCount
	mu.Unlock()

	assert.Equal(t, count, newCount, "Task should not run after being canceled")
}

func TestTaskManager_ResetTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskManager := NewTaskManager(ctx)

	var mu sync.Mutex
	var executionCount int

	fn := func() {
		mu.Lock()
		defer mu.Unlock()
		executionCount++
	}

	taskManager.AddTask("task2", 100*time.Millisecond, fn)

	// Let the task run for a while
	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	count := executionCount
	mu.Unlock()

	taskManager.ResetTask("task2")

	// Wait for a few more ticks to ensure reset worked
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	newCount := executionCount
	mu.Unlock()

	assert.Greater(t, newCount, count, "Task should continue after being reset")
}

func TestTaskManager_CancelAllTasks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskManager := NewTaskManager(ctx)

	var mu sync.Mutex
	var executionCount1, executionCount2 int

	fn1 := func() {
		mu.Lock()
		defer mu.Unlock()
		executionCount1++
	}

	fn2 := func() {
		mu.Lock()
		defer mu.Unlock()
		executionCount2++
	}

	taskManager.AddTask("task4", 100*time.Millisecond, fn1)
	taskManager.AddTask("task5", 100*time.Millisecond, fn2)

	time.Sleep(250 * time.Millisecond)

	taskManager.CancelAllTasks()

	// Capture execution counts after cancellation
	mu.Lock()
	count1 := executionCount1
	count2 := executionCount2
	mu.Unlock()

	// Wait and verify no further execution
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	newCount1 := executionCount1
	newCount2 := executionCount2
	mu.Unlock()

	assert.Equal(t, count1, newCount1, "Task1 should not execute after being canceled")
	assert.Equal(t, count2, newCount2, "Task2 should not execute after being canceled")
}
