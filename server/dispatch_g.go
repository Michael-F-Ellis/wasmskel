
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
		
		case "Alpha":
			err = unsettableErr("Alpha")
		case "Beta":
			err = unsettableErr("Beta")
		case "Gamma":
		    var value float64
			err = json.Unmarshal(*rawval, &value)
			if err != nil {
				err = fmt.Errorf("couldn't unmarshal value for Gamma: %v", err)
				return
			}
			sp := &State
			sp.DirectUpdate(func(p *common.State) { p.Gamma=value })
		case "Delta":
			err = unsettableErr("Delta")
		case "Zeta":
		    var value float64
			err = json.Unmarshal(*rawval, &value)
			if err != nil {
				err = fmt.Errorf("couldn't unmarshal value for Zeta: %v", err)
				return
			}
			sp := &State
			sp.DirectUpdate(func(p *common.State) { p.Zeta=value })
		}
		return
	}
	