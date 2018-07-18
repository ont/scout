package main

import (
	"net/http"
	_ "net/http/pprof"
)

func RunProfiler(port string) {
	go http.ListenAndServe("0.0.0.0:"+port, nil)
}
