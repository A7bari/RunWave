package taskqueue

import "encoding/json"

type TaskMsg struct {
	// TaskID is the unique identifier of the task
	TaskID string `json:"task_id"`

	// Lang is the language of the code
	Lang string `json:"lang"`

	// Code is the code to be executed
	Code string `json:"code"`

	// RetryCnt is the retry count
	RetryCnt int `json:"retry_cnt"`
}

// NewTaskMsg creates a new TaskMsg
func NewTaskMsg(taskID, lang, code string, retryCnt int) *TaskMsg {
	return &TaskMsg{
		TaskID:   taskID,
		Lang:     lang,
		Code:     code,
		RetryCnt: retryCnt,
	}
}

// Decode decodes the TaskMsg from a byte slice
func Decode(data []byte) (*TaskMsg, error) {
	taskMsg := &TaskMsg{}
	err := json.Unmarshal(data, taskMsg)
	if err != nil {
		return nil, err
	}
	return taskMsg, nil
}

// Encode encodes the TaskMsg to a byte slice
func (t *TaskMsg) Encode() ([]byte, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return data, nil
}
