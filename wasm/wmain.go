// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"syscall/js"
	"time"

	"github.com/Michael-F-Ellis/wasmskel/internal/common"
)

// global copy of state obtained periodically from server
var State = common.State{}
var SP = &State // a pointer to refer to the global state

// A few predefined errors
// An Http error that actually indicates success
var Http200Error = fmt.Errorf("%d OK", http.StatusOK)

// NoDocumentError is returned if the global document is not available
var NoDocumentError = errors.New("unable to get document object")

// Main exports a setter function that can be called from javascript.
// Then it launches the Server Interface as a goroutine and finally
// waits forever on an empty select.
func main() {
	fmt.Println("Go Web Assembly") // fmt.Print outputs go to the js console.
	js.Global().Set("Setter", SetterWrapper())
	go ServerInterface()
	select {}
}

// SetterWrapper exports a function that allows javascript to enqueue json
// messages to be sent to the server as /set commands.
func SetterWrapper() (jsf js.Func) {
	jsf = js.FuncOf(
		func(this js.Value, args []js.Value) (result interface{}) {
			if len(args) != 1 {
				result := map[string]interface{}{
					"error": "Invalid no of arguments passed",
				}
				return result
			}
			setterChan <- []byte(args[0].String())
			return
		},
	)
	return
}

// getElementById is a wasm-side call to get a js Value by its id
func getElementById(id string) (el js.Value, err error) {
	jsDoc := js.Global().Get("document")
	if !jsDoc.Truthy() {
		err = NoDocumentError
		return
	}
	el = jsDoc.Call("getElementById", id)
	if !el.Truthy() {
		err = fmt.Errorf("Unable to get element with id %s", id)
	}
	return
}

// setElementAttributeById assigns a string value to a DOM element
// with the given id.
func setElementAttributeById(id, attr, value string) (err error) {
	el, err := getElementById(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	el.Set(attr, value)
	return
}

// setterChan accepts json byte strings to be sent to the server.
var setterChan = make(chan []byte, 1)

// ServerInterface listens on setterChan for changes to post to the server. When
// no messages are available on the channel, it fetches State from the server
// once per second. It must be invoked as a goroutine.
func ServerInterface() {
	for {
		var err error
		select {
		case jsonData := <-setterChan:
			err = SetFloat(jsonData, "/set", 2)
			_ = setElementAttributeById("SetMsg", "textContent", err.Error())
			if err != nil {
				fmt.Println(err)
			}
		case <-time.After(time.Second):
		}
		// in either case update the state
		_, err = getStateFromServer() // always returns an error even on success so we can update status line
		_ = setElementAttributeById("GetMsg", "textContent", err.Error())
		if err != Http200Error {
			fmt.Println(err)
			continue
		}
		// Write new values to readouts in web page
		UpdateParmReadouts()
	}
}

// getStateFromServer fetches the current values in State from the server and
// updates a local copy. It also returns the JSON byte slice that came from
// the server.
func getStateFromServer() (jbytes []byte, err error) {
	// Compose the request
	req, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Send the request
	client := &http.Client{}
	client.Timeout = 500 * time.Millisecond
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// Read and check status of the response
	jbytes, _ = ioutil.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		err = Http200Error // actually success
	default:
		err = fmt.Errorf("%s: %s", resp.Status, string(jbytes))
		fmt.Printf("%v", err) // also log it to the console
	}

	// Decode the response and update the global state
	mp := &common.State{}
	e := json.Unmarshal(jbytes, mp)
	if e != nil {
		err = fmt.Errorf("%v", e)
		fmt.Println(err)
		return
	}
	SP.DirectUpdate(func(p *common.State) { *p = *mp })
	return
}

// SetFloat posts a /set request to the server to change the value of a float
// parameter. It always returns an error, which will be Http200Error when the
// request is successful.
func SetFloat(jsonData []byte, url string, timeout int64) (err error) {
	// compose the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	//fmt.Println("Request Header:", req.Header)
	fmt.Println("Request Body:", req.Body)

	// Send the request
	client := &http.Client{}
	client.Timeout = time.Duration(timeout * 1e9) // nanoseconds, hence the 1e9 to get seconds
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%v", err) // also log it to the console
		return
	}
	defer resp.Body.Close()

	// Read and check the response.
	body, _ := ioutil.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		err = Http200Error // actually success
	default:
		err = fmt.Errorf("%s: %s: %s", resp.Status, string(jsonData), string(body))
		fmt.Printf("%v", err) // also log it to the console
	}
	return
}
