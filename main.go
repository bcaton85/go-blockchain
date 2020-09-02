package main

import (
	blockchain "go-blockchain/src"
	"os"
)

func main() {

	// Could add flag parsing library, doing this for now
	var PORT string
	if len(os.Args) > 1 {
		PORT = os.Args[1]
	} else {
		PORT = "8000"
	}

	blockchain.Run(PORT)
}
