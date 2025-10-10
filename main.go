package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

// ANSI Color Constants for modern, structured CLI output
const (
	ColorReset     = "\033[0m"
	ColorRed       = "\033[31m"
	ColorGreen     = "\033[32m"
	ColorYellow    = "\033[33m"
	ColorBlue      = "\033[34m"
	ColorCyan      = "\033[36m"
	ColorWhiteBold = "\033[1;37m"
	StyleBold      = "\033[1m"
	StyleItalic    = "\033[3m"
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

var projectsDir string

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
			fmt.Printf("\n"+ColorRed+StyleBold+"[ ❌ FATAL ] "+ColorReset+"Cleaning up directory: %s\n", projectPath)
			os.RemoveAll(projectPath)
		}
	}()

	// Use the project name for the module path if it's not provided.
	if modulePath == "" {
		modulePath = name
	}

	// Modern, styled check for existence and confirmation.
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Printf(ColorYellow+StyleBold+"\n[ ⚠️  Warning ]"+ColorReset+" Directory exists: %s\n", projectPath)
		fmt.Printf(ColorCyan + "         > " + ColorReset + "Overwrite contents? " + StyleBold + "(y/N): " + ColorReset)

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(response)) != "y" {
			cleanupOnFailure = false
			return cli.Exit(ColorYellow+"Operation aborted by user."+ColorReset, 1)
		}
	}

	fmt.Println("\n" + StyleBold + "--- Starting Forge Project Generation ---" + ColorReset)

	// 1. Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return cli.Exit(fmt.Sprintf(ColorRed+"Error creating directory: %v"+ColorReset, err), 1)
	}
	fmt.Println(ColorGreen + "[ ✔️ ] " + ColorReset + "Directory created: " + StyleItalic + projectPath + ColorReset)

	// 2. Initialize Go module
	if err := runCommand(projectPath, "go", "mod", "init", modulePath); err != nil {
		return cli.Exit(fmt.Sprintf(ColorRed+"Error initializing Go module: %v"+ColorReset, err), 1)
	}
	fmt.Println(ColorGreen + "[ ✔️ ] " + ColorReset + "Initialized Go module: " + ColorBlue + modulePath + ColorReset)

	// 3. Initialize Git repository
	if git {
		if err := runCommand(projectPath, "git", "init"); err != nil {
			return cli.Exit(fmt.Sprintf(ColorRed+"Error initializing Git repository: %v"+ColorReset, err), 1)
		}
		fmt.Println(ColorGreen + "[ ✔️ ] " + ColorReset + "Initialized Git repository")
	}

	// 4. Add dependencies and write template file
	if template, ok := templates[kind]; ok {
		// Check for specific CLI library dependencies.
		if kind == "cli" && cliLib != "flag" {
			if deps, ok := dependencies[cliLib]; ok {
				fmt.Printf(ColorBlue+"[ 📦 ] "+ColorReset+"Adding dependencies for " + StyleBold + cliLib + ColorReset + "\n")
				if err := runCommand(projectPath, "go", append([]string{"get"}, deps...)...); err != nil {
					return cli.Exit(fmt.Sprintf(ColorRed+"Error getting dependencies: %v"+ColorReset, err), 1)
				}
				fmt.Println(ColorGreen + "[ ✔️ ] " + ColorReset + "Dependencies successfully installed")
			} else {
				return cli.Exit(fmt.Sprintf(ColorRed+"Unknown CLI library: %s"+ColorReset, cliLib), 1)
			}
		}

		mainGoPath := path.Join(projectPath, "main.go")
		if err := os.WriteFile(mainGoPath, []byte(template), 0644); err != nil {
			return cli.Exit(fmt.Sprintf(ColorRed+"Error writing main.go: %v"+ColorReset, err), 1)
		}
		fmt.Println(ColorGreen + "[ ✔️ ] " + ColorReset + "Wrote initial main.go file")
	} else {
		return cli.Exit(fmt.Sprintf(ColorRed+"Unknown project kind: %s"+ColorReset, kind), 1)
	}

	// Final success block (modern, structured output)
	fmt.Printf("\n"+ColorGreen+StyleBold+"[ ✨ SUCCESS ] Project '%s' is ready!\n"+ColorReset, name)
	fmt.Printf(ColorWhiteBold+"  Kind: %-10s Module: %s\n\n"+ColorReset, kind, modulePath)

	cleanupOnFailure = false
	return nil
}

