// +build js,wasm

package main

import (
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
var MonitoredParametersState = common.MonitoredParameters{}
var MP = &MonitoredParametersState

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("formatJSON", jsonWrapper())
	go getter()
	select {}
}

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

// getter fetches MonitoredParametersState from the server once per second
// and updates the jsonInputTextArea in the document.  It must be invoked
// as a goroutine.
func getter() {
	for {
		var err error
		time.Sleep(time.Second)
		jbytes, err := getMPFromServer()
		if err != nil {
			fmt.Println(err)
			continue
		}

		jsonInputTextArea, err := getElementById("jsoninput")
		if err != nil {
			fmt.Println(err)
			continue
		}
		jsonInputTextArea.Set("value", string(jbytes))

	}
}

// getMP fetches the current values in MonitoredParametersState from the server and
// updates a local copy. It also returns the JSON byte slice that came from
// the server.
func getMPFromServer() (jbytes []byte, err error) {
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
	mp := &common.MonitoredParameters{}
	err = json.Unmarshal(jbytes, mp)
	if err != nil {
		fmt.Println(err)
		return
	}
	MP.DirectUpdate(func(p *common.MonitoredParameters) { *MP = *p })
	return
}
