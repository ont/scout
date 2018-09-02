package main

type Storage interface {
	Save(pair *ReqResPair) error
	Close() error
}
