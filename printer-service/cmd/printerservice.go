package main

import (
	"nantipov/printerservice/internal"
	"nantipov/printerservice/internal/dispatcher"
	"nantipov/printerservice/internal/machine"
	"nantipov/printerservice/internal/web"
)

func main() {
	internal.LoadSetting()
	internal.InitDatabase()
	machine.Init()
	dispatcher.Start()
	web.StartWebServer()
}
