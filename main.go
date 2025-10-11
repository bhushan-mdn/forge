package main

import (
	"bufio"
	_ "embed" // Required for the embed directive
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

// --- SIMULATED LIPGLOSS STYLING CONSTANTS & HELPERS ---

const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorCyan    = "\033[36m"
	ColorMagenta = "\033[35m"
	StyleBold    = "\033[1m"
	StyleItalic  = "\033[3m"
	StyleFaint   = "\033[2m"
)

func Style(s string, styles ...string) string {
	return strings.Join(styles, "") + s + ColorReset
}

// --- TEMPLATES AND CONFIGURATION ---

// Template Files (Embedded for a fully featured application)
//
//go:embed templates/main-cli.go
var cliCode string

//go:embed templates/main-api.go
var apiCode string

//go:embed templates/main-app.go
var appCode string

//go:embed templates/main-flag.go
var flagCode string

//go:embed templates/GITIGNORE_GO
var goGitIgnore string // Added from github.com/github/gitignore

//go:embed templates/README.md
var readmeTemplate string // Added

var templates = map[string]string{
	"cli":  cliCode,
	"api":  apiCode,
	"app":  appCode,
	"flag": flagCode,
}

var dependencies = map[string][]string{
	"cobra": {"github.com/spf13/cobra"},
	"cli":   {"github.com/urfave/cli/v2"},
}

var projectsDir string

// --- UTILITY FUNCTIONS ---

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func confirmPrompt(prompt string) bool {
	fmt.Printf(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(response)) == "y"
}

// --- COMMAND IMPLEMENTATIONS (CREATE) ---

func createAndCleanup(projectPath, name, kind, modulePath, cliLib string, git bool) error {
	cleanupOnFailure := true
	defer func() {
		if cleanupOnFailure {
			fmt.Printf("\n%s\n", Style(" [ ❌ FAILURE ] Project creation failed. Cleaning up...", ColorRed, StyleBold))
			os.RemoveAll(projectPath)
		}
	}()

	// 1. Interactive confirmation prompt
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Printf(Style("\n [ ⚠️  Warning ] Directory exists: ", ColorYellow, StyleBold)+"%s\n", projectPath)
		if !confirmPrompt(Style("         > Overwrite contents? (", ColorCyan) + Style("Y", StyleBold) + Style("/n): ", ColorCyan)) {
			cleanupOnFailure = false
			return cli.Exit(Style("Operation aborted by user.", ColorYellow), 1)
		}
	}

	fmt.Println("\n" + Style("--- Generating Project: "+name+" ---", StyleBold, ColorMagenta))

	// 2. Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf(Style("error creating directory: %v", ColorRed), err)
	}
	fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Directory created:", StyleBold) + " " + Style(projectPath, StyleItalic))

	// 3. Initialize Go module
	if err := runCommand(projectPath, "go", "mod", "init", modulePath); err != nil {
		return fmt.Errorf(Style("error initializing Go module: %v", ColorRed), err)
	}
	fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Initialized module:", StyleBold) + " " + Style(modulePath, ColorBlue))

	// 4. Add dependencies, main.go, .gitignore, and README.md
	templateKey := kind
	if kind == "cli" && cliLib != "flag" {
		templateKey = "cli"
		if deps, ok := dependencies[cliLib]; ok {
			fmt.Printf(Style(" [ 📦 ] ", ColorBlue)+Style("Installing dependencies:", StyleBold)+" %s...\n", cliLib)
			if err := runCommand(projectPath, "go", append([]string{"get"}, deps...)...); err != nil {
				return fmt.Errorf(Style("error getting dependencies: %v", ColorRed), err)
			}
			fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Dependencies installed", StyleBold))
		} else {
			return fmt.Errorf(Style("unknown CLI library: %s", ColorRed), cliLib)
		}
	} else if kind == "cli" && cliLib == "flag" {
		templateKey = "flag"
	}

	template, ok := templates[templateKey]
	if !ok {
		return fmt.Errorf(Style("unknown project kind or template: %s", ColorRed), kind)
	}

	// Write main.go
	mainGoPath := path.Join(projectPath, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(template), 0644); err != nil {
		return fmt.Errorf(Style("error writing main.go: %v", ColorRed), err)
	}
	fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Wrote main.go template", StyleBold))

	// Write .gitignore
	gitIgnorePath := path.Join(projectPath, ".gitignore")
	if err := os.WriteFile(gitIgnorePath, []byte(goGitIgnore), 0644); err != nil {
		return fmt.Errorf(Style("error writing .gitignore: %v", ColorRed), err)
	}
	fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Wrote Go .gitignore", StyleBold))

	// Write README.md (replace placeholder)
	finalReadme := strings.ReplaceAll(readmeTemplate, "{{PROJECT_NAME}}", name)
	finalReadme = strings.ReplaceAll(finalReadme, "{{PROJECT_KIND}}", kind)
	readmePath := path.Join(projectPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(finalReadme), 0644); err != nil {
		return fmt.Errorf(Style("error writing README.md: %v", ColorRed), err)
	}
	fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Wrote README.md", StyleBold))

	// 5. Initialize Git repository
	if git {
		if err := runCommand(projectPath, "git", "init", "--quiet"); err != nil { // Use --quiet for less output
			return fmt.Errorf(Style("error initializing Git repository: %v", ColorRed), err)
		}
		if err := runCommand(projectPath, "git", "add", "."); err != nil {
			return fmt.Errorf(Style("error staging files: %v", ColorRed), err)
		}
		if err := runCommand(projectPath, "git", "commit", "-m", fmt.Sprintf("feat: initial commit for %s (%s)", name, kind)); err != nil {
			return fmt.Errorf(Style("error committing files: %v", ColorRed), err)
		}
		fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Initialized & Committed Git repository", StyleBold))
	}

	// Final success block
	fmt.Printf("\n%s\n", Style("------------------------------------------------", ColorGreen))
	fmt.Printf("%s\n", Style(fmt.Sprintf(" [ ✨ SUCCESS ] Project '%s' (%s) is READY!", name, kind), ColorGreen, StyleBold))
	fmt.Printf("   %s\n", Style(fmt.Sprintf("Location: %s", projectPath), StyleBold, ColorCyan))
	fmt.Printf("%s\n\n", Style("------------------------------------------------", ColorGreen))

	cleanupOnFailure = false
	return nil
}

