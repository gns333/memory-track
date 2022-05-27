package main

import (
	"os"
	"os/signal"
	"syscall"
)

func init() {
	setupSignHandler()
}

func setupSignHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		StopRecordMem()
	}()
}
