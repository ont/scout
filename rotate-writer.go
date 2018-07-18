package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type RotateWriter struct {
	fd *os.File

	base   string
	ext    string
	rotate time.Duration

	lock sync.Mutex
}

func NewRotateWriter(filename string, rotate time.Duration) *RotateWriter {
	ext := filepath.Ext(filename)
	w := &RotateWriter{
		base:   strings.TrimSuffix(filename, ext),
		ext:    ext,
		rotate: rotate,
	}

	if rotate > 0 {
		go w.Watch()
	}

	w.Rotate(time.Now())

	return w
}

func (w *RotateWriter) Write(bytes []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.fd.Write(bytes)
}

func (w *RotateWriter) Watch() {
	ticker := time.NewTicker(w.rotate)
	for t := range ticker.C {
		w.Rotate(t)
	}
}

func (w *RotateWriter) Rotate(t time.Time) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.fd != nil {
		if err := w.fd.Close(); err != nil {
			log.Fatal("Error during RotateWriter.fd.Close():", err)
		}
	}

	var filename string
	if w.rotate > 0 {
		filename = w.base + t.Format("-2006-01-02_15:04:05") + w.ext
	} else {
		filename = w.base + w.ext
	}

	var err error
	w.fd, err = os.Create(filename)

	if err != nil {
		log.Fatal("Error during RotateWriter new file creation:", err)
	}
}
