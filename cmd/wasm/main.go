// +build js,wasm
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"syscall/js"
	"time"
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
		jsDoc := js.Global().Get("document")
		if !jsDoc.Truthy() {
			result := map[string]interface{}{
				"error": "Unable to get document object",
			}
			return result
		}
		jsonOuputTextArea := jsDoc.Call("getElementById", "jsonoutput")
		if !jsonOuputTextArea.Truthy() {
			result := map[string]interface{}{
				"error": "Unable to get output text area",
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
	mp := &MonitoredParameters{}
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
