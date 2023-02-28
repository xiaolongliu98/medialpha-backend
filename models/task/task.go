package task

import (
	"time"
)

type TaskInfo struct {
	Name        string
	Message     string
	SubMsg      string
	Duration    time.Duration
	TotalSteps  int
	CurrentStep int
	StartTime   time.Time
	Error       error
	Returns     map[string]any
}

func NewTaskInfo(name string) *TaskInfo {
	t := &TaskInfo{
		Name:        name,
		Message:     "初始化中",
		StartTime:   time.Time{},
		Error:       nil,
		TotalSteps:  1,
		CurrentStep: 0,
	}
	return t
}

// totalSteps 有多少个Update+1
func (t *TaskInfo) Start(totalSteps int, message string) {
	t.Message = message
	t.StartTime = time.Now()
	t.TotalSteps = totalSteps
}

func (t *TaskInfo) Step(message string) {
	t.CurrentStep++
	t.Message = message
}

func (t *TaskInfo) UpdateMsg(message string) {
	t.Message = message
}

func (t *TaskInfo) UpdateSubMsg(message string) {
	t.SubMsg = message
}

func (t *TaskInfo) Success(returnVals ...*map[string]any) {
	if len(returnVals) > 0 {
		t.Returns = *returnVals[0]
	}
	t.CurrentStep++
	t.Duration = time.Now().Sub(t.StartTime)
	t.Message = "success"
	t.SubMsg = ""
}

func (t *TaskInfo) Abort(message string, err error, returnVals ...*map[string]any) {
	if len(returnVals) > 0 {
		t.Returns = *returnVals[0]
	}
	t.Duration = time.Now().Sub(t.StartTime)
	t.Error = err
	t.Message = message
	t.SubMsg = ""
}

func (t *TaskInfo) Stopped() bool {
	return t == nil || t.Duration != 0 || t.Error != nil || (t.Message == "success" && t.SubMsg == "")
}

func (t *TaskInfo) ToTaskInfoResp() *TaskInfoResp {
	if t == nil {
		return &TaskInfoResp{}
	}
	data := &TaskInfoResp{
		Name:        t.Name,
		Message:     t.Message,
		Duration:    int(t.Duration.Milliseconds()),
		TotalSteps:  t.TotalSteps,
		CurrentStep: t.CurrentStep,
		StartTime:   int(t.StartTime.UnixMilli()),
		Returns:     t.Returns,
		Running:     !t.Stopped(),
	}
	if t.Error != nil {
		data.ErrorStr = t.Error.Error()
	}
	if t.SubMsg != "" {
		data.Message += ": " + t.SubMsg
	}

	return data
}

type TaskInfoResp struct {
	Name        string
	Running     bool
	Message     string
	Duration    int
	TotalSteps  int
	CurrentStep int
	StartTime   int
	Returns     map[string]any
	ErrorStr    string
}
