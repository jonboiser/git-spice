package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"go.abhg.dev/gs/internal/git"
	"go.abhg.dev/gs/internal/spice"
	"go.abhg.dev/gs/internal/spice/state"
	"go.abhg.dev/gs/internal/text"
	"go.abhg.dev/gs/internal/ui"
)

type branchSquashCmd struct {
	Message string `short:"m" placeholder:"MSG" help:"Use the given message as the commit message."`
}

func (*branchSquashCmd) Help() string {
	return text.Dedent(`
		Squash all commits in the current branch into a single commit
		and restack upstack branches.
	`)
}

func (cmd *branchSquashCmd) Run(
	ctx context.Context,
	log *log.Logger,
	view ui.View,
	repo *git.Repository,
	store *state.Store,
	svc *spice.Service,
) (err error) {
	branchName, err := repo.CurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}

	branch, err := svc.LookupBranch(ctx, branchName)
	if err != nil {
		return fmt.Errorf("lookup branch %q: %w", branchName, err)
	}

	if err := svc.VerifyRestacked(ctx, branchName); err != nil {
		var restackErr *spice.BranchNeedsRestackError
		if errors.As(err, &restackErr) {
			return fmt.Errorf("branch %v needs to be restacked before it can be squashed", branchName)
		}
		return fmt.Errorf("verify restacked: %w", err)
	}

	// If no message was specified,
	// combine the commit messages of all commits in the branch
	// to form the initial commit message for the squashed commit.
	var commitTemplate string
	if cmd.Message == "" {
		commitMessages, err := repo.CommitMessageRange(ctx, branch.Head.String(), branch.BaseHash.String())
		if err != nil {
			return fmt.Errorf("get commit messages: %w", err)
		}

		var sb strings.Builder
		sb.WriteString("The original commit messages were:\n\n")
		for i, msg := range commitMessages {
			if i > 0 {
				sb.WriteString("\n")
			}

			fmt.Fprintf(&sb, "%v\n", msg)
		}

		commitTemplate = sb.String()
	}

	// Detach the HEAD so that we don't mess with the current branch
	// until the operation is confirmed successful.
	if err := repo.DetachHead(ctx, branchName); err != nil {
		return fmt.Errorf("detach HEAD: %w", err)
	}
	var reattachedHead bool
	defer func() {
		// Reattach the HEAD to the original branch
		// if we return early before the operation is complete.
		if !reattachedHead {
			if cerr := repo.Checkout(ctx, branchName); cerr != nil {
				log.Error("could not check out original branch",
					"branch", branchName,
					"error", cerr)
				err = errors.Join(err, cerr)
			}
		}
	}()

	// Reset the detached branch to the base commit
	if err := repo.Reset(ctx, branch.BaseHash.String(), git.ResetOptions{Mode: git.ResetSoft}); err != nil {
		return err
	}

	if err := repo.Commit(ctx, git.CommitRequest{
		Message:  cmd.Message,
		Template: commitTemplate,
	}); err != nil {
		return fmt.Errorf("commit squashed changes: %w", err)
	}

	// Replace the HEAD ref of `branchName` with the new commit
	headHash, err := repo.Head(ctx)
	if err != nil {
		return err
	}

	if err := repo.SetRef(ctx, git.SetRefRequest{
		Ref:  "refs/heads/" + branchName,
		Hash: headHash,
	}); err != nil {
		return fmt.Errorf("update branch ref: %w", err)
	}

	if cerr := repo.Checkout(ctx, branchName); cerr != nil {
		return fmt.Errorf("checkout branch: %w", cerr)
	}
	reattachedHead = true

	return (&upstackRestackCmd{}).Run(ctx, log, repo, store, svc)
}
