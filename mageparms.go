// +build mage

package main

import (
	"os"
	"path"
	"text/template"

	"github.com/magefile/mage/sh"
)

const (
	// Type name constants for code generation
	Float = "float64"
)

// Meta instances define each parameter the app supports.
type Meta struct {
	Name     string
	Type     string
	Settable bool
}

// MetaParms is a slice of Meta, one for each supported parameter.
var MetaParms = []Meta{
	{Name: "Alpha", Type: Float},
	{Name: "Beta", Type: Float},
	{Name: "Gamma", Type: Float, Settable: true},
	{Name: "Delta", Type: Float},
	{Name: "Zeta", Type: Float, Settable: true},
}

// genState generates common/state_g.go, the data definitions shared
// by the server and the web client.
func genState() (err error) {
	tmpl := `
	// Code generated by Mage. DO NOT EDIT.

	package common

	type State struct {
	{{range .}}
	    {{.Name}} {{.Type}}
		{{- end}}
	}
	`
	t, err := template.New("state").Parse(tmpl)
	if err != nil {
		return
	}
	fpath := path.Join(CommonPath, "state_g.go")
	dst, err := os.Create(fpath)
	if err != nil {
		return
	}
	defer func() { dst.Close() }()

	err = t.Execute(dst, MetaParms)
	if err != nil {
		return
	}
	err = sh.Run("go", "fmt", fpath)
	return
}

// mkUpdate generates wasm/updater.go
func genUpdater() (err error) {
	tmpl := `
	// +build js,wasm
	// Code generated by Mage. DO NOT EDIT.

	package main

	import "fmt"

	// UpdateParmReadouts copies the current values from the global state into
	// the corresponding cells in the Parameter Values table.
	func UpdateParmReadouts(){
	var err error
	{{range .}}
	err = setElementAttributeById("{{.Name}}", "textContent", fmt.Sprintf("%0.2f", SP.{{.Name}}))
	if err != nil {
		fmt.Println(err)
	}
	{{- end}}
	return
	}
	`
	t, err := template.New("updater").Parse(tmpl)
	if err != nil {
		return
	}
	fpath := path.Join(WasmPath, "updater_g.go")
	dst, err := os.Create(fpath)
	if err != nil {
		return
	}
	defer func() { dst.Close() }()

	err = t.Execute(dst, MetaParms)
	if err != nil {
		return
	}
	err = sh.Run("go", "fmt", fpath)
	return
}

func genDispatcher() (err error) {
	tmpl := `
	// Code generated by Mage. DO NOT EDIT.

	package main
	
	import (
		"fmt"
		"encoding/json"
		"github.com/Michael-F-Ellis/wasmskel/common"
	)

	// UnsettableErr returns an err whose string value indicates an attempt to
    // set an unsettable variable
    func UnsettableErr(varName string) error {
        return fmt.Errorf("%s is not settable", varName)
    }

	// Dispatcher invokes the setter function for the requested jsonName
	func Dispatcher(jsonName string, rawval *json.RawMessage) (err error) {
		switch jsonName {
		{{range .}}
		case "{{.Name}}":
		  {{- if not .Settable}}
			err = UnsettableErr("{{.Name}}")
		  {{- else if eq .Type "float64"}}
		    var value float64
			err = json.Unmarshal(*rawval, &value)
			if err != nil {
				err = fmt.Errorf("couldn't unmarshal value for {{.Name}}: %v", err)
				return
			}
			sp := &State
			sp.DirectUpdate(func(p *common.State) { p.{{.Name}}=value })
			{{- end}}
		{{- end}}
		}
		return
	}
	`
	t, err := template.New("dispatcher").Parse(tmpl)
	if err != nil {
		return
	}
	fpath := path.Join(ServerPath, "dispatch_g.go")
	dst, err := os.Create(fpath)
	if err != nil {
		return
	}
	defer func() { dst.Close() }()

	err = t.Execute(dst, MetaParms)
	if err != nil {
		return
	}
	err = sh.Run("go", "fmt", fpath)
	return

}
