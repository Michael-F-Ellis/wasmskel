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
	el = jsDoc.Call("getElementById", "jsonoutput")
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
	mp := &common.MonitoredParameters{}
	for {
		var err error
		time.Sleep(time.Second)
		req, err := http.NewRequest("GET", "/get", nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		client := &http.Client{}
		client.Timeout = 500 * time.Millisecond
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer resp.Body.Close()
		jsDoc := js.Global().Get("document")
		if !jsDoc.Truthy() {
			err = fmt.Errorf("unable to get document object")
			fmt.Println(err)
			continue
		}
		jsonInputTextArea := jsDoc.Call("getElementById", "jsoninput")
		if !jsonInputTextArea.Truthy() {
			err = fmt.Errorf("unable to get jsoninput text area")
			fmt.Println(err)
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, mp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// jsonInputTextArea.Set("value", fmt.Sprintf("%v", *mp))
		jsonInputTextArea.Set("value", string(body))

	}
}
