package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

type Feature string

const (
	FeatureHTTP       Feature = "http"
	FeatureGRPC       Feature = "grpc"
	FeaturePostgreSQL Feature = "postgres"
	FeatureRedis      Feature = "redis"
	FeatureJWTAuth    Feature = "jwt"
	FeatureRateLimit  Feature = "ratelimit"
)

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Scaffold a new Go backend project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runNew,
}

func runNew(cmd *cobra.Command, args []string) error {
	var (
		projectName string
		modulePath  string
		features    []string
		httpPort    string
		grpcPort    string
	)

	if len(args) > 0 {
		projectName = args[0]
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("Directory name for the new project").
				Value(&projectName).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("project name is required")
					}
					if strings.ContainsAny(s, " /\\") {
						return fmt.Errorf("project name must not contain spaces or slashes")
					}
					return nil
				}),

			huh.NewInput().
				Title("Module path").
				Description("Go module path (e.g. github.com/user/my-service)").
				Value(&modulePath).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("module path is required")
					}
					if !strings.Contains(s, "/") {
						return fmt.Errorf("module path should contain at least one slash")
					}
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Features to include").
				Options(
					huh.NewOption("HTTP API (Gin)", string(FeatureHTTP)).Selected(true),
					huh.NewOption("gRPC API", string(FeatureGRPC)),
					huh.NewOption("PostgreSQL", string(FeaturePostgreSQL)).Selected(true),
					huh.NewOption("Redis", string(FeatureRedis)),
					huh.NewOption("JWT Authentication", string(FeatureJWTAuth)),
					huh.NewOption("Rate Limiting", string(FeatureRateLimit)),
				).
				Value(&features),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("HTTP port").
				Description("Port for the HTTP server").
				Value(&httpPort).
				Placeholder("8080"),

			huh.NewInput().
				Title("gRPC port").
				Description("Port for the gRPC server").
				Value(&grpcPort).
				Placeholder("50051"),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form cancelled: %w", err)
	}

	if httpPort == "" {
		httpPort = "8080"
	}
	if grpcPort == "" {
		grpcPort = "50051"
	}

	featureSet := make(map[Feature]bool)
	for _, f := range features {
		featureSet[Feature(f)] = true
	}

	gen := &Generator{
		ProjectName: projectName,
		ModulePath:  modulePath,
		Features:    featureSet,
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
	}

	if err := gen.Run(); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	printSummary(gen)
	return nil
}

func printSummary(g *Generator) {
	fmt.Println()
	fmt.Println("✓ Project created successfully!")
	fmt.Println()
	fmt.Printf("  Name:   %s\n", g.ProjectName)
	fmt.Printf("  Module: %s\n", g.ModulePath)
	fmt.Printf("  Path:   ./%s/\n", g.ProjectName)
	fmt.Println()

	fmt.Println("  Features:")
	featureNames := map[Feature]string{
		FeatureHTTP:       "HTTP API (Gin)",
		FeatureGRPC:       "gRPC API",
		FeaturePostgreSQL: "PostgreSQL",
		FeatureRedis:      "Redis",
		FeatureJWTAuth:    "JWT Authentication",
		FeatureRateLimit:  "Rate Limiting",
	}
	for feat, name := range featureNames {
		if g.Features[feat] {
			fmt.Printf("    • %s\n", name)
		}
	}

	fmt.Println()
	fmt.Println("  Get started:")
	fmt.Printf("    cd %s\n", g.ProjectName)
	fmt.Println("    go mod tidy")
	fmt.Println("    make run")
	fmt.Println()

	if g.Features[FeaturePostgreSQL] || g.Features[FeatureRedis] {
		fmt.Println("  Start infrastructure:")
		fmt.Println("    make docker-up")
		fmt.Println()
	}

	dir, _ := os.Getwd()
	fmt.Printf("  Full path: %s/%s\n", dir, g.ProjectName)
	fmt.Println()
}
