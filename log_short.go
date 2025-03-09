package main

import (
	"context"

	"github.com/charmbracelet/log"
	"go.abhg.dev/gs/internal/forge"
	"go.abhg.dev/gs/internal/git"
	"go.abhg.dev/gs/internal/secret"
	"go.abhg.dev/gs/internal/spice"
	"go.abhg.dev/gs/internal/spice/state"
	"go.abhg.dev/gs/internal/text"
)

type logShortCmd struct {
	branchLogCmd
}

func (*logShortCmd) Help() string {
	return text.Dedent(`
		Only branches that are upstack and downstack from the current
		branch are shown.
		Use with the -a/--all flag to show all tracked branches.
	`)
}

func (cmd *logShortCmd) Run(
	ctx context.Context,
	secretStash secret.Stash,
	log *log.Logger,
	repo *git.Repository,
	store *state.Store,
	svc *spice.Service,
	forges *forge.Registry,
) (err error) {
	return cmd.run(ctx, secretStash, &branchLogOptions{
		Log: log,
	}, repo, store, svc, forges)
}
