package main

import (
	"io"
	"sync"
	"time"
)

type TeeRotator struct {
	fabrics []func(stamp string) io.Writer

	lock sync.Mutex
	out  io.Writer // output of io.MultiWriter()
}

func NewTeeRotator(fabrics ...func(stamp string) io.Writer) *TeeRotator {
	tee := &TeeRotator{
		fabrics: fabrics,
	}

	tee.Rotate(time.Now())

	return tee
}

func (t *TeeRotator) RotateEvery(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for ttime := range ticker.C {
		t.Rotate(ttime)
	}
}

func (t *TeeRotator) Rotate(ttime time.Time) {
	t.lock.Lock()
	defer t.lock.Unlock()

	outs := make([]io.Writer, len(t.fabrics))
	for i, fabric := range t.fabrics {
		outs[i] = fabric(ttime.Format("2006-01-02_15:04:05")) // TODO: configurable stamp format
	}
	t.out = io.MultiWriter(outs...)
}

func (t *TeeRotator) Write(bytes []byte) (int, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.out.Write(bytes)
}
