package context

import (
	"os"
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
	ProjectTypeGeneric    ProjectType = "generic"
)

// Context holds information about the current working directory
type Context struct {
	WorkingDir   string
	ProjectTypes []ProjectType
	HasGit       bool
	HasDocker    bool
	Language     string
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
		WorkingDir:   dir,
		ProjectTypes: []ProjectType{},
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
		case name == "package.json" || name == "node_modules" || name == "yarn.lock" || name == "pnpm-lock.yaml":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeNode)
			if ctx.Language == "" {
				ctx.Language = "javascript"
			}
			
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
			
		// Java
		case name == "pom.xml" || name == "build.gradle" || name == "gradle.properties":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeJava)
			if ctx.Language == "" {
				ctx.Language = "java"
			}
			
		// .NET
		case strings.HasSuffix(name, ".csproj") || strings.HasSuffix(name, ".sln") || name == "global.json":
			ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeDotNet)
			if ctx.Language == "" {
				ctx.Language = "csharp"
			}
		}
	}
	
	// If no specific project type detected, mark as generic
	if len(ctx.ProjectTypes) == 0 {
		ctx.ProjectTypes = append(ctx.ProjectTypes, ProjectTypeGeneric)
	}
	
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
		}
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
		case ProjectTypeGeneric:
			descriptions = append(descriptions, "generic directory")
		}
	}
	
	return strings.Join(descriptions, ", ")
}
