package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	_ "embed"
)

//go:embed templates/main-cli.go
var cliCode string

//go:embed templates/main-api.go
var apiCode string

//go:embed templates/main-app.go
var appCode string

// A map to hold the template code for different project kinds.
var templates = map[string]string{
	"cli": cliCode,
	"api": apiCode,
	"app": appCode,
}

// A map to hold the Go dependencies for different frameworks.
var dependencies = map[string][]string{
	"cobra": {"github.com/spf13/cobra"},
	"cli":   {"github.com/urfave/cli/v2"},
}

func main() {
	log.SetFlags(0)
	projectsDir := os.ExpandEnv("$HOME/projects/")

	name := flag.String("n", "", "name of the project")
	kind := flag.String("k", "cli", "kind of the project. kinds: cli, api, app")
	modulePath := flag.String("m", "", "Go module path (e.g., github.com/user/myproject)")
	cliLib := flag.String("cli-lib", "flag", "CLI framework to use. Options: flag, cobra, cli")
	git := flag.Bool("g", false, "initialize a git repository")

	flag.Parse()

	if *name == "" {
		log.Fatal("Please provide a project name with -n")
	}

	// Use the project name for the module path if it's not provided.
	if *modulePath == "" {
		*modulePath = *name
	}

	projectPath := path.Clean(path.Join(projectsDir, *name))

	// Check if directory exists and ask for confirmation.
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Printf("Directory %s already exists. Do you want to continue? (y/n): ", projectPath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(response) != "y" {
			log.Fatal("Operation aborted.")
		}
	}

	// 1. Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}
	fmt.Println("[✓] Created project directory")

	// 2. Initialize Go module
	cmd := exec.Command("go", "mod", "init", *modulePath)
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		// Clean up the created directory on failure.
		os.RemoveAll(projectPath)
		log.Fatalf("Error initializing Go module: %v", err)
	}
	fmt.Println("[✓] Initialized Go module")

	// 3. Initialize Git repository
	if *git {
		if err := runCommand(projectPath, "git", "init"); err != nil {
			os.RemoveAll(projectPath)
			log.Fatalf("Error initializing Git repository: %v", err)
		}
		fmt.Println("[✓] Initialized Git repository")
	}

	// 4. Add dependencies and write template file
	if template, ok := templates[*kind]; ok {
		// Check for specific CLI library dependencies.
		if *kind == "cli" && *cliLib != "flag" {
			if deps, ok := dependencies[*cliLib]; ok {
				fmt.Printf("Adding dependencies for %s\n", *cliLib)
				if err := runCommand(projectPath, "go", append([]string{"get"}, deps...)...); err != nil {
					os.RemoveAll(projectPath)
					log.Fatalf("Error getting dependencies: %v", err)
				}
				fmt.Println("[✓] Dependencies added")
			} else {
				log.Fatalf("Unknown CLI library: %s", *cliLib)
			}
		}

		mainGoPath := path.Join(projectPath, "main.go")
		if err := os.WriteFile(mainGoPath, []byte(template), 0644); err != nil {
			os.RemoveAll(projectPath)
			log.Fatalf("Error writing main.go: %v", err)
		}
		fmt.Println("[✓] Wrote main.go file")
	} else {
		os.RemoveAll(projectPath)
		log.Fatalf("Unknown project kind: %s", *kind)
	}

	fmt.Printf("[✓] Done. Created project '%s' of kind '%s' in '%s'\n", *name, *kind, projectPath)
}

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

