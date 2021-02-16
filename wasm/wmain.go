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
var SP = &State

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("Setter", SetterWrapper())
	go getter()
	select {}
}

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

/*
// jsonWrapper wraps prettyJson so it can be calld from javascript
func jsonWrapper() js.Func {
	jsonfunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			result := map[string]interface{}{
				"error": "Invalid no of arguments passed",
			}
			return result
		}
		jsonOuputTextArea, err := getElementById("jsonoutput")
		if err != nil {
			result := map[string]interface{}{
				"error": err.Error(),
			}
			return result
		}
		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		pretty, err := prettyJson(inputJSON)
		if err != nil {
			errStr := fmt.Sprintf("unable to parse JSON. Error %s occurred\n", err)
			result := map[string]interface{}{
				"error": errStr,
			}
			return result
		}
		jsonOuputTextArea.Set("value", pretty)
		return nil
	})
	return jsonfunc
}
*/

// NoDocumentError is returned if the global document is not available
var NoDocumentError = errors.New("unable to get document object")

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
// with id.
func setElementAttributeById(id, attr, value string) (err error) {
	el, err := getElementById(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	el.Set(attr, value)

	return
}

/*
// prettyJson prints indented JSON
func prettyJson(input string) (string, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}
*/

var setterChan = make(chan []byte, 1)

// getter fetches State from the server once per second. It must be invoked as a
// goroutine.
func getter() {
	for {
		var err error
		select {
		case jsonData := <-setterChan:
			err = setFloat(jsonData, "/set", 2)
			if err != nil {
				fmt.Println(err)
			}
		case <-time.After(time.Second):
		}
		// in either case update the state
		_, err = getStateFromServer()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// Write new values to readouts in web page
		updateStateTable()
	}
}

// getStateFromServer fetches the current values in State from the server and
// updates a local copy. It also returns the JSON byte slice that came from
// the server.
func getStateFromServer() (jbytes []byte, err error) {
	req, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &http.Client{}
	client.Timeout = 500 * time.Millisecond
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	jbytes, _ = ioutil.ReadAll(resp.Body)
	mp := &common.State{}
	err = json.Unmarshal(jbytes, mp)
	if err != nil {
		fmt.Println(err)
		return
	}
	SP.DirectUpdate(func(p *common.State) { *p = *mp })
	return
}

func setFloat(jsonData []byte, url string, timeout int64) (err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	//fmt.Println("Request Header:", req.Header)
	fmt.Println("Request Body:", req.Body)
	client := &http.Client{}
	client.Timeout = time.Duration(timeout * 1e9) // nanoseconds, hence the 1e9 to get seconds
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return
}
