package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

//go:embed all:templates
var templatesFS embed.FS

// ProjectType defines the category of project to scaffold.
type ProjectType string

const (
	ProjectTypeCLI ProjectType = "cli"
	ProjectTypeLib ProjectType = "lib"
)

// Options holds parameters for scaffolding a new Go CLI or library project.
type Options struct {
	Type        ProjectType
	TargetDir   string
	Name        string
	Binary      string
	PkgName     string
	Module      string
	Description string
	Author      string
	Year        int
	GoVersion   string
	CreatedDate string
	Force       bool
	DryRun      bool
	InitGit     bool
	RunTidy     bool
}

// Result describes the outcome of a scaffolding generation.
type Result struct {
	Type        ProjectType
	ProjectDir  string
	ProjectName string
	BinaryName  string
	PkgName     string
	ModuleName  string
	Files       []string
	DryRun      bool
	GitError    error
	TidyError   error
	Warnings    []string
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

// CleanPackageName sanitizes a string into a valid Go package identifier.
func CleanPackageName(name string) string {
	cleaned := nonAlphanumericRegex.ReplaceAllString(strings.ToLower(name), "")
	if cleaned == "" || (cleaned[0] >= '0' && cleaned[0] <= '9') {
		cleaned = "pkg" + cleaned
	}
	return cleaned
}

// DefaultOptions returns an Options struct populated with sensible defaults.
func DefaultOptions(targetDir string, pType ProjectType) Options {
	if targetDir == "" {
		targetDir = "."
	}

	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		absDir = targetDir
	}

	if pType == "" {
		pType = ProjectTypeCLI
	}

	opts := Options{
		Type:        pType,
		TargetDir:   absDir,
		Author:      "Alex Gorbatchev",
		Year:        time.Now().Year(),
		GoVersion:   "1.26.2",
		CreatedDate: time.Now().Format("2006-01-02 15:04"),
		InitGit:     true,
		RunTidy:     true,
	}
	return opts
}

// Normalize ensures derived names and defaults are populated.
func (o *Options) Normalize() {
	if o.Type == "" {
		o.Type = ProjectTypeCLI
	}

	if o.TargetDir == "" {
		o.TargetDir = "."
	}
	if abs, err := filepath.Abs(o.TargetDir); err == nil {
		o.TargetDir = abs
	}

	if o.Name == "" {
		o.Name = filepath.Base(o.TargetDir)
		if o.Name == "." || o.Name == "/" || o.Name == "" {
			if o.Type == ProjectTypeLib {
				o.Name = "mylib"
			} else {
				o.Name = "my-cli"
			}
		}
	}

	if o.Type == ProjectTypeLib {
		if o.PkgName == "" {
			o.PkgName = CleanPackageName(o.Name)
		}
		if o.Description == "" {
			o.Description = fmt.Sprintf("A Go library for %s", o.Name)
		}
	} else {
		if o.Binary == "" {
			if strings.HasSuffix(o.Name, "-cli") && len(o.Name) > 4 {
				o.Binary = strings.TrimSuffix(o.Name, "-cli")
			} else {
				o.Binary = o.Name
			}
		}
		if o.Description == "" {
			o.Description = fmt.Sprintf("A command-line tool for %s", o.Name)
		}
	}

	if o.Module == "" {
		o.Module = "github.com/alexgorbatchev/" + o.Name
	}

	if o.Author == "" {
		o.Author = "Alex Gorbatchev"
	}

	if o.Year == 0 {
		o.Year = time.Now().Year()
	}

	if o.GoVersion == "" {
		o.GoVersion = "1.26.2"
	}

	if o.CreatedDate == "" {
		o.CreatedDate = time.Now().Format("2006-01-02 15:04")
	}
}

// Generate renders and writes all scaffolding files for a Go project.
func Generate(opts Options) (*Result, error) {
	opts.Normalize()

	if !opts.DryRun {
		// Check target directory
		entries, err := os.ReadDir(opts.TargetDir)
		if err == nil && len(entries) > 0 && !opts.Force {
			return nil, fmt.Errorf("target directory %q is not empty (use --force to overwrite)", opts.TargetDir)
		}
	}

	res := &Result{
		Type:        opts.Type,
		ProjectDir:  opts.TargetDir,
		ProjectName: opts.Name,
		BinaryName:  opts.Binary,
		PkgName:     opts.PkgName,
		ModuleName:  opts.Module,
		DryRun:      opts.DryRun,
		Warnings:    make([]string, 0),
	}

	tmplRoot := filepath.Join("templates", string(opts.Type))

	// Walk templates embed.FS for the given project type
	err := fs.WalkDir(templatesFS, tmplRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		// Calculate relative path inside templates/<type>/
		relPath, err := filepath.Rel(tmplRoot, path)
		if err != nil {
			return fmt.Errorf("computing rel path for %s: %w", path, err)
		}

		// Destination path removes the trailing .tmpl extension
		destRel := strings.TrimSuffix(relPath, ".tmpl")

		// Handle type-specific file renames
		if opts.Type == ProjectTypeCLI {
			slashPath := filepath.ToSlash(destRel)
			if strings.HasPrefix(slashPath, "cmd/app/") {
				subPath := strings.TrimPrefix(slashPath, "cmd/app/")
				destRel = filepath.Join("cmd", opts.Binary, filepath.FromSlash(subPath))
			} else if slashPath == "cmd/app" {
				destRel = filepath.Join("cmd", opts.Binary)
			}
		} else if opts.Type == ProjectTypeLib {
			if destRel == "lib.go" {
				destRel = opts.PkgName + ".go"
			} else if destRel == "lib_test.go" {
				destRel = opts.PkgName + "_test.go"
			}
		}

		// Read template content from embed.FS
		tmplBytes, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading embedded template %s: %w", path, err)
		}

		t, err := template.New(relPath).Parse(string(tmplBytes))
		if err != nil {
			return fmt.Errorf("parsing template for %s: %w", relPath, err)
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, opts); err != nil {
			return fmt.Errorf("rendering template for %s: %w", relPath, err)
		}

		res.Files = append(res.Files, destRel)

		if !opts.DryRun {
			destPath := filepath.Join(opts.TargetDir, destRel)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("creating directory for %s: %w", destRel, err)
			}

			if err := os.WriteFile(destPath, buf.Bytes(), 0644); err != nil {
				return fmt.Errorf("writing file %s: %w", destRel, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("generating files: %w", err)
	}

	if !opts.DryRun {
		// Initialize git if requested and .git does not exist
		if opts.InitGit {
			gitDir := filepath.Join(opts.TargetDir, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				gitCmd := exec.Command("git", "init")
				gitCmd.Dir = opts.TargetDir
				if out, err := gitCmd.CombinedOutput(); err != nil {
					res.GitError = fmt.Errorf("git init failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
					res.Warnings = append(res.Warnings, res.GitError.Error())
				}
			}
		}

		// Run go mod tidy if requested
		if opts.RunTidy {
			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = opts.TargetDir
			if out, err := tidyCmd.CombinedOutput(); err != nil {
				res.TidyError = fmt.Errorf("go mod tidy failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
				res.Warnings = append(res.Warnings, res.TidyError.Error())
			}
		}
	}

	return res, nil
}
