package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Michael-F-Ellis/wasmskel/internal/common"
)

var State = common.State{}

func main() {
	go updater()
	mux := http.NewServeMux()
	mux.HandleFunc("/wasm_exec.js", wasmExecRequestHandler)
	mux.HandleFunc("/json.wasm", jsonWasmRequestHandler)
	mux.HandleFunc("/get", getRequestHandler)
	mux.HandleFunc("/set", setRequestHandler)
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
	stateP := State.Get()
	jsonRecord, err := GetJSON(stateP)
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
func GetJSON(sp *common.State) (jsn []byte, err error) {

	mpcopy := sp.Get()
	jsn, err = json.Marshal(mpcopy)
	return
}

// updater continually changes MonitoredParameters state
func updater() {
	f := func(p *common.State) {
		p.Alpha += 1
		p.Beta += 2
	}
	for {
		time.Sleep(time.Second)
		State.DirectUpdate(f)
	}
}

// //////////////////////////
// Handler for /set requests
// //////////////////////////

// setRequestHandler processes JSON requests that specify new
// values for changeable parameters.
func setRequestHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	// try to read the request body
	jsn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fail(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()
	// Extract the json to a map of RawMessage values
	// to allow piecemeal unmarshalling of fields
	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(jsn, &objmap)
	if err != nil {
		fail(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Permit only 1 item to be changed
	if len(objmap) != 1 {
		err = fmt.Errorf("only one item per set request, please, found %d", len(objmap))
		fail(w, err.Error(), http.StatusBadRequest)
		return
	}

	for name, rawval := range objmap {
		err = dispatcher(name, rawval)
		if err != nil {
			err = fmt.Errorf("couldn't set new value for %s: %v", name, err)
			fail(w, err.Error(), http.StatusBadRequest)
			return
		}

		success(w, []byte(`{"Err":null}`))
		return

	}

}