func createProject(c *cli.Context) error {
	name := c.Args().First()
	kind := c.String("kind")
	modulePath := c.String("module-path")
	cliLib := c.String("cli-lib")
	git := c.Bool("git")

	if name == "" {
		return cli.Exit(Style("Error: Project name is required.", ColorRed, StyleBold), 1)
	}

	if strings.ContainsAny(name, "/\\") || name == "." || name == ".." {
		return cli.Exit(Style(fmt.Sprintf("Error: Invalid project name '%s'.", name), ColorRed, StyleBold), 1)
	}

	projectPath := path.Clean(path.Join(projectsDir, name))

	if modulePath == "" {
		modulePath = name
	}

	err := createAndCleanup(projectPath, name, kind, modulePath, cliLib, git)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}
	return nil
}

// --- COMMAND IMPLEMENTATIONS (DELETE & LIST - Unchanged from previous review) ---

func deleteProject(c *cli.Context) error {
	name := c.Args().First()

	if name == "" {
		return cli.Exit(Style("Error: Project name is required.", ColorRed, StyleBold), 1)
	}

	projectPath := path.Clean(path.Join(projectsDir, name))

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return cli.Exit(Style(fmt.Sprintf("Error: Project '%s' not found at %s", name, projectPath), ColorRed, StyleBold), 1)
	}

	fmt.Printf(Style("\n [ ⚠️  DANGER ] Are you sure you want to delete project: ", ColorRed, StyleBold)+"%s\n", projectPath)
	if !confirmPrompt(Style("         > Delete all contents? (", ColorRed) + Style("Y", StyleBold) + Style("/n): ", ColorRed)) {
		return cli.Exit(Style("Deletion aborted by user.", ColorYellow), 1)
	}

	fmt.Println("\n" + Style("--- Deleting Project: "+name+" ---", StyleBold, ColorMagenta))
	if err := os.RemoveAll(projectPath); err != nil {
		return cli.Exit(fmt.Sprintf(Style("Error deleting directory: %v", ColorRed), err), 1)
	}

	fmt.Println(Style(" [ ✔️ ] ", ColorGreen) + Style("Project deleted:", StyleBold) + " " + Style(projectPath, StyleItalic))
	fmt.Printf("\n%s\n", Style(fmt.Sprintf(" [ 🗑️  SUCCESS ] Project '%s' is GONE.", name), ColorGreen, StyleBold))

	return nil
}

func detectProjectType(contents []os.DirEntry) (projectType string, isGitRepo bool) {
	projectType = "unknown"
	isGitRepo = false
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
		if name == ".git" && c.IsDir() {
			isGitRepo = true
		}
		if projectType == "unknown" && !c.IsDir() {
			if t, ok := typeCheck[name]; ok {
				projectType = t
			}
		}
		if projectType != "unknown" && isGitRepo {
			break
		}
	}
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

