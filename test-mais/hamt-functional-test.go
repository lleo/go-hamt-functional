package main

import (
	"flag"
	"fmt"

	hamt "github.com/lleo/go-hamt-functional"
)

func main() {
	flag.Parse()

	var h = hamt.EMPTY

	fmt.Println("hello world")
	fmt.Println(h)
}
