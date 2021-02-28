// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

// Project directory tree. Values populated initPaths()
var (
	MageRoot     string // location of this file
	GoRoot       string // path to go installation
	AssetsPath   string // assets subdir
	InternalPath string // cmd/internal subdir
	CommonPath   string // common subdir
	ServerPath   string // server subdir
	WasmPath     string // wasm subdir
)

// initPaths populates the global path variables that define the project tree
func initPaths() {
	must := func(_err error) {
		if _err != nil {
			log.Fatal(_err)
		}
	}
	var err error
	GoRoot, err = sh.Output("go", "env", "GOROOT")
	must(err)
	MageRoot, err = os.Getwd()
	must(err)
	fmt.Println(MageRoot)
	AssetsPath = path.Join(MageRoot, "server", "assets")
	InternalPath = path.Join(MageRoot, "internal")
	CommonPath = path.Join(InternalPath, "common")
	ServerPath = path.Join(MageRoot, "server")
	WasmPath = path.Join(MageRoot, "wasm")
}

func Build() {
	initPaths()
	must := func(_err error) {
		if _err != nil {
			log.Fatal(_err)
		}
	}
	defer os.Chdir(MageRoot)
	// Generate the common state struct
	must(mkState())
	// Generate the web page
	must(IndexPage())
	// Install fresh copy of wasm_exec.js from go installation
	must(sh.Run("cp", fmt.Sprintf("%s/misc/wasm/wasm_exec.js", GoRoot), AssetsPath))
	// Build and install the WASM
	must(mkUpdater())
	must(os.Chdir(WasmPath))
	must(sh.Run("go", "fmt", "updater_g.go"))
	must(sh.Run("env", "GOOS=js", "GOARCH=wasm", "go", "build", "-o", path.Join(AssetsPath, "app.wasm")))
	// Build and install the server
	must(mkDispatcher())
	must(os.Chdir(ServerPath))
	must(sh.Run("go", "fmt", "dispatch_g.go"))
	must(sh.Run("go", "build", "-o", path.Join(MageRoot, "serve")))
}

func Run() {
	mg.Deps(Build)
	// launch the server
	sh.Run(path.Join(MageRoot, "serve"))
}

func Clean() {
	initPaths()
	must := func(_err error) {
		if _err != nil {
			log.Fatal(_err)
		}
	}
	must(os.Remove(path.Join(MageRoot, "serve")))
	must(os.Remove(path.Join(AssetsPath, "app.wasm")))
	must(os.Remove(path.Join(AssetsPath, "wasm_exec.js")))
	must(os.Remove(path.Join(AssetsPath, "index.html")))
}
