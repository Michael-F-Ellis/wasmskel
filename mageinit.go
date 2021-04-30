// +build mage

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/magefile/mage/sh"
)

// Init initializes a clone of the repository
func Init() {
	var must = func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	modname, remotename, err := getModNameAndRemoteOrigin()
	must(err)
	// Already initialized if module and remote names are the same
	if modname == remotename {
		return
	}
	// Otherwise ask user to confirm initialization
	if !prompter.YN(fmt.Sprintf(`Update go module from "%s" to "%s"?`, modname, remotename), true) {
		log.Fatal(errors.New("Init cancelled."))
	}
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
			newContents := strings.Replace(string(read), modname, remotename, -1)
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
	must(sh.Run("go", "mod", "edit", "--module", remotename))
	must(sh.Run("go", "mod", "tidy"))

}

// getModNameAndRemoteOrigin returns the module name and the remote origin
func getModNameAndRemoteOrigin() (modname, remotename string, err error) {
	// read srcmod from go.mod
	modname, err = sh.Output("go", "list", "-m")
	if err != nil {
		err = fmt.Errorf("failed to read module name from go.mod: %v", err)
		return
	}
	// Read remote origin URL with git
	remoteOriginUrl, err := sh.Output("git", "remote", "get-url", "origin")
	if err != nil {
		err = fmt.Errorf("failed to read remote origin: %v", err)
		return
	}
	// need to strip protocol, e.g. "https://" and ".git" suffix from URL so
	// that "https://github.com/SomeOne/somerepo.git" becomes
	// "github.com/SomeOne/somerepo" which is the proper form for a go module
	// reference.
	splitUrl := strings.Split(remoteOriginUrl, "://")
	noProtocol := splitUrl[len(splitUrl)-1]
	remotename = strings.TrimSuffix(noProtocol, ".git")
	return
}
