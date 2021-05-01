// +build mage

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"

	. "github.com/Michael-F-Ellis/goht" // dot import makes sense here
)

// genIndexPage generates assets/index.html
func genIndexPage() (err error) {
	var buf bytes.Buffer
	// <head>
	head := Head("",
		Title(``, "Wasm Skeleton Demo"),
		Meta(`name="viewport" content="width=device-width, initial-scale=1"`),
		Meta(`name="description", content="PGC Remote Interface"`),
		Link(`rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css"`),
		IndexCSS(),
		// indexJS(), // js for this page

		// Load the Go wasm interface library
		Script(`src="/wasm_exec.js" charset=UTF-8`),
		Script("", `
		// Load and launch our wasm component
		const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/app.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });`),
	)

	// Put the head and body together
	page := Html("",
		Null("\n<!-- Code generated by Mage. DO NOT EDIT -->"),
		head,
		IndexBody(),
	)

	// Render the html
	err = Render(page, &buf, 0)
	if err != nil {
		return
	}
	// Write the buffer to assets/index.html
	indexPath := path.Join(AssetsPath, "index.html")
	err = ioutil.WriteFile(indexPath, buf.Bytes(), 0644)
	return
}

// StatusTable returns a div containing a table element with 2 rows with 2 cells
// in each:
// | Get   | (latest status) |
// | Set   | (latest status) |
func StatusTable() (tbl *HtmlTree) {
	var rows []interface{}
	rows = append(rows, Tr(`class="STATUS"`, Td(`class=STATUS"`, "GET:"), Td(`class="STATUS" id="GetMsg"`)))
	rows = append(rows, Tr(`class="STATUS"`, Td(`class=STATUS"`, "SET:"), Td(`class="STATUS" id="SetMsg"`)))
	tbl = Div(``, H4(``, "HTTP Status Messages"), Table(`class="STATUS"`, rows...))
	return
}

// ParmTable returns a table element with rows for each parameter
// defined in MetaParms. Each row has 3 cells. For settable parameters,
// the first cell contains a "Set" button. For non-settable parameters it
// is empty.  The second cell contains the parameter name.  The third
// contains the latest value read from the server for that parameter.
func ParmTable() (tbl *HtmlTree) {
	var rows []interface{}
	for _, parm := range MetaParms {
		var btn *HtmlTree
		switch parm.Settable {
		case false:
			btn = Td(`class="PARM"`) // empty cell
		case true:
			onclick := fmt.Sprintf(`onclick='SetterPrompt("%s")'`, parm.Name)
			btn = Td(`class="PARM"`, Button(`class="PARM" `+onclick, "Set"))
		}
		readout := Td(fmt.Sprintf(`id="%s" class="PARM"`, parm.Name))
		label := Td(`class="PARM"`, parm.Name)
		rows = append(rows, Tr(`class="PARM"`, btn, label, readout))
	}
	tbl = Div(``, H4(``, "Parameter Values"), Table(`class="PARM"`, rows...))
	return
}

// SetterScript returns a script element that raises a window prompt
// when a user clicks one of the parameter "Set" buttons.
func SetterScript() (scrpt *HtmlTree) {
	scrpt = Script(``, `
		SetterPrompt = function (name) {
			var oldvalue = document.getElementById(name).innerText
    		var value = prompt("Enter new value for " +  name, oldvalue);
    		if (value != null) {
        		Setter('{"' + name + '":' + value + '}')
    		}
		}`)
	return
}

// IndexBody returns the body element for this page.
func IndexBody() (body *HtmlTree) {
	body = Body(``,
		H3(``, "Go Web Assembly Skeleton App"),
		StatusTable(),
		ParmTable(),
		SetterScript())
	return
}
