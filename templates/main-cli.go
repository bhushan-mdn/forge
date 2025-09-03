package main

import (
    "fmt"
    "flag"
)

func main() {
    name := flag.String("name", "there", "put your name here")
    flag.Parse()
    fmt.Printf("hello, %s!\n", *name)
}
