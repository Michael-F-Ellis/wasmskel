// +build mage

package main

import (
	"os"
	"path"
	"text/template"
)

const (
	// Type name constants for code generation
	Float = "float64"
)

type Meta struct {
	Name     string
	Type     string
	Settable bool
}

var MetaParms = []Meta{
	{Name: "Alpha", Type: Float},
	{Name: "Beta", Type: Float},
	{Name: "Delta", Type: Float, Settable: true},
}

// mkState generates internal/common/state.go
func mkState() (err error) {
	tmpl := `
	// Automatically generated. Do not edit.

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
	dst, err := os.Create(path.Join(CommonPath, "state.go"))
	if err != nil {
		return
	}
	defer func() { dst.Close() }()

	err = t.Execute(dst, MetaParms)
	return
}

// mkUpdate generates wasm/updater.go
func mkUpdater() (err error) {
	tmpl := `
	// +build js,wasm
	// Automatically generated. Do not edit.

	package main

	import "fmt"

	func updateStateTable(){
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
	dst, err := os.Create(path.Join(WasmPath, "updater.go"))
	if err != nil {
		return
	}
	defer func() { dst.Close() }()

	err = t.Execute(dst, MetaParms)
	return
}

func mkDispatcher() (err error) {
	tmpl := `
	package main
	
	import (
		"fmt"
		"encoding/json"
		"github.com/Michael-F-Ellis/wasmskel/internal/common"
	)

	// UnsettableErr returns an err whose string value indicates an attempt to
    // set an unsettable variable
    func unsettableErr(varName string) error {
        return fmt.Errorf("%s is not settable", varName)
    }

	// dispatcher invokes the setter function for the requested jsonName
	func dispatcher(jsonName string, rawval *json.RawMessage) (err error) {
		switch jsonName {
		{{range .}}
		case "{{.Name}}":
		  {{- if not .Settable}}
			err = unsettableErr("{{.Name}}")
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
	dst, err := os.Create(path.Join(ServerPath, "dispatch.go"))
	if err != nil {
		return
	}
	defer func() { dst.Close() }()

	err = t.Execute(dst, MetaParms)
	return

}
