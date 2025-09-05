package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	_ "embed"
)

//go:embed templates/main-cli.go
var cliAppCode string

func main() {
	projectsDir := os.ExpandEnv("$HOME/projects/")

	name := flag.String("n", "my-project", "name of the project")
	kind := flag.String("k", "cli", "kind of the project. kinds: cli, api, app") // TODO: add bgworker, change app to webapp if possible
	// TODO: extras: add git init, options to choose libs / frameworks used.
	// like for cli - flag, cli/v2, cobra
	// api - chi, echo, fiber, gin, beego
	// app - this is where it gets complicated
	// 	be: any of the above api servers
	// 	fe: htmx+templ, vue, alpine, svelte
	// 	ui: tailwind, daisyui, bulma, pico, shadcn/ui
	flag.Parse()

	fmt.Printf("name: %+v, kind: %+v\n", *name, *kind)

	projectPath := path.Clean(path.Join(projectsDir, *name))

	// TODO: Add rollbacks for entire project creation
	// 1. Check for existence
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Println("sorry name already exists, come up with a better name already!")
		os.Exit(1)
	}

	// 2. Create directory
	err := os.Mkdir(projectPath, 0755)
	if err != nil {
		fmt.Println("error creating dir", err)
		os.Exit(1)
	}
	fmt.Println("created project dir")

	// 3. Initialize go module
	// TODO: ask if use local path or scm path
	cmd := exec.Command("go", "mod", "init", *name)
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		fmt.Println("error initializing project", err)
		os.Exit(1)
	}
	fmt.Println("initialized go project")

	// 4. Based on kind, add deps and copy templates
	switch *kind {
	case "cli":
		fmt.Println("forgin' cli project")
	case "api":
		fmt.Println("forgin' api project")
	case "app":
		fmt.Println("forgin' app project")
	}

	fmt.Printf("done. created project `%s` of kind `%s`\n", *name, *kind)
}
