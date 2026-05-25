package manifest

import (
	"context"
	"fmt"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// SyncResult describes what happened during a sync operation.
type SyncResult struct {
	Merged         types.AppManifest
	WriteBack      WriteBackResult
	HasDifferences bool
}

// Sync performs two-way manifest sync between local and remote. It fetches
// both manifests, computes diffs, prompts the user for resolution, writes
// the merged result to both the API and the local file, and returns the result.
func Sync(ctx context.Context, clients *shared.ClientFactory, app types.App, auth types.SlackAuth) (*SyncResult, error) {
	localManifest, err := clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
	if err != nil {
		return nil, slackerror.New("Failed to get local manifest").WithRootCause(err).WithCode(slackerror.ErrInvalidManifest)
	}

	remoteManifest, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return nil, slackerror.New("Failed to get remote manifest from app settings").WithRootCause(err)
	}

	diffs, err := Diff(localManifest.AppManifest, remoteManifest.AppManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to compute manifest differences: %w", err)
	}

	if !diffs.HasDifferences() {
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji:     "books",
			Text:      "App Manifest",
			Secondary: []string{"Project manifest and app settings are in sync"},
		}))
		return &SyncResult{Merged: localManifest.AppManifest, HasDifferences: false}, nil
	}

	DisplayDiffs(ctx, clients.IO, diffs)

	if !clients.IO.IsTTY() {
		return nil, slackerror.New(slackerror.ErrAppManifestUpdate).
			WithRemediation("Run %s interactively to resolve manifest differences, or use %s to overwrite app settings",
				style.CommandText("slack manifest sync"),
				style.CommandText("--force"),
			)
	}

	merged, err := resolveInteractively(ctx, clients, localManifest.AppManifest, remoteManifest.AppManifest, diffs)
	if err != nil {
		return nil, err
	}

	// Push merged manifest to API
	clients.IO.PrintInfo(ctx, false, "\n  Syncing manifest...")
	_, err = clients.API().UpdateApp(ctx, auth.Token, app.AppID, merged, true, true)
	if err != nil {
		return nil, slackerror.New("Failed to update app settings with merged manifest").WithRootCause(err)
	}
	clients.IO.PrintInfo(ctx, false, "  %s Updated app settings", style.Green("✓"))

	// Write back to local file
	workingDir := clients.SDKConfig.WorkingDirectory
	writeResult, err := WriteManifestLocal(clients.Fs, workingDir, merged)
	if err != nil {
		return nil, err
	}
	if writeResult.Written {
		clients.IO.PrintInfo(ctx, false, "  %s Updated %s", style.Green("✓"), manifestFileName)
	} else if writeResult.Warning != "" {
		clients.IO.PrintInfo(ctx, false, "  %s %s", style.Yellow("!"), writeResult.Warning)
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "books",
		Text:      "App Manifest",
		Secondary: []string{fmt.Sprintf("Finished manifest sync for %q", localManifest.DisplayInformation.Name)},
	}))

	return &SyncResult{Merged: merged, WriteBack: writeResult, HasDifferences: true}, nil
}

func resolveInteractively(ctx context.Context, clients *shared.ClientFactory, local, remote types.AppManifest, diffs *DiffResult) (types.AppManifest, error) {
	strategy, err := PromptResolutionStrategy(ctx, clients.IO)
	if err != nil {
		return types.AppManifest{}, err
	}

	switch strategy {
	case MergeAllLocal:
		return MergeAllFrom(local, remote, diffs, MergeAllLocal)
	case MergeAllRemote:
		return MergeAllFrom(local, remote, diffs, MergeAllRemote)
	default:
		resolutions, err := PromptFieldResolutions(ctx, clients.IO, diffs)
		if err != nil {
			return types.AppManifest{}, err
		}
		return Merge(local, remote, resolutions)
	}
}
