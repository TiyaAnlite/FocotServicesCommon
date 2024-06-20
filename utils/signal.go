package utils

import (
	"os"
	"os/signal"
)

func Wait4CtrlC() {
	<-CtrlCChannel()
}

func CtrlCChannel() chan os.Signal {
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	return ctrlc
}
