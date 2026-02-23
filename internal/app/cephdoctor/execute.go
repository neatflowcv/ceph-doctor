package cephdoctor

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/neatflowcv/ceph-doctor/internal/infrastructure/fscluster"
)

func Execute() error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	repo, err := fscluster.NewRepository("")
	if err != nil {
		return fmt.Errorf("new repository: %w", err)
	}

	var command cli

	parser, err := kong.New(
		&command,
		kong.Name("cephdoctor"),
		kong.Description("Ceph Doctor CLI"),
		kong.BindTo(repo, (*domain.ClusterRepository)(nil)),
	)
	if err != nil {
		return fmt.Errorf("create parser: %w", err)
	}

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) && parseErr.Context != nil {
			_ = parseErr.Context.PrintUsage(false)
		}

		return fmt.Errorf("parse args: %w", err)
	}

	err = ctx.Run()
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	return nil
}
