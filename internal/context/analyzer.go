package context

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType represents different types of projects detected in the current directory
type ProjectType string

const (
	ProjectTypeGit        ProjectType = "git"
	ProjectTypeDocker     ProjectType = "docker"
	ProjectTypeNode       ProjectType = "node"
	ProjectTypePython     ProjectType = "python"
	ProjectTypeGo         ProjectType = "go"
	ProjectTypeRust       ProjectType = "rust"
	ProjectTypeJava       ProjectType = "java"
	ProjectTypeDotNet     ProjectType = "dotnet"
	ProjectTypeRuby       ProjectType = "ruby"
	ProjectTypePHP        ProjectType = "php"
	ProjectTypeC          ProjectType = "c"
	ProjectTypeCpp        ProjectType = "cpp"
	ProjectTypeKubernetes ProjectType = "kubernetes"
	ProjectTypeTerraform  ProjectType = "terraform"
	ProjectTypeAnsible    ProjectType = "ansible"
	ProjectTypeWebpack    ProjectType = "webpack"
	ProjectTypeVite       ProjectType = "vite"
	ProjectTypeMake       ProjectType = "make"
	ProjectTypeGeneric    ProjectType = "generic"
)

// Context holds information about the current working directory
type Context struct {
	WorkingDir     string
	ProjectTypes   []ProjectType
	HasGit         bool
	HasDocker      bool
	Language       string
	PackageScripts map[string]string // For package.json scripts
	MakeTargets    []string          // For Makefile targets
	BuildSystem    string            // maven, gradle, cmake, etc.
}

// Analyzer analyzes the current directory for project context
type Analyzer struct{}

// NewAnalyzer creates a new context analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzeCurrentDirectory analyzes the current working directory
func (a *Analyzer) AnalyzeCurrentDirectory() (*Context, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return a.AnalyzeDirectory(wd)
}

// AnalyzeDirectory analyzes a specific directory for project context
func (a *Analyzer) AnalyzeDirectory(dir string) (*Context, error) {
	ctx := &Context{
		WorkingDir:     dir,
		ProjectTypes:   []ProjectType{},
		PackageScripts: make(map[string]string),
		MakeTargets:    []string{},
	}

	// Check for various project indicators
	files, err := os.ReadDir(dir)
	if err != nil {
		return ctx, nil // Return what we have, don't fail
	}

	for _, file := range files {
		name := file.Name()

		switch {
		// Git repository
		case name == ".git":
			ctx.HasGit = true
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeGit)

		// Docker
		case name == "Dockerfile" || name == "docker-compose.yml" || name == "docker-compose.yaml":
			ctx.HasDocker = true
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeDocker)

		// Node.js
		case name == "package.json":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeNode)
			if ctx.Language == "" {
				ctx.Language = "javascript"
			}
			// Extract npm scripts
			a.extractPackageScripts(filepath.Join(dir, name), ctx)

		case name == "node_modules" || name == "yarn.lock" || name == "pnpm-lock.yaml":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeNode)
			if ctx.Language == "" {
				ctx.Language = "javascript"
			}

		// Webpack/Vite
		case name == "webpack.config.js" || name == "webpack.config.ts":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeWebpack)
		case name == "vite.config.js" || name == "vite.config.ts":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeVite)

		// Python
		case name == "requirements.txt" || name == "setup.py" || name == "pyproject.toml" || name == "Pipfile":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypePython)
			if ctx.Language == "" {
				ctx.Language = "python"
			}

		// Go
		case name == "go.mod" || name == "go.sum":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeGo)
			if ctx.Language == "" {
				ctx.Language = "go"
			}

		// Rust
		case name == "Cargo.toml" || name == "Cargo.lock":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeRust)
			if ctx.Language == "" {
				ctx.Language = "rust"
			}

		// Java/Maven/Gradle
		case name == "pom.xml":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeJava)
			ctx.BuildSystem = "maven"
			if ctx.Language == "" {
				ctx.Language = "java"
			}
		case name == "build.gradle" || name == "build.gradle.kts":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeJava)
			ctx.BuildSystem = "gradle"
			if ctx.Language == "" {
				ctx.Language = "java"
			}

		// .NET
		case strings.HasSuffix(name, ".csproj") || strings.HasSuffix(name, ".vbproj") || strings.HasSuffix(name, ".fsproj"):
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeDotNet)
			if ctx.Language == "" {
				ctx.Language = "csharp"
			}
		case name == "global.json" || name == "nuget.config":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeDotNet)

		// Ruby
		case name == "Gemfile" || name == "Rakefile":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeRuby)
			if ctx.Language == "" {
				ctx.Language = "ruby"
			}

		// PHP
		case name == "composer.json" || name == "composer.lock":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypePHP)
			if ctx.Language == "" {
				ctx.Language = "php"
			}

		// C/C++
		case name == "CMakeLists.txt":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeC)
			ctx.BuildSystem = "cmake"
			if ctx.Language == "" {
				ctx.Language = "c"
			}
		case name == "Makefile" || name == "makefile":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeMake)
			// Extract make targets
			a.extractMakeTargets(filepath.Join(dir, name), ctx)

		// Kubernetes
		case strings.Contains(name, "k8s") || strings.Contains(name, "kubernetes"):
			if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
				ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeKubernetes)
			}
		case name == "kustomization.yaml" || name == "kustomization.yml":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeKubernetes)

		// Terraform
		case strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tfvars"):
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeTerraform)

		// Ansible
		case name == "ansible.cfg" || name == "hosts" || name == "inventory":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeAnsible)
		case strings.Contains(name, "playbook") && (strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")):
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeAnsible)
		}
	}

	// Remove duplicates from project types
	ctx.ProjectTypes = removeDuplicateProjectTypes(ctx.ProjectTypes)

	return ctx, nil
}

