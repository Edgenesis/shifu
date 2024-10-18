package client

import (
	"context"
	"sync"
	"time"

	"github.com/edgenesis/shifu/pkg/logger"
)

type Task struct {
	interval   time.Duration
	Ticker     *time.Ticker
	CancelFunc context.CancelFunc
}

type TaskManager struct {
	Tasks map[string]*Task
	Lock  sync.Mutex

	ctx context.Context
}

func NewTaskManager(ctx context.Context) *TaskManager {
	return &TaskManager{
		Tasks: make(map[string]*Task),
		ctx:   ctx,
	}
}

// AddTask adds a new task to the task manager.
// The task will be executed every interval and run the given function and assign the task to the given id.
func (m *TaskManager) AddTask(id string, interval time.Duration, fn func()) {
	logger.Infof("AddTask: %s", id)
	m.Lock.Lock()
	defer m.Lock.Unlock()

	// Check if the task already exists
	if _, exists := m.Tasks[id]; exists {
		return
	}

	ticker := time.NewTicker(interval)
	ctx, cancel := context.WithCancel(m.ctx)
	task := &Task{
		Ticker:     ticker,
		CancelFunc: cancel,
		interval:   interval,
	}
	m.Tasks[id] = task

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-task.Ticker.C:
				fn()
			}
		}
	}()
}

// reset the task ticker for the given id
func (m *TaskManager) ResetTask(id string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	task, exists := m.Tasks[id]
	if !exists {
		return
	}

	task.Ticker.Reset(task.interval)
}

// CancelTask cancels the task with the given id.
func (m *TaskManager) CancelTask(id string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	task, exists := m.Tasks[id]
	if !exists {
		return
	}

	task.CancelFunc()
	task.Ticker.Stop()
	delete(m.Tasks, id)
}

// CancelAllTasks cancels all the tasks in the task manager.
func (m *TaskManager) CancelAllTasks() {
	for taskId := range m.Tasks {
		m.CancelTask(taskId)
	}
}
