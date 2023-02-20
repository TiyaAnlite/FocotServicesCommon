package utils

import (
	"os"
	"os/signal"
)

func Wait4CtrlC() {
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	<-ctrlc
}