// GetContextBoosts returns keyword boosts based on the detected context
func (ctx *Context) GetContextBoosts() map[string]float64 {
	boosts := make(map[string]float64)

	for _, projectType := range ctx.ProjectTypes {
		switch projectType {
		case ProjectTypeGit:
			boosts["git"] = 2.0
			boosts["commit"] = 1.5
			boosts["branch"] = 1.5
			boosts["merge"] = 1.5
			boosts["pull"] = 1.5
			boosts["push"] = 1.5
			boosts["clone"] = 1.5
			boosts["checkout"] = 1.5

		case ProjectTypeDocker:
			boosts["docker"] = 2.0
			boosts["container"] = 1.8
			boosts["image"] = 1.5
			boosts["build"] = 1.3
			boosts["run"] = 1.3
			boosts["compose"] = 1.5

		case ProjectTypeNode:
			boosts["npm"] = 2.0
			boosts["yarn"] = 2.0
			boosts["node"] = 1.8
			boosts["javascript"] = 1.5
			boosts["package"] = 1.3
			boosts["install"] = 1.3

		case ProjectTypePython:
			boosts["python"] = 2.0
			boosts["pip"] = 2.0
			boosts["virtual"] = 1.5
			boosts["venv"] = 1.5
			boosts["conda"] = 1.5
			boosts["requirements"] = 1.3

		case ProjectTypeGo:
			boosts["go"] = 2.0
			boosts["mod"] = 1.8
			boosts["build"] = 1.5
			boosts["test"] = 1.5
			boosts["run"] = 1.3

		case ProjectTypeRust:
			boosts["cargo"] = 2.0
			boosts["rust"] = 1.8
			boosts["build"] = 1.5
			boosts["test"] = 1.5

		case ProjectTypeJava:
			boosts["java"] = 2.0
			boosts["maven"] = 1.8
			boosts["gradle"] = 1.8
			boosts["build"] = 1.5
			boosts["compile"] = 1.5

		case ProjectTypeDotNet:
			boosts["dotnet"] = 2.0
			boosts["nuget"] = 1.8
			boosts["build"] = 1.5
			boosts["restore"] = 1.5

		case ProjectTypeRuby:
			boosts["ruby"] = 2.0
			boosts["gem"] = 1.8
			boosts["bundle"] = 1.5
			boosts["rake"] = 1.5

		case ProjectTypePHP:
			boosts["php"] = 2.0
			boosts["composer"] = 1.8
			boosts["artisan"] = 1.5
			boosts["laravel"] = 1.3

		case ProjectTypeC, ProjectTypeCpp:
			boosts["gcc"] = 2.0
			boosts["make"] = 1.8
			boosts["cmake"] = 1.8
			boosts["compile"] = 1.5
			boosts["build"] = 1.5

		case ProjectTypeKubernetes:
			boosts["kubectl"] = 2.0
			boosts["kubernetes"] = 1.8
			boosts["k8s"] = 1.8
			boosts["pod"] = 1.5
			boosts["service"] = 1.3
			boosts["deploy"] = 1.3

		case ProjectTypeTerraform:
			boosts["terraform"] = 2.0
			boosts["plan"] = 1.8
			boosts["apply"] = 1.8
			boosts["destroy"] = 1.5
			boosts["init"] = 1.5

		case ProjectTypeAnsible:
			boosts["ansible"] = 2.0
			boosts["playbook"] = 1.8
			boosts["inventory"] = 1.5
			boosts["vault"] = 1.5

		case ProjectTypeWebpack:
			boosts["webpack"] = 2.0
			boosts["build"] = 1.5
			boosts["bundle"] = 1.5

		case ProjectTypeVite:
			boosts["vite"] = 2.0
			boosts["build"] = 1.5
			boosts["dev"] = 1.5

		case ProjectTypeMake:
			boosts["make"] = 2.0
			boosts["build"] = 1.5
		}
	}

	// Add boosts for detected scripts and targets
	for script := range ctx.PackageScripts {
		boosts[script] = 1.3 // Boost npm script names
	}

	for _, target := range ctx.MakeTargets {
		boosts[target] = 1.3 // Boost make target names
	}

	return boosts
}

