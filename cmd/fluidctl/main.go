package main

import (
	"fmt"
	"os"

	"github.com/fluidstackio/fluidctl/internal/filesystem"
	"github.com/fluidstackio/fluidctl/internal/instance"
	"github.com/fluidstackio/fluidctl/internal/project"
	"github.com/spf13/cobra"
)

var Version = "v0.0.0"

func rootCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "fluidctl",
		Short: "fluidctl is a command line tool for managing Fluidstack infrastructure",
		Long:  "fluidctl is a command line tool for managing Fluidstack infrastructure",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		Version: Version,
	}

	cmd.PersistentFlags().StringP("url", "U", "https://atlas.fluidstack.io", "Atlas Server URL")
	cmd.PersistentFlags().StringP("format", "F", "yaml", "Output format (json, yaml)")
	cmd.PersistentFlags().StringP("token", "T", "", "Auth token")
	cmd.PersistentFlags().String("client-id", "", "OAuth Client ID")
	cmd.PersistentFlags().String("client-secret", "", "OAuth Client Secret")

	cmd.AddCommand(
		instance.Command(),
		project.Command(),
		filesystem.Command(),
	)

	return &cmd
}

func main() {
	if err := rootCommand().Execute(); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
