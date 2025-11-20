package cmd

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/lthummus/big-dill/internal/server"
)

var rootCmd = &cli.Command{
	Name:  "bigdill",
	Usage: "run the big dill voting server",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		log.Info().Msg("hello world")
		s := server.New()
		s.ListenAndServe()

		return nil
	},
}

func Execute() {
	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
