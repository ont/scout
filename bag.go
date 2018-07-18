package main

type Bag interface {
	Write(pair *ReqResPair, bytes []byte) error
}
