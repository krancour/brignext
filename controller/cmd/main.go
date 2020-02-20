package main

import (
	"github.com/krancour/brignext/pkg/signals"
)

func main() {
	ctx := signals.Context()
	<-ctx.Done()
}
