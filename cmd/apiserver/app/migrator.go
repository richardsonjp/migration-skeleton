package app

import (
	"flag"
	"fmt"
	"os"

	"strconv"
	
	"migration-skeleton/cmd/apiserver/app/db"
	"migration-skeleton/config"
)

// Migrate runs database migrations
// direction: "up", "down", "force <version>", "version"
func Migrate(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: migrate <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  up [steps]      Run up migrations (all or specific steps)")
		fmt.Println("  down [steps]    Run down migrations (all or specific steps)")
		fmt.Println("  version         Show current migration version")
		fmt.Println("  force <version> Force specific version")
		fmt.Println("\nFlags:")
		fmt.Println("  --schema=<schema>  Specific schema (default: all)")
		os.Exit(1)
	}

	command := args[0]
	
	// Use a local FlagSet to avoid global state issues and allow multiple calls if needed (though unlikely here)
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	schemaFlag := fs.String("schema", "", "Database schema to run migrations in (default: all schemas)")
	
	// Parse flags from args[1:]
	// Expected format: migrate <command> [--schema=...] [steps|version]
	// Note: flags must come before positional arguments for the Go flag package to parse them correctly.
	// We will adjust Makefile to ensure this order.
	if err := fs.Parse(args[1:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	targetSchema := *schemaFlag
	remainingArgs := fs.Args()
	
	var steps int
	if len(remainingArgs) > 0 {
		if s, err := strconv.Atoi(remainingArgs[0]); err == nil {
			steps = s
		}
	}

	// Determine if we're running for a specific schema or all schemas
	if targetSchema == "" {
		// Run for all schemas
		schemas := config.Config.GetSchemas()
		fmt.Printf("Running migrations for all schemas: %v\n", schemas)
	} else {
		fmt.Printf("Running migrations for schema: %s\n", targetSchema)
	}

	switch command {
	case "up":
		if targetSchema == "" {
			if err := db.RunUpAll(steps); err != nil {
				fmt.Printf("Migrate up failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("All migrations up completed")
		} else {
			if err := db.RunUp(targetSchema, steps); err != nil {
				fmt.Printf("Migrate up failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Migration up completed for schema: %s\n", targetSchema)
		}
	case "down":
		if targetSchema == "" {
			if err := db.RunDownAll(steps); err != nil {
				fmt.Printf("Migrate down failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("All migrations down completed")
		} else {
			if err := db.RunDown(targetSchema, steps); err != nil {
				fmt.Printf("Migrate down failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Migration down completed for schema: %s\n", targetSchema)
		}
	case "version":
		if targetSchema == "" {
			versions := db.GetAllVersions()
			fmt.Println("\nMigration versions:")
			for schema, info := range versions {
				if info.Error != nil {
					fmt.Printf("  %s: ERROR - %v\n", schema, info.Error)
				} else {
					fmt.Printf("  %s: Version %d, Dirty: %v\n", schema, info.Version, info.Dirty)
				}
			}
		} else {
			version, dirty, err := db.GetVersion(targetSchema)
			if err != nil {
				fmt.Printf("Check version failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Schema: %s, Version: %d, Dirty: %v\n", targetSchema, version, dirty)
		}
	case "force":
		if len(remainingArgs) < 1 {
			fmt.Println("Force requires a version argument")
			os.Exit(1)
		}
		var v int
		_, err := fmt.Sscanf(remainingArgs[0], "%d", &v)
		if err != nil {
			fmt.Printf("Invalid version: %v\n", err)
			os.Exit(1)
		}

		if targetSchema == "" {
			schemas := config.Config.GetSchemas()
			fmt.Printf("Forcing version %d for all schemas: %v\n", v, schemas)
			for _, schema := range schemas {
				if err := db.Force(v, schema); err != nil {
					fmt.Printf("Force failed for schema %s: %v\n", schema, err)
					os.Exit(1)
				}
				fmt.Printf("Force completed for schema: %s\n", schema)
			}
		} else {
			if err := db.Force(v, targetSchema); err != nil {
				fmt.Printf("Force failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Force completed for schema: %s\n", targetSchema)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