// GetContextDescription returns a human-readable description of the detected context
func (ctx *Context) GetContextDescription() string {
	if len(ctx.ProjectTypes) == 0 {
		return "generic directory"
	}

	var descriptions []string
	for _, projectType := range ctx.ProjectTypes {
		switch projectType {
		case ProjectTypeGit:
			descriptions = append(descriptions, "Git repository")
		case ProjectTypeDocker:
			descriptions = append(descriptions, "Docker project")
		case ProjectTypeNode:
			descriptions = append(descriptions, "Node.js project")
		case ProjectTypePython:
			descriptions = append(descriptions, "Python project")
		case ProjectTypeGo:
			descriptions = append(descriptions, "Go project")
		case ProjectTypeRust:
			descriptions = append(descriptions, "Rust project")
		case ProjectTypeJava:
			descriptions = append(descriptions, "Java project")
		case ProjectTypeDotNet:
			descriptions = append(descriptions, ".NET project")
		case ProjectTypeRuby:
			descriptions = append(descriptions, "Ruby project")
		case ProjectTypePHP:
			descriptions = append(descriptions, "PHP project")
		case ProjectTypeC:
			descriptions = append(descriptions, "C/C++ project")
		case ProjectTypeCpp:
			descriptions = append(descriptions, "C++ project")
		case ProjectTypeKubernetes:
			descriptions = append(descriptions, "Kubernetes deployment")
		case ProjectTypeTerraform:
			descriptions = append(descriptions, "Terraform infrastructure")
		case ProjectTypeAnsible:
			descriptions = append(descriptions, "Ansible playbook")
		case ProjectTypeWebpack:
			descriptions = append(descriptions, "Webpack project")
		case ProjectTypeVite:
			descriptions = append(descriptions, "Vite project")
		case ProjectTypeMake:
			descriptions = append(descriptions, "Makefile project")
		case ProjectTypeGeneric:
			descriptions = append(descriptions, "generic directory")
		}
	}

	result := strings.Join(descriptions, ", ")

	// Add build system info if available
	if ctx.BuildSystem != "" {
		result += " (" + ctx.BuildSystem + ")"
	}

	return result
}

// extractPackageScripts extracts npm scripts from package.json
func (a *Analyzer) extractPackageScripts(packagePath string, ctx *Context) {
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return
	}

	ctx.PackageScripts = pkg.Scripts
}

// extractMakeTargets extracts targets from Makefile
func (a *Analyzer) extractMakeTargets(makefilePath string, ctx *Context) {
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for targets (lines ending with :)
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "\t") {
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				target := strings.TrimSpace(parts[0])
				// Skip variables and special targets
				if !strings.Contains(target, "=") && !strings.HasPrefix(target, ".") && target != "" {
					ctx.MakeTargets = append(ctx.MakeTargets, target)
				}
			}
		}
	}
}

// removeDuplicateProjectTypes removes duplicate project types
func removeDuplicateProjectTypes(types []ProjectType) []ProjectType {
	seen := make(map[ProjectType]bool)
	var result []ProjectType

	for _, t := range types {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}

	return result
}
