package main

import (
	"os"
	"text/template"
)

type Data struct {
	Imports  []string
	MainFunc string
}

func main() {
	t, err := template.ParseFiles("main.go.tmpl")
	if err != nil {
		panic(err)
	}
	data := Data{
		Imports: []string{"fmt"},
		MainFunc: `fmt.Println("hello friend")
	fmt.Println("hello friend")`,
	}
	err = t.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}
