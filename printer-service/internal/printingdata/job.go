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

	DocumentTypePDF    = "application/pdf"
	DocumentTypeRaster = "application/raster" //todo: revisit mime type
)

type IppDocument struct {
	Filename     string
	DocumentType string
}

type IppData struct {
	Id        uint32
	Documents []IppDocument
}

type Job struct {
	Id    string
	State JobState
	Error error
	Ipp   IppData
	Pages []*SketchedPage
	mutex sync.RWMutex
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

func (j *Job) AddDocument(filename string, documentType string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.Ipp.Documents = append(j.Ipp.Documents, IppDocument{Filename: filename, DocumentType: documentType})
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