type ListComponent struct {
	Dir string
}

const (
	colProjectNameWidth = 25
	colTypeWidth        = 10
	colGitWidth         = 5
	colTotalWidth       = colProjectNameWidth + colTypeWidth + colGitWidth + 6
)

// ... printHeader, printRow, printFooter are omitted for brevity, they are unchanged ...
func (l *ListComponent) printHeader() {
	headerLine := strings.Repeat("-", colTotalWidth)
	fmt.Printf("\n%s\n", Style(fmt.Sprintf(" PROJECTS IN %s ", strings.ToUpper(l.Dir)), ColorMagenta, StyleBold))
	fmt.Println(Style(headerLine, ColorBlue))
	fmt.Printf(Style("| %-*s | %-*s | %-*s |\n", ColorBlue, StyleBold),
		colProjectNameWidth, "PROJECT",
		colTypeWidth, "TYPE",
		colGitWidth, "GIT")
	fmt.Println(Style(headerLine, ColorBlue))
}
func (l *ListComponent) printRow(name, pType string, isGit bool) {
	gitRepoStatus := Style("❌", ColorRed)
	if isGit {
		gitRepoStatus = Style("✔️", ColorGreen)
	}
	typeStr := pType
	typeColor := ColorReset
	switch pType {
	case "go":
		typeColor = ColorBlue
	case "rust":
		typeColor = ColorRed
	case "node":
		typeColor = ColorGreen
	case "python":
		typeColor = ColorYellow
	}
	coloredType := Style(typeStr, typeColor)
	fmt.Printf("| %s%-*s%s | %s%-*s%s | %s%-*s%s |\n",
		StyleBold, colProjectNameWidth, name, ColorReset,
		"", colTypeWidth, coloredType, ColorReset,
		"", colGitWidth, gitRepoStatus, ColorReset)
}
func (l *ListComponent) printFooter(count int) {
	footerLine := strings.Repeat("-", colTotalWidth)
	fmt.Println(Style(footerLine, ColorBlue))
	fmt.Printf("\n%s\n\n", Style(fmt.Sprintf("Total projects found: %d", count), StyleBold, ColorCyan))
}

func listProjects(c *cli.Context) error {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return cli.Exit(fmt.Sprintf(Style("Error reading projects directory: %v", ColorRed), err), 1)
	}
	listComp := ListComponent{Dir: projectsDir}
	var projectsCount int
	listComp.printHeader()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectName := entry.Name()
		projectPath := path.Join(projectsDir, projectName)
		contents, err := os.ReadDir(projectPath)
		if err != nil {
			log.Printf(Style("[ ⚠️ ] Warning: Could not read directory %s: %v", ColorYellow), projectName, err)
			continue
		}
		projectType, isGitRepo := detectProjectType(contents)
		projectsCount++
		listComp.printRow(projectName, projectType, isGitRepo)
	}
	listComp.printFooter(projectsCount)
	return nil
}

// --- MAIN APPLICATION ENTRY POINT ---

func main() {
	log.SetFlags(0)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting user home directory: %v", err)
	}

	projectsDir = os.Getenv("PROJECTS_DIR")
	if projectsDir == "" {
		projectsDir = path.Join(homeDir, "projects")
	}

	app := &cli.App{
		Name:    "forge",
		Usage:   "The ultimate project launcher. Fast. Minimal. Dope.",
		Version: "0.2.1",
		Commands: []*cli.Command{
			{
				Name:      "new",
				Usage:     "Creates a fresh project in the projects directory",
				ArgsUsage: Style("<project-name>", StyleBold, ColorCyan),
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
						Usage:   "Initialize, **add files**, and **commit** a git repository " + Style("🤘", ColorMagenta),
					},
					&cli.StringFlag{
						Name:    "cli-lib",
						Aliases: []string{"cl"},
						Usage:   "CLI framework to use when kind is 'cli'. Options: flag, cobra, cli",
						Value:   "flag",
					},
				},
				Action: createProject,
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "See what's cookin' in the projects directory",
				Action:  listProjects,
			},
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     Style("Permanently deletes a project!", ColorRed, StyleBold),
				ArgsUsage: Style("<project-name>", StyleBold, ColorCyan),
				Action:    deleteProject,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		if _, ok := err.(cli.ExitCoder); !ok {
			log.Fatal(err)
		}
	}
}
