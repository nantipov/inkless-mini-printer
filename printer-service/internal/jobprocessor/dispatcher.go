package jobprocessor

import (
	"log"
	"nantipov/printerservice/internal/compiler"
	"nantipov/printerservice/internal/machine"
	"nantipov/printerservice/internal/printingdata"
	"sync"
	"time"

	"github.com/OpenPrinting/goipp"
)

var (
	mutex sync.RWMutex
	jobs  = []*printingdata.Job{}

	ticker      = time.NewTicker(10 * time.Second)
	doneChannel = make(chan bool)
)

func CreateDraftJob(ippAttributes goipp.Attributes) (*printingdata.Job, error) {
	mutex.Lock()
	defer mutex.Unlock()
	job := printingdata.NewDraftJob()

	jobs = append(jobs, &job)

	return &job, nil
}

func FindJob(ippId int) (*printingdata.Job, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	return nil, nil
}

func StartDispatcher() {
	go dispatcherDuty()
}

func dispatcherDuty() {
	log.Println("start dispatcher")
	for {
		select {
		case <-doneChannel:
			return
		case <-ticker.C:
			mutex.RLock()
			for _, job := range jobs {
				dispatch(job)
			}
			mutex.RUnlock()
		}
	}
}

func dispatch(job *printingdata.Job) {
	switch job.State {
	case printingdata.JobStatePosted:
		compile(job)
	case printingdata.JobStateCompiled:
		print(job)
	}
}

func compile(job *printingdata.Job) {
	//todo: thread pool? limit number of running and sleeping tasks/threads
	go compiler.Compile(job)
	job.State = printingdata.JobStateCompiling
}

func print(job *printingdata.Job) {
	acceptedForPrinting := machine.PrintIfReady(job)
	if acceptedForPrinting {
		job.State = printingdata.JobStatePrinting
	}
}
