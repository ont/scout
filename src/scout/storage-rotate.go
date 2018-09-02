package main

import (
	"log"
	"sync"
	"time"
)

type StorageRotate struct {
	fabric func(stamp string) Storage

	ticker  time.Ticker
	lock    sync.Mutex
	storage Storage
}

func NewRotateStorage(fabric func(stamp string) Storage) *StorageRotate {
	rstorage := &StorageRotate{
		fabric: fabric,
	}

	rstorage.Rotate(time.Now())

	return rstorage
}

func (s *StorageRotate) RotateEvery(interval time.Duration) {
	if interval == 0 {
		return
	}

	ticker := time.NewTicker(interval)
	for ttime := range ticker.C {
		s.Rotate(ttime)
	}
}

func (s *StorageRotate) Rotate(stamp time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.storage.Close(); err != nil {
		log.Fatal("Can't close storage during rotation:", err)
	}
	s.storage = s.fabric(stamp.Format("2006-01-02_15:04:05"))
}

func (s *StorageRotate) Save(pair *ReqResPair) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.storage.Save(pair)
}

func (s *StorageRotate) Close() error {
	s.ticker.Stop()
	return s.storage.Close()
}
