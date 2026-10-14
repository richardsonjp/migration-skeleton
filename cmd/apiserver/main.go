package main

import (
	"os"

	"github.com/spf13/cobra"

	"migration-skeleton/cmd/apiserver/app"
	"migration-skeleton/config"
)

var (
	rootCMD = &cobra.Command{
		Short: "migration-skeleton",
	}

	migrateCMD = &cobra.Command{
		Use:   "migrate [command]",
		Short: "Run database migrations",
		// app.Migrate owns its own FlagSet (--schema, plus positional steps).
		// Without this, cobra parses the args first and rejects --schema as an
		// unknown flag before app.Migrate ever sees it.
		DisableFlagParsing: true,
		Run: func(c *cobra.Command, args []string) {
			app.Migrate(args)
		},
	}
)

func main() {
	cobra.OnInitialize(config.Init)

	// Regist
	rootCMD.AddCommand(migrateCMD)
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
}
