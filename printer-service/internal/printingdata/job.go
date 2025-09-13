package printingdata

import (
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type JobState int

const (
	JobStateDraft JobState = iota
	JobStatePosted
	JobStateCompiling
	JobStateCompiled
	JobStatePrinting
	JobStatePrinted
	JobStateCompleted
	JobStateFailed
)

type IppData struct {
	Id uint32
}

type Job struct {
	Id    string
	State JobState
	Error error
	mutex sync.RWMutex
	Ipp   IppData
}

var (
	ippCounter atomic.Uint32
)

func NewDraftJob() Job {
	return Job{
		Id:    uuid.NewString(),
		State: JobStateDraft,
		Ipp: IppData{
			Id: ippCounter.Add(1),
		},
	}
}

func (j *Job) AttachDocument(filename string) {
	//todo: content-type?
}

func (j *Job) GetState(state JobState) JobState {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	return j.State
}

func (j *Job) SetState(state JobState) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.State = state
}
