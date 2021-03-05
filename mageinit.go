// +build mage

package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

func Init() {
	var must = func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	// look for NEWMODULE in environment
	newmod := os.Getenv("NEWMODULE")
	if newmod == "" {
		must(fmt.Errorf("Init requires NEWMODULE to be defined in the environment"))
	}
	// Compose the source module name string (so we don't overwrite it in this func)
	srcmod := "github.com/" + "Michael-F-Ellis/" + "wasmskel"
	fmt.Printf("Source Module is %s\n", srcmod)

	// Walk the tree and change all the instances of srcmod
	var walker filepath.WalkFunc = func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			fmt.Printf("Processing %s\n", info.Name())
			read, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			newContents := strings.Replace(string(read), srcmod, newmod, -1)
			err = ioutil.WriteFile(path, []byte(newContents), 0)
			if err != nil {
				return err
			}
		}
		return nil
	}
	// Make the changes
	must(filepath.Walk(".", walker))
	// Edit go.mod
	must(sh.Run("go", "mod", "edit", "--module", newmod))
	must(sh.Run("go", "mod", "tidy"))

}
