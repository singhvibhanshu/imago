// Command imago is a fully offline image toolkit: convert, compress and resize
// images entirely on your own machine. Nothing is ever uploaded.
package main

import (
	"os"

	"imago/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
