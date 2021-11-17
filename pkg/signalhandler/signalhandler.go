package signalhandler

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type sh struct {
	c <-chan os.Signal
}

type SignalHandler interface {
	// CatchSignals catches os.Interrupt, syscall.SIGTERM and cancels returned context
	CatchSignals(parentCtx context.Context) context.Context
}

func NewSignalHandler() SignalHandler {
	c := make(chan os.Signal, 1)
	// os.Kill cannot be trapped
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	return &sh{
		c: c,
	}
}

func (s *sh) CatchSignals(parentCtx context.Context) context.Context {
	ctx, cancel := context.WithCancel(parentCtx)

	go func() {
		<-s.c
		cancel()
	}()

	return ctx
}
