package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Michael-F-Ellis/wasmskel/cmd/internal/common"
)

var MonitoredParametersState = common.MonitoredParameters{}

func main() {
	go updater()
	mux := http.NewServeMux()
	mux.HandleFunc("/wasm_exec.js", wasmExecRequestHandler)
	mux.HandleFunc("/json.wasm", jsonWasmRequestHandler)
	mux.HandleFunc("/get", getRequestHandler)
	mux.HandleFunc("/", indexRequestHandler)
	log.Fatal(http.ListenAndServe(":9090", mux))
}

func indexRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./assets/index.html")
}

func wasmExecRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./assets/wasm_exec.js")
}

func jsonWasmRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./assets/json.wasm")
}

func getRequestHandler(w http.ResponseWriter, r *http.Request) {
	mpStateP := MonitoredParametersState.Get()
	jsonRecord, err := GetJSON(mpStateP)
	if err != nil { // should never happen in this scenario
		err = fmt.Errorf("can't marshal the record: %v", err)
		fail(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// good to go
	success(w, jsonRecord)
}

// success sends a JSON formatted response. It's called when a request for data
// succeeds.
func success(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// fail sends a JSON formatted response. It's called when we detect a problem in
// a handler func.
func fail(w http.ResponseWriter, msg string, status int) {
	responseJSON, _ := json.Marshal(map[string]string{"Err": msg})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(responseJSON)
}

// GetJSON returns a JSON representation of the values of
// MonitorParameters that are part of the JSON API.
func GetJSON(mp *MonitoredParameters) (jsn []byte, err error) {

	mpcopy := mp.Get()
	jsn, err = json.Marshal(mpcopy)
	return
}

// updater continually changes MonitoredParameters state
func updater() {
	f := func(p *MonitoredParameters) {
		p.A += 1
		p.B += 2
	}
	for {
		time.Sleep(time.Second)
		MonitoredParametersState.directUpdate(f)
	}
}
