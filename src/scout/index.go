package main

import (
	bolt "go.etcd.io/bbolt"
)

type Index struct {
	db *bolt.DB

	offset int // TODO: protect with sync.Lock ?
}

func NewIndex(fname string) (*Index, error) {
	db, err := bolt.Open(fname, 0666, nil)
	if err != nil {
		return nil, err
	}

	return &Index{
		db: db,
	}, nil
}

func (i *Index) Write(bytes []byte) (int, error) {

	i.offset += len(bytes)
	return len(bytes), nil
}
