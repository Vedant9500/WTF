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

var projectBoosts = map[ProjectType]map[string]float64{
	ProjectTypeGit: {
		"git": 2.0, "commit": 1.5, "branch": 1.5, "merge": 1.5,
		"pull": 1.5, "push": 1.5, "clone": 1.5, "checkout": 1.5,
	},
	ProjectTypeDocker: {
		"docker": 2.0, "container": 1.8, "image": 1.5, "build": 1.3,
		"run": 1.3, "compose": 1.5,
	},
	ProjectTypeNode: {
		"npm": 2.0, "yarn": 2.0, "node": 1.8, "javascript": 1.5,
		"package": 1.3, "install": 1.3,
	},
	ProjectTypePython: {
		"python": 2.0, "pip": 2.0, "virtual": 1.5, "venv": 1.5,
		"conda": 1.5, "requirements": 1.3,
	},
	ProjectTypeGo: {
		"go": 2.0, "mod": 1.8, "build": 1.5, "test": 1.5, "run": 1.3,
	},
	ProjectTypeRust: {
		"cargo": 2.0, "rust": 1.8, "build": 1.5, "test": 1.5,
	},
	ProjectTypeJava: {
		"java": 2.0, "maven": 1.8, "gradle": 1.8, "build": 1.5, "compile": 1.5,
	},
	ProjectTypeDotNet: {
		"dotnet": 2.0, "nuget": 1.8, "build": 1.5, "restore": 1.5,
	},
	ProjectTypeRuby: {
		"ruby": 2.0, "gem": 1.8, "bundle": 1.5, "rake": 1.5,
	},
	ProjectTypePHP: {
		"php": 2.0, "composer": 1.8, "artisan": 1.5, "laravel": 1.3,
	},
	ProjectTypeC: {
		"gcc": 2.0, "make": 1.8, "cmake": 1.8, "compile": 1.5, "build": 1.5,
	},
	ProjectTypeCpp: {
		"gcc": 2.0, "make": 1.8, "cmake": 1.8, "compile": 1.5, "build": 1.5,
	},
	ProjectTypeKubernetes: {
		"kubectl": 2.0, "kubernetes": 1.8, "k8s": 1.8, "pod": 1.5,
		"service": 1.3, "deploy": 1.3,
	},
	ProjectTypeTerraform: {
		"terraform": 2.0, "plan": 1.8, "apply": 1.8, "destroy": 1.5, "init": 1.5,
	},
	ProjectTypeAnsible: {
		"ansible": 2.0, "playbook": 1.8, "inventory": 1.5, "vault": 1.5,
	},
	ProjectTypeWebpack: {
		"webpack": 2.0, "build": 1.5, "bundle": 1.5,
	},
	ProjectTypeVite: {
		"vite": 2.0, "build": 1.5, "dev": 1.5,
	},
	ProjectTypeMake: {
		"make": 2.0, "build": 1.5,
	},
}

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
		if boostsMap, ok := projectBoosts[projectType]; ok {
			for k, v := range boostsMap {
				boosts[k] = v
			}
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
// GetContextDescription returns a human-readable description of the detected context
func (ctx *Context) GetContextDescription() string {
	if len(ctx.ProjectTypes) == 0 {
		return "generic directory"
	}

	projectDescriptions := map[ProjectType]string{
		ProjectTypeGit:        "Git repository",
		ProjectTypeDocker:     "Docker project",
		ProjectTypeNode:       "Node.js project",
		ProjectTypePython:     "Python project",
		ProjectTypeGo:         "Go project",
		ProjectTypeRust:       "Rust project",
		ProjectTypeJava:       "Java project",
		ProjectTypeDotNet:     ".NET project",
		ProjectTypeRuby:       "Ruby project",
		ProjectTypePHP:        "PHP project",
		ProjectTypeC:          "C/C++ project",
		ProjectTypeCpp:        "C++ project",
		ProjectTypeKubernetes: "Kubernetes deployment",
		ProjectTypeTerraform:  "Terraform infrastructure",
		ProjectTypeAnsible:    "Ansible playbook",
		ProjectTypeWebpack:    "Webpack project",
		ProjectTypeVite:       "Vite project",
		ProjectTypeMake:       "Makefile project",
		ProjectTypeGeneric:    "generic directory",
	}

	var descriptions []string
	for _, projectType := range ctx.ProjectTypes {
		if desc, ok := projectDescriptions[projectType]; ok {
			descriptions = append(descriptions, desc)
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
