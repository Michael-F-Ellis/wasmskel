
	// +build js,wasm
	// Automatically generated. Do not edit.

	package main

	import "fmt"

	func updateStateTable(){
	var err error
	
	err = setElementAttributeById("Alpha", "textContent", fmt.Sprintf("%0.2f", SP.Alpha))
	if err != nil {
		fmt.Println(err)
	}
	err = setElementAttributeById("Beta", "textContent", fmt.Sprintf("%0.2f", SP.Beta))
	if err != nil {
		fmt.Println(err)
	}
	return
	}
	