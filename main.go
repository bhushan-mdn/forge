package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	_ "embed"

	"github.com/urfave/cli/v2"
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

// runCommand is a helper function to run an external command.
func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// createProject contains all the logic for creating a new project.
func createProject(c *cli.Context) error {
	projectsDir := os.ExpandEnv("$HOME/projects/")
	name := c.Args().First()
	kind := c.String("kind")
	modulePath := c.String("module-path")
	cliLib := c.String("cli-lib")
	git := c.Bool("git")

	projectPath := path.Clean(path.Join(projectsDir, name))

	// Set a cleanup flag that will be checked on defer
	cleanupOnFailure := true
	defer func() {
		if cleanupOnFailure {
			log.Printf("An error occurred. Cleaning up directory: %s", projectPath)
			os.RemoveAll(projectPath)
		}
	}()

	// Use the project name for the module path if it's not provided.
	if modulePath == "" {
		modulePath = name
	}

	// Check if directory exists and ask for confirmation.
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Printf("Directory %s already exists. Do you want to continue? (y/n): ", projectPath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(response) != "y" {
			cleanupOnFailure = false
			return cli.Exit("Operation aborted.", 1)
		}
	}

	// 1. Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return cli.Exit(fmt.Sprintf("Error creating directory: %v", err), 1)
	}
	fmt.Println("[✓] Created project directory")

	// 2. Initialize Go module
	if err := runCommand(projectPath, "go", "mod", "init", modulePath); err != nil {
		return cli.Exit(fmt.Sprintf("Error initializing Go module: %v", err), 1)
	}
	fmt.Println("[✓] Initialized Go module")

	// 3. Initialize Git repository
	if git {
		if err := runCommand(projectPath, "git", "init"); err != nil {
			return cli.Exit(fmt.Sprintf("Error initializing Git repository: %v", err), 1)
		}
		fmt.Println("[✓] Initialized Git repository")
	}

	// 4. Add dependencies and write template file
	if template, ok := templates[kind]; ok {
		// Check for specific CLI library dependencies.
		if kind == "cli" && cliLib != "flag" {
			if deps, ok := dependencies[cliLib]; ok {
				fmt.Printf("Adding dependencies for %s\n", cliLib)
				if err := runCommand(projectPath, "go", append([]string{"get"}, deps...)...); err != nil {
					return cli.Exit(fmt.Sprintf("Error getting dependencies: %v", err), 1)
				}
				fmt.Println("[✓] Dependencies added")
			} else {
				return cli.Exit(fmt.Sprintf("Unknown CLI library: %s", cliLib), 1)
			}
		}

		mainGoPath := path.Join(projectPath, "main.go")
		if err := os.WriteFile(mainGoPath, []byte(template), 0644); err != nil {
			return cli.Exit(fmt.Sprintf("Error writing main.go: %v", err), 1)
		}
		fmt.Println("[✓] Wrote main.go file")
	} else {
		return cli.Exit(fmt.Sprintf("Unknown project kind: %s", kind), 1)
	}

	fmt.Printf("[✓] Done. Created project '%s' of kind '%s' in '%s'\n", name, kind, projectPath)
	cleanupOnFailure = false // Set cleanup to false on successful completion
	return nil
}

func main() {
	log.SetFlags(0)

	app := &cli.App{
		Name:    "forge",
		Usage:   "A simple Go project initializer",
		Version: "0.1.0",
		Commands: []*cli.Command{
			{
				Name:      "new",
				Usage:     "Creates a new project",
				ArgsUsage: "<project-name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "kind",
						Aliases: []string{"k"},
						Usage:   "Kind of the project. kinds: cli, api, app",
						Value: "cli",
					},
					&cli.StringFlag{
						Name:    "module-path",
						Aliases: []string{"m"},
						Usage:   "Go module path (e.g., github.com/user/myproject)",
					},
					&cli.BoolFlag{
						Name:    "git",
						Aliases: []string{"g"},
						Usage:   "Initialize a git repository",
					},
					&cli.StringFlag{
						Name:    "cli-lib",
						Aliases: []string{"cl"},
						Usage:   "CLI framework to use. Options: flag, cobra, cli",
						Value: "flag",
					},
				},
				Action: createProject,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

