package lwm2m

import (
	"sync"
	"time"

	"github.com/edgenesis/shifu/pkg/logger"
)

type Task struct {
	interval time.Duration
	Ticker   *time.Ticker
	Quit     chan struct{}
}

type TaskManager struct {
	Tasks map[string]*Task
	Lock  sync.Mutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		Tasks: make(map[string]*Task),
	}
}

func (m *TaskManager) AddTask(id string, interval time.Duration, fn func()) {
	logger.Infof("AddTask: %s", id)
	m.Lock.Lock()
	defer m.Lock.Unlock()

	if _, exists := m.Tasks[id]; exists {
		return
	}

	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	task := &Task{Ticker: ticker, Quit: quit, interval: interval}
	m.Tasks[id] = task

	go func() {
		for {
			select {
			case <-task.Quit:
				return
			case <-task.Ticker.C:
				fn()
			}
		}
	}()
}

func (m *TaskManager) ResetTask(id string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	task, exists := m.Tasks[id]
	if !exists {
		return
	}

	task.Ticker.Reset(task.interval)
}

func (m *TaskManager) CancelTask(id string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	task, exists := m.Tasks[id]
	if !exists {
		return
	}

	task.Ticker.Stop()
	close(task.Quit)
	delete(m.Tasks, id)
}

func (m *TaskManager) CancelAllTasks() {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	for _, task := range m.Tasks {
		task.Ticker.Stop()
		close(task.Quit)
	}
	m.Tasks = make(map[string]*Task)
}
