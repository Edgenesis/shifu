package main

import (
	"os"

	"github.com/edgenesis/shifu/tools/release/gpt"
)

func main() {
	args := os.Args

	if len(args) < 2 {
		panic("no response body")
	}

	err := gpt.Start(args[1])
	if err != nil {
		panic(err.Error())
	}
}
