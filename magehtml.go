// +build mage

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"

	. "github.com/Michael-F-Ellis/goht" // dot import makes sense here
)

// mkIndexPage generates assets/index.html
func mkIndexPage() (err error) {
	var buf bytes.Buffer
	// <head>
	head := Head("",
		Meta(`name="viewport" content="width=device-width, initial-scale=1"`),
		Meta(`name="description", content="PGC Remote Interface"`),
		Link(`rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css"`),
		// indexCSS(),
		// indexJS(), // js for this page

		// Load the Go wasm interface library
		Script(`src="/wasm_exec.js" charset=UTF-8`),
		Script("", `
		// Load and launch our wasm component
		const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/json.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });`),
	)

	// Put the head and body together
	page := Html("", head, indexBody())

	// Generate the html
	err = Render(page, &buf, 0)
	if err != nil {
		return
	}
	// Write the buffer to assets/index.html
	indexPath := path.Join(AssetsPath, "index.html")
	err = ioutil.WriteFile(indexPath, buf.Bytes(), 0644)
	return
}
func indexBody() (body *HtmlTree) {
	var rows []interface{}
	for _, parm := range MetaParms {
		readout := Td(fmt.Sprintf(`id="%s" class="PARM"`, parm.Name))
		label := Td(`class="PARM"`, parm.Name)
		rows = append(rows, Tr(`class="PARM"`, label, readout))
	}
	body = Body(``, Table(`class="PARM"`, rows...))
	return
}
