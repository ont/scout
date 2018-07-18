package main

import (
	"net/http"
	_ "net/http/pprof"
)

func RunProfiler() {
	go http.ListenAndServe("0.0.0.0:8080", nil)
}
