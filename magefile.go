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
	MageRoot   string // location of this file
	GoRoot     string // path to go installation
	AssetsPath string // assets subdir
	CmdPath    string // cmd subdir
	ServerPath string // server subdir
	WasmPath   string // wasm subdir
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
	AssetsPath = path.Join(MageRoot, "assets")
	CmdPath = path.Join(MageRoot, "cmd")
	ServerPath = path.Join(CmdPath, "server")
	WasmPath = path.Join(CmdPath, "wasm")
}

func Build() {
	initPaths()
	must := func(_err error) {
		if _err != nil {
			log.Fatal(_err)
		}
	}
	// Install fresh copy of wasm_exec.js from go installation
	must(sh.Run("cp", fmt.Sprintf("%s/misc/wasm/wasm_exec.js", GoRoot), AssetsPath))
	// Build and install the WASM
	must(os.Chdir(WasmPath))
	must(sh.Run("env", "GOOS=js", "GOARCH=wasm", "go", "build", "-o", path.Join(AssetsPath, "json.wasm")))
	// Build and install the server
	must(os.Chdir(ServerPath))
	must(sh.Run("go", "build", "-o", path.Join(MageRoot, "serve")))
}

func Test() {
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
	must(os.Remove(path.Join(AssetsPath, "json.wasm")))
	must(os.Remove(path.Join(AssetsPath, "wasm_exec.js")))
}
