package main

import (
	"log"
	"os"
)

type StorageBase struct {
	out *os.File
}

func (s *StorageBase) Save(pair *ReqResPair) error {
	bytes, err := pair.AsJson()
	if err != nil {
		return err
	}

	_, err = s.out.Write(bytes)
	if err != nil {
		return err
	}

	_, err = s.out.Write([]byte{'\n'})
	return err
}

func (s *StorageBase) Close() error {
	return s.out.Close()
}

func NewStorageBase(basename string) *StorageBase {
	out, err := os.Create(basename + ".json")
	if err != nil {
		log.Fatal("Can't create json file (storage):", err)
	}

	return &StorageBase{out}
}
