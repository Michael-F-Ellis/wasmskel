package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Michael-F-Ellis/wasmskel/internal/common"
)

// This is the global state that is shared via a JSON API.
var State = common.State{}

// The assets directory contains static files served by the application.
//go:embed assets
var assets embed.FS

// Main launches a goroutine that continually updates the global state. Then it
// defines the allowed http requests and enters a ListenAndServe loop.
func main() {
	go Updater()
	mux := http.NewServeMux()
	// fs.Sub returns a file system rooted under our embedded assets directory
	// so that a request for, say, "/app.wasm" returns the file in "assets/app.wasm"
	assetSys, err := fs.Sub(assets, "assets")
	if err != nil {
		panic("failed to create sub-tree of assets") // should never happen
	}
	// The "/get" and "/set" urls are dynamic, i.e. they return the results
	// from computation rather than static files.
	mux.HandleFunc("/get", getRequestHandler)
	mux.HandleFunc("/set", setRequestHandler)
	// The following creates a handler for static file requests.
	mux.Handle("/", http.FileServer(http.FS(assetSys)))
	// Launch the http service
	log.Fatal(http.ListenAndServe(":9090", mux))
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
// State that are part of the JSON API.
func GetJSON(sp *common.State) (jsn []byte, err error) {
	mpcopy := sp.Get()
	jsn, err = json.Marshal(mpcopy)
	return
}

// Updater continually changes State, simulating
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
