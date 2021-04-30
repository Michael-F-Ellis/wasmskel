// +build mage

package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

// Project directory tree. Values populated initPaths()
var (
	MageRoot   string // location of this file
	GoRoot     string // path to go installation
	AssetsPath string // assets subdir
	CommonPath string // common subdir
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
	AssetsPath = path.Join(MageRoot, "server", "assets")
	CommonPath = path.Join(MageRoot, "common")
	ServerPath = path.Join(MageRoot, "server")
	WasmPath = path.Join(MageRoot, "wasm")
}

func Build() {
	mg.Deps(Init)
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

// Clean removes executables and  generated source files.
func Clean() {
	initPaths()
	check := func(_err error) {
		if _err != nil {
			log.Println(_err)
		}
	}
	check(os.Remove(path.Join(MageRoot, "serve")))
	check(os.Remove(path.Join(AssetsPath, "app.wasm")))
	check(os.Remove(path.Join(AssetsPath, "wasm_exec.js")))
	check(os.Remove(path.Join(AssetsPath, "index.html")))
	// By convention, names of generated Go files in this module end with
	// "_g.go". Walk the directory tree and remove them
	var walker filepath.WalkFunc = func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_g.go") {
			check(os.Remove(path))
		}
		return nil
	}
	// Take a walk ...
	check(filepath.Walk(".", walker))

}
