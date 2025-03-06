package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/log"
)

func main() {
	exeName := os.Args[0][strings.LastIndex(os.Args[0], string(os.PathSeparator))+1:]
	logLevel := &logLevelFlag{level: slog.LevelInfo}
	root := &cobra.Command{
		Use: exeName,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			ctx := log.ContextWithLogger(cmd.Context(), slog.New(log.WithAttrsFromContextHandler{
				Parent:            slog.NewTextHandler(cmd.OutOrStdout(), &slog.HandlerOptions{Level: logLevel.level}),
				IgnoredAttributes: []string{"time"},
			}))

			cmd.SetContext(ctx)
			return nil
		},
	}

	root.AddCommand(
		cloudfrontCmd(),
		instanceCmd(),
		logGroupCmd(),
		logStreamCmd(),
		s3BucketCmd(),
		tagCmd(),
		vpcCmd(),
		vpcEndpointCmd(),
		vpcEndpointServiceCmd(),
	)
	root.Flags().Var(logLevel, "log-level", "Level to log at")

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
