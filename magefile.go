// +build mage

package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

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

// Generate creates source files that depend on MetaParms
func Generate() {
	mg.Deps(Init)
	initPaths()
	must := func(_err error) {
		if _err != nil {
			log.Fatal(_err)
		}
	}
	defer os.Chdir(MageRoot)
	// Generate the common state struct
	must(genState())
	// Generate the web page
	must(genIndexPage())
	// Install fresh copy of wasm_exec.js from go installation
	must(sh.Run("cp", fmt.Sprintf("%s/misc/wasm/wasm_exec.js", GoRoot), AssetsPath))
	// Generate the wasm client's updater function
	must(genUpdater())
	// Generate the server's dispatcher function
	must(genDispatcher())

}

// Build compiles the server and the Web Assembly client.
func Build() {
	mg.Deps(Generate)
	must := func(_err error) {
		if _err != nil {
			log.Fatal(_err)
		}
	}
	defer os.Chdir(MageRoot)
	// Build and install the WASM
	must(os.Chdir(WasmPath))
	must(sh.Run("env", "GOOS=js", "GOARCH=wasm", "go", "build", "-o", path.Join(AssetsPath, "app.wasm")))
	// Build and install the server
	must(os.Chdir(ServerPath))
	must(sh.Run("go", "build", "-o", path.Join(MageRoot, "serve")))
}

// Run builds and executes the server
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

	// Other generated files have names ending "_g.*"
	re := regexp.MustCompile(`_g\.\S+$`) // the pattern to match

	// Walk the directory tree and remove them
	var walker filepath.WalkFunc = func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() && ("" != re.FindString(info.Name())) {
			check(os.Remove(path))
		}
		return nil
	}
	// Take a walk ...
	check(filepath.Walk(".", walker))

}
