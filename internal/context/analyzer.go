// Package context provides intelligent project context detection and analysis.
//
// This package analyzes the current working directory to detect:
//   - Project types (Git, Docker, Node.js, Python, Go, Rust, Java, etc.)
//   - Build systems (Maven, Gradle, CMake, Make)
//   - Infrastructure tools (Kubernetes, Terraform, Ansible)
//   - Package managers and their scripts
//   - Development environment context for search relevance boosting
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

	// Process each file to detect project types
	for _, file := range files {
		a.analyzeFile(file.Name(), dir, ctx)
	}

	// Finalize context
	a.finalizeContext(ctx)

	return ctx, nil
}

// analyzeFile processes a single file to detect project indicators
func (a *Analyzer) analyzeFile(filename, dir string, ctx *Context) {
	// Check different categories of files
	a.checkVersionControl(filename, ctx)
	a.checkContainerization(filename, ctx)
	a.checkJavaScript(filename, dir, ctx)
	a.checkPython(filename, ctx)
	a.checkGo(filename, ctx)
	a.checkRust(filename, ctx)
	a.checkJava(filename, ctx)
	a.checkDotNet(filename, ctx)
	a.checkRuby(filename, ctx)
	a.checkPHP(filename, ctx)
	a.checkCCpp(filename, dir, ctx)
	a.checkInfrastructure(filename, ctx)
}

// checkVersionControl detects version control systems
func (a *Analyzer) checkVersionControl(filename string, ctx *Context) {
	if filename == ".git" {
		ctx.HasGit = true
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeGit)
	}
}

// checkContainerization detects containerization tools
func (a *Analyzer) checkContainerization(filename string, ctx *Context) {
	if filename == "Dockerfile" || filename == "docker-compose.yml" || filename == "docker-compose.yaml" {
		ctx.HasDocker = true
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeDocker)
	}
}

// checkJavaScript detects JavaScript/Node.js projects
func (a *Analyzer) checkJavaScript(filename, dir string, ctx *Context) {
	switch filename {
	case "package.json":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeNode)
		a.setLanguageIfEmpty(ctx, "javascript")
		a.extractPackageScripts(filepath.Join(dir, filename), ctx)
	case "node_modules", "yarn.lock", "pnpm-lock.yaml":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeNode)
		a.setLanguageIfEmpty(ctx, "javascript")
	case "webpack.config.js", "webpack.config.ts":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeWebpack)
	case "vite.config.js", "vite.config.ts":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeVite)
	}
}

// checkPython detects Python projects
func (a *Analyzer) checkPython(filename string, ctx *Context) {
	if filename == "requirements.txt" || filename == "setup.py" || filename == "pyproject.toml" || filename == "Pipfile" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypePython)
		a.setLanguageIfEmpty(ctx, "python")
	}
}

// checkGo detects Go projects
func (a *Analyzer) checkGo(filename string, ctx *Context) {
	if filename == "go.mod" || filename == "go.sum" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeGo)
		a.setLanguageIfEmpty(ctx, "go")
	}
}

// checkRust detects Rust projects
func (a *Analyzer) checkRust(filename string, ctx *Context) {
	if filename == "Cargo.toml" || filename == "Cargo.lock" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeRust)
		a.setLanguageIfEmpty(ctx, "rust")
	}
}

// checkJava detects Java projects and build systems
func (a *Analyzer) checkJava(filename string, ctx *Context) {
	switch filename {
	case "pom.xml":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeJava)
		ctx.BuildSystem = "maven"
		a.setLanguageIfEmpty(ctx, "java")
	case "build.gradle", "build.gradle.kts":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeJava)
		ctx.BuildSystem = "gradle"
		a.setLanguageIfEmpty(ctx, "java")
	}
}

// checkDotNet detects .NET projects
func (a *Analyzer) checkDotNet(filename string, ctx *Context) {
	if strings.HasSuffix(filename, ".csproj") || strings.HasSuffix(filename, ".vbproj") || strings.HasSuffix(filename, ".fsproj") {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeDotNet)
		a.setLanguageIfEmpty(ctx, "csharp")
	} else if filename == "global.json" || filename == "nuget.config" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeDotNet)
	}
}

// checkRuby detects Ruby projects
func (a *Analyzer) checkRuby(filename string, ctx *Context) {
	if filename == "Gemfile" || filename == "Rakefile" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeRuby)
		a.setLanguageIfEmpty(ctx, "ruby")
	}
}

// checkPHP detects PHP projects
func (a *Analyzer) checkPHP(filename string, ctx *Context) {
	if filename == "composer.json" || filename == "composer.lock" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypePHP)
		a.setLanguageIfEmpty(ctx, "php")
	}
}

// checkCCpp detects C/C++ projects and build systems
func (a *Analyzer) checkCCpp(filename, dir string, ctx *Context) {
	switch filename {
	case "CMakeLists.txt":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeC)
		ctx.BuildSystem = "cmake"
		a.setLanguageIfEmpty(ctx, "c")
	case "Makefile", "makefile":
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeMake)
		a.extractMakeTargets(filepath.Join(dir, filename), ctx)
	}
}

// checkInfrastructure detects infrastructure and DevOps tools
func (a *Analyzer) checkInfrastructure(filename string, ctx *Context) {
	// Kubernetes
	if (strings.Contains(filename, "k8s") || strings.Contains(filename, "kubernetes")) &&
		(strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")) {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeKubernetes)
	} else if filename == "kustomization.yaml" || filename == "kustomization.yml" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeKubernetes)
	}

	// Terraform
	if strings.HasSuffix(filename, ".tf") || strings.HasSuffix(filename, ".tfvars") {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeTerraform)
	}

	// Ansible
	if filename == "ansible.cfg" || filename == "hosts" || filename == "inventory" {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeAnsible)
	} else if strings.Contains(filename, "playbook") && (strings.HasSuffix(filename, ".yml") || strings.HasSuffix(filename, ".yaml")) {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeAnsible)
	}
}

// setLanguageIfEmpty sets the language if it's not already set
func (a *Analyzer) setLanguageIfEmpty(ctx *Context, language string) {
	if ctx.Language == "" {
		ctx.Language = language
	}
}

// finalizeContext performs final processing on the context
func (a *Analyzer) finalizeContext(ctx *Context) {
	// Remove duplicates from project types
	ctx.ProjectTypes = removeDuplicateProjectTypes(ctx.ProjectTypes)

	// If no specific project types were detected, mark as generic
	if len(ctx.ProjectTypes) == 0 {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeGeneric)
	}
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

	for script := range ctx.PackageScripts {
		boosts[script] = 1.3
	}

	for _, target := range ctx.MakeTargets {
		boosts[target] = 1.3
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
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "\t") {
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				target := strings.TrimSpace(parts[0])
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
