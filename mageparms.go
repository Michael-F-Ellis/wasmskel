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
	Name string
	Type string
}

var MetaParms = []Meta{
	{Name: "Alpha", Type: Float},
	{Name: "Beta", Type: Float},
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
