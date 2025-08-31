package dispatcher

import (
	"log"

	"github.com/OpenPrinting/goipp"
)

func CreateDraftJob(ippAttributes goipp.Attributes) (*Job, error) {
	return nil, nil
}

func FindJob(ippId int) (*Job, error) {
	return nil, nil
}

func Start() {
	go dispatcherDuty()
}

func dispatcherDuty() {
	log.Println("start dispatcher")
	// for ;; {

	// }
}
