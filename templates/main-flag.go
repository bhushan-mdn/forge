package main

import (
	"flag"
	"fmt"
)

func main() {
	name := flag.String("name", "Go Developer", "The name of the user to greet.")
	flag.Parse()
	fmt.Printf("👋 Hello, %s! This is a simple flag-based CLI.\n", *name)
}
