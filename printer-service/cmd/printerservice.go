package main

import (
	"nantipov/printerservice/internal"
	"nantipov/printerservice/internal/jobprocessor"
	"nantipov/printerservice/internal/machine"
	"nantipov/printerservice/internal/web"
)

func main() {
	internal.LoadSetting()
	machine.Init()
	jobprocessor.StartDispatcher()
	web.StartWebServer()
}
