package task

import (
	"fmt"
	"sync"
)

type TaskHandler struct {
	currentTask *TaskInfo
	busy        bool
	taskCh      chan string
	params      map[string]any
	lock        *sync.RWMutex
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		currentTask: nil,
		busy:        false,
		taskCh:      make(chan string, 1),
		params:      map[string]any{},
		lock:        &sync.RWMutex{},
	}
}

func (t *TaskHandler) IsBusy() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.busy
}

func (t *TaskHandler) WaitTask() <-chan string {
	return t.taskCh
}

func (t *TaskHandler) SubmitTask(name string, args ...*map[string]any) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.busy {
		return fmt.Errorf("正在执行任务")
	}
	t.busy = true

	taskInfo := NewTaskInfo(name)

	oldParams := t.params
	if len(args) > 0 {
		t.params = *args[0]
	}
	oldTask := t.currentTask
	t.currentTask = taskInfo

	select {
	case t.taskCh <- taskInfo.Name:
	default:
		t.busy = false
		t.currentTask = oldTask
		t.params = oldParams
		return fmt.Errorf("正在执行任务")
	}

	return nil
}

func (t *TaskHandler) CurrentTaskInfo() *TaskInfo {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.currentTask
}

func (t *TaskHandler) AcceptTask(name string) (*TaskInfo, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if !t.busy || t.currentTask == nil {
		return nil, fmt.Errorf("当前无任务")
	}

	if t.currentTask.Name != name {
		return nil, fmt.Errorf("不存在目标任务")
	}

	return t.currentTask, nil
}

// with defer
func (t *TaskHandler) FinishTask() {
	t.lock.Lock()
	defer t.lock.Unlock()
	//if !t.busy || t.currentTask == nil {
	//	return
	//}
	t.busy = false
}

func (t *TaskHandler) GetTaskStatus(name string) (*TaskInfo, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.currentTask == nil || t.currentTask.Name != name {
		return nil, fmt.Errorf("当前无该任务")
	}
	return t.currentTask, nil
}

func (t *TaskHandler) Params() map[string]any {
	return t.params
}

func (t *TaskHandler) ClearTask() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.busy {
		return false
	}

	t.params = nil
	t.currentTask = nil
	return true
}