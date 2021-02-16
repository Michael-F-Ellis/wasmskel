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

// This is the global state that is shared via a JSON API.
var State = common.State{}

// Main launches a goroutine that continually updates the global state. Then it
// defines the allowed http requests and enters a ListenAndServe loop.
func main() {
	go Updater()
	mux := http.NewServeMux()
	mux.HandleFunc("/wasm_exec.js", wasmExecRequestHandler)
	mux.HandleFunc("/app.wasm", appWasmRequestHandler)
	mux.HandleFunc("/get", getRequestHandler)
	mux.HandleFunc("/set", setRequestHandler)
	mux.HandleFunc("/", indexRequestHandler)
	log.Fatal(http.ListenAndServe(":9090", mux))
}

// indexRequestHandler serves the index web page.
func indexRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./assets/index.html")
}

// wasmExecRequestHandler serves the Go wasm_exec.js file required for interface
// Go WebAssembly code to a browser's javascript system.
func wasmExecRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./assets/wasm_exec.js")
}

// appWasmRequestHandler serves the app's compiled WebAssembly file.
func appWasmRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./assets/app.wasm")
}

// getRequestHandler sends the global state in JSON encoded format.
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

// Updater continually changes MonitoredParameters state, simulating
// an arbitrary back-end process.
func Updater() {
	f := func(p *common.State) {
		p.Alpha += 1
		p.Beta += 2
	}
	for {
		time.Sleep(time.Second)
		State.DirectUpdate(f)
	}
}

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

	// Dispatch the request and return the result
	for name, rawval := range objmap {
		err = Dispatcher(name, rawval)
		if err != nil {
			err = fmt.Errorf("couldn't set new value for %s: %v", name, err)
			fail(w, err.Error(), http.StatusBadRequest)
			return
		}
		success(w, []byte(`{"Err":null}`))
		return
	}
}
