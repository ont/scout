package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func RunProfiler(port string) {
	go func() {
		err := http.ListenAndServe("0.0.0.0:"+port, nil)
		if err != nil {
			log.Println("ERROR: can't start profiler: ", err)
		}
	}()
}
