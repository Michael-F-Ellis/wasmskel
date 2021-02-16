// +build mage

package main

import (
	. "github.com/Michael-F-Ellis/goht" // dot import makes sense here
)

// IndexCSS defines CSS styling for index.html. In a real app, stylesheets may
// become large and intricate, hence the choice to put the generation in a
// separate mage file. Note that the text content here is straight CSS with no
// need for special quoting.
func IndexCSS() *HtmlTree {
	return Style("", `
	/* Status class styling */
	table.STATUS {
		margin-left: 5vh;
		font-size: small;
	}
	td.STATUS {
		margin-left: 1vh;
	}

	/* Parameters class styling */
	table.PARM {
		margin-left: 5vh;
		font-size: small;
	}
	td.PARM {
		margin-left: 1vh;
	}
	button.PARM {
		font-style: italic;
	}
	`)
}
