package main

type Bag interface {
	Write(pair *ReqResPair) error
}