// detectProjectType checks directory contents to determine the project's language/type.
func detectProjectType(contents []os.DirEntry) (projectType string, isGitRepo bool) {
	projectType = "unknown"
	isGitRepo = false

	// Define priority list of files to check for project type
	typeCheck := map[string]string{
		"go.mod":           "go",
		"Cargo.toml":       "rust",
		"package.json":     "node",
		"requirements.txt": "python",
		"mix.exs":          "elixir",
		"build.zig":        "zig",
	}

	for _, c := range contents {
		name := c.Name()

		// Check for Git repository
		if name == ".git" && c.IsDir() {
			isGitRepo = true
		}

		// Check for project type (only set if currently unknown)
		if projectType == "unknown" && !c.IsDir() {
			if t, ok := typeCheck[name]; ok {
				projectType = t
			}
		}

		if projectType != "unknown" && isGitRepo {
			break
		}
	}

	// Handle Python secondary check
	if projectType == "unknown" {
		for _, c := range contents {
			if c.Name() == ".venv" && c.IsDir() {
				projectType = "python"
				break
			}
		}
	}

	return projectType, isGitRepo
}

// listProjects handles the 'forge list' command.
func listProjects(c *cli.Context) error {
	// 1. Read the projects directory
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return cli.Exit(fmt.Sprintf(ColorRed+"Error reading projects directory: %v"+ColorReset, err), 1)
	}

	var projectsCount int
	const sepLine = "------------------------------------------------------"

	// Modern, colorized header block
	fmt.Printf(ColorCyan+StyleBold+"\n--- PROJECTS IN %s ---\n"+ColorReset, strings.ToUpper(projectsDir))
	fmt.Println(ColorBlue + sepLine + ColorReset)
	fmt.Printf(ColorBlue + StyleBold + "| %-20s | %-10s | %-5s |\n" + ColorReset, "PROJECT NAME", "TYPE", "GIT")
	fmt.Println(ColorBlue + sepLine + ColorReset)


	// 2. Iterate and process each directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip files
		}

		projectName := entry.Name()
		projectPath := path.Join(projectsDir, projectName)

		// Read directory contents to check for project type and git status
		contents, err := os.ReadDir(projectPath)
		if err != nil {
			log.Printf(ColorYellow+"[ ⚠️ ] Warning: Could not read contents of directory %s: %v"+ColorReset, projectName, err)
			continue
		}

		// Use the detection function
		projectType, isGitRepo := detectProjectType(contents)

		projectsCount++

		// Colorize status symbols
		gitRepoStatus := ColorRed + "❌" + ColorReset
		if isGitRepo {
			gitRepoStatus = ColorGreen + "✔️" + ColorReset
		}

		// Colorize type names
		typeColor := ColorReset
		switch projectType {
		case "go":
			typeColor = ColorBlue
		case "rust":
			typeColor = ColorRed
		case "node":
			typeColor = ColorGreen
		case "python":
			typeColor = ColorYellow
		default:
			typeColor = ColorReset
		}

		// 3. Print the results using fixed-width alignment and embedded colors
		fmt.Printf("| "+StyleBold+"%-20s "+ColorReset+"| %s%-10s%s | %s%-5s %s|\n",
			projectName,
			typeColor, projectType, ColorReset,
			gitRepoStatus, "", ColorReset)
	}

	// Clean footer
	fmt.Println(ColorBlue + sepLine + ColorReset)
	fmt.Printf(StyleBold+"Total projects found: %d\n\n"+ColorReset, projectsCount)

	return nil
}

func main() {
	log.SetFlags(0)

	projectsDir = os.Getenv("PROJECTS_DIR")
	if projectsDir == "" {
		// use default projects dir
		projectsDir = os.ExpandEnv("$HOME/projects/")
	}

	app := &cli.App{
		Name:    "forge",
		Usage:   "The ultimate project launcher. Fast. Minimal. Dope.",
		Version: "0.1.4",
		Commands: []*cli.Command{
			{
				Name:      "new",
				Usage:     "Creates a fresh project in $HOME/projects",
				ArgsUsage: "<project-name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "kind",
						Aliases: []string{"k"},
						Usage:   "Project flavor: cli, api, or app",
						Value:   "cli",
					},
					&cli.StringFlag{
						Name:    "module-path",
						Aliases: []string{"m"},
						Usage:   "Go module path (e.g., github.com/user/myproject)",
					},
					&cli.BoolFlag{
						Name:    "git",
						Aliases: []string{"g"},
						Usage:   "Initialize a git repository 🤘",
					},
					&cli.StringFlag{
						Name:    "cli-lib",
						Aliases: []string{"cl"},
						Usage:   "CLI framework to use. Options: flag, cobra, cli",
						Value:   "flag",
					},
				},
				Action: createProject,
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "See what's cookin' in $HOME/projects",
				Action:  listProjects,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

