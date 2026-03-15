package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type Generator struct {
	ProjectName string
	ModulePath  string
	Features    map[Feature]bool
	HTTPPort    string
	GRPCPort    string
}

func (g *Generator) Has(feat Feature) bool {
	return g.Features[feat]
}

type templateFile struct {
	tmpl    string
	output  string
}

func (g *Generator) Run() error {
	if err := g.createDirectories(); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	if err := g.renderTemplates(); err != nil {
		return fmt.Errorf("render templates: %w", err)
	}

	return nil
}

func (g *Generator) createDirectories() error {
	dirs := []string{
		"cmd/server",
		"configs",
		"internal/app",
		"internal/core",
		"pkg",
	}

	if g.Has(FeatureHTTP) {
		dirs = append(dirs, "internal/adapters/http")
	}
	if g.Has(FeatureGRPC) {
		dirs = append(dirs, "internal/adapters/grpc", "proto")
	}
	if g.Has(FeaturePostgreSQL) {
		dirs = append(dirs, "internal/adapters/repository", "scripts")
	}
	if g.Has(FeatureJWTAuth) {
		dirs = append(dirs, "pkg/jwt")
	}
	if g.Has(FeatureRateLimit) {
		dirs = append(dirs, "pkg/ratelimit")
	}
	if g.Has(FeaturePostgreSQL) || g.Has(FeatureRedis) {
		dirs = append(dirs, "deploy")
	}

	dirs = append(dirs, "pkg/logger", "pkg/validator")

	for _, dir := range dirs {
		fullPath := filepath.Join(g.ProjectName, dir)
		if err := os.MkdirAll(fullPath, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", fullPath, err)
		}
	}
	return nil
}

func (g *Generator) renderTemplates() error {
	files := []templateFile{
		{tmpl: "templates/go.mod.tmpl", output: "go.mod"},
		{tmpl: "templates/main.go.tmpl", output: "cmd/server/main.go"},
		{tmpl: "templates/config.yaml.tmpl", output: "configs/config.local.yaml"},
		{tmpl: "templates/Makefile.tmpl", output: "Makefile"},
		{tmpl: "templates/Dockerfile.tmpl", output: "Dockerfile"},
		{tmpl: "templates/README.md.tmpl", output: "README.md"},
	}

	funcMap := template.FuncMap{
		"has": func(feat string) bool {
			return g.Features[Feature(feat)]
		},
	}

	for _, f := range files {
		data, err := templateFS.ReadFile(f.tmpl)
		if err != nil {
			return fmt.Errorf("read template %s: %w", f.tmpl, err)
		}

		tmpl, err := template.New(filepath.Base(f.tmpl)).Funcs(funcMap).Parse(string(data))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", f.tmpl, err)
		}

		outPath := filepath.Join(g.ProjectName, f.output)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", outPath, err)
		}

		out, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", outPath, err)
		}

		if err := tmpl.Execute(out, g); err != nil {
			out.Close()
			return fmt.Errorf("execute template %s: %w", f.tmpl, err)
		}
		out.Close()
	}

	return nil
}
