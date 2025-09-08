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
var cliCode string
//go:embed templates/main-api.go
var apiCode string
//go:embed templates/main-app.go
var appCode string

func main() {
	projectsDir := os.ExpandEnv("$HOME/projects/")

	name := flag.String("n", "", "name of the project")
	kind := flag.String("k", "cli", "kind of the project. kinds: cli, api, app") // TODO: add bgworker, change app to webapp if possible
	git := flag.Bool("g", false, "initialize a git repository")
	// TODO: extras: options to choose libs / frameworks used.
	// like for cli - flag, cli/v2, cobra
	// api - chi, echo, fiber, gin, beego
	// app - this is where it gets complicated
	// 	be: any of the above api servers
	// 	fe: htmx+templ, vue, alpine, svelte
	// 	ui: tailwind, daisyui, bulma, pico, shadcn/ui
	flag.Parse()

	if *name == "" {
		fmt.Println("please provide a name for project with -n")
		os.Exit(1)
	}

	fmt.Printf("name: %+v, kind: %+v\n", *name, *kind)

	projectPath := path.Clean(path.Join(projectsDir, *name))

	// TODO: Add rollbacks for entire project creation,
	// also even if directory exists, proceed if user wants to continue
	// 1. Check for existence
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Println("sorry name already exists, come up with a better name already!")
		os.Exit(1)
	}

	// 2. Create directory
	err := os.Mkdir(projectPath, 0755)
	if err != nil {
		fmt.Println("error creating directory:", err)
		os.Exit(1)
	}
	fmt.Println("[✓] created project directory")

	if *git {
		fmt.Println("-g flag passed. initializing git repository")
		// 2.5. Initialize git repository
		gitInitCmd := exec.Command("git", "init")
		gitInitCmd.Dir = projectPath
		if err := gitInitCmd.Run(); err != nil {
			fmt.Println("error initializing project:", err)
			os.Exit(1)
		}
		fmt.Println("[✓] initialized git repository")
	}

	// 3. Initialize go module
	// TODO: ask if use local path or scm path
	cmd := exec.Command("go", "mod", "init", *name)
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		fmt.Println("error initializing project", err)
		os.Exit(1)
	}
	fmt.Println("[✓] initialized go project")

	// 4. Based on kind, add deps and copy templates
	switch *kind {
	case "cli":
		fmt.Println("forgin' cli project")
		mainGoPath := path.Join(projectPath, "main.go")
		fmt.Println("writing to", mainGoPath)
		err := os.WriteFile(mainGoPath, []byte(cliCode), 0644)
		if err != nil {
			fmt.Println("error writing to main.go:", err)
			os.Exit(1)
		}
		fmt.Println("[✓] successfully written to", mainGoPath)
	case "api":
		fmt.Println("forgin' api project")
		mainGoPath := path.Join(projectPath, "main.go")
		fmt.Println("writing to", mainGoPath)
		err := os.WriteFile(mainGoPath, []byte(apiCode), 0644)
		if err != nil {
			fmt.Println("error writing to main.go:", err)
			os.Exit(1)
		}
		fmt.Println("[✓] successfully written to", mainGoPath)
	case "app":
		fmt.Println("forgin' app project")
		mainGoPath := path.Join(projectPath, "main.go")
		fmt.Println("writing to", mainGoPath)
		err := os.WriteFile(mainGoPath, []byte(appCode), 0644)
		if err != nil {
			fmt.Println("error writing to main.go:", err)
			os.Exit(1)
		}
		fmt.Println("[✓] successfully written to", mainGoPath)
	default:
		fmt.Println("unknown project kind")
		os.Exit(1)
	}

	fmt.Printf("[✓] done. created project `%s` of kind `%s`\n", *name, *kind)
}
