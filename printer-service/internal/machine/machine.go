package machine

import (
	"log"
	"nantipov/printerservice/internal"
	"nantipov/printerservice/internal/machine/grbl"
	"nantipov/printerservice/internal/printingdata"
	"sync"

	"go.bug.st/serial"
)

var (
	port         serial.Port
	readyToPrint bool = false //todo: atomic?
	machineMutex sync.RWMutex
	currentJob   *printingdata.Job
)

func Init() {
	settings := internal.GetSettings()

	connect()
	version := grbl.GetVersion(port)

	if settings.GrblDeviceFirmwareVersion != version {
		//todo: always update to higher version
		//todo: should a version be an integer value?
		log.Printf("machine firmware version mismatch: available: %s, installed: %s", settings.GrblDeviceFirmwareVersion, version)
		updateFirmware()
	}

	readyToPrint = true
}

//todo: consider atomic printing - e.g. printIfReady() bool
// func IsReadyForPrint() bool {
// 	taskMutex.RLock()
// 	defer taskMutex.RUnlock()
// 	return readyToPrint
// }

func PrintIfReady(job *printingdata.Job) bool {
	machineMutex.Lock()
	defer machineMutex.Unlock()

	if !readyToPrint {
		return false
	}

	readyToPrint = false
	currentJob = job
	return true
}

func connect() { //todo: return error for further hanlding
	settings := internal.GetSettings()

	log.Printf("connect to printing machine controller: %s", settings.GrblDevicePort)

	//todo: from settings?
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	var err error
	port, err = serial.Open(settings.GrblDevicePort, mode)

	//todo: auto reconnect?

	if err != nil {
		log.Fatalf("could not connect to device: %s: %s", settings.GrblDevicePort, err.Error())
	}
}

func updateFirmware() {
	log.Printf("updating machine firmware")
}
