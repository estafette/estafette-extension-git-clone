package main

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"time"

	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

func gitCloneRevision(ctx context.Context, gitName, gitURL, gitBranch, gitRevision string, shallowClone bool, shallowCloneDepth int) (err error) {

	log.Info().Msgf("Cloning git repository %v to branch %v and revision %v with shallow clone is %v and depth %v...", gitName, gitBranch, gitRevision, shallowClone, shallowCloneDepth)

	// git clone
	err = gitCloneWithRetry(ctx, gitName, gitURL, gitBranch, shallowClone, shallowCloneDepth, ".", 3)
	if err != nil {
		return
	}

	// checkout specific revision
	if gitRevision != "" {
		err = gitCheckout(ctx, gitRevision)
		if err != nil {
			return
		}
	}

	log.Info().Msgf("Finished cloning git repository %v to branch %v and revision %v with shallow clone is %v and depth %v", gitName, gitBranch, gitRevision, shallowClone, shallowCloneDepth)

	return
}

func gitCloneOverride(ctx context.Context, gitName, gitURL, gitBranch, subdir string, shallowClone bool, shallowCloneDepth int) (err error) {

	log.Info().Msgf("Cloning git repository %v to branch %v into subdir %v with shallow clone is %v and depth %v...", gitName, gitBranch, subdir, shallowClone, shallowCloneDepth)

	// git clone
	err = gitCloneWithRetry(ctx, gitName, gitURL, gitBranch, shallowClone, shallowCloneDepth, subdir, 3)
	if err != nil {
		return
	}

	log.Info().Msgf("Finished cloning git repository %v to branch %v into subdir %v with shallow clone is %v and depth %v", gitName, gitBranch, subdir, shallowClone, shallowCloneDepth)

	return
}

func gitCloneWithRetry(ctx context.Context, gitName, gitURL, gitBranch string, shallowClone bool, shallowCloneDepth int, subdir string, retries int) (err error) {

	attempt := 0

	for attempt == 0 || (err != nil && attempt < retries) {

		err = gitClone(ctx, gitName, gitURL, gitBranch, shallowClone, shallowCloneDepth, subdir)
		if err != nil {
			log.Info().Msgf("Attempt %v cloning git repository %v to branch %v and revision %v failed: %v", attempt, gitName, gitBranch, gitRevision, err)
		}

		// wait with exponential backoff
		<-time.After(time.Duration(math.Pow(2, float64(attempt))) * time.Second)

		attempt++
	}

	return
}

func gitClone(ctx context.Context, gitName, gitURL, gitBranch string, shallowClone bool, shallowCloneDepth int, subdir string) (err error) {

	targetDirectory := getTargetDir(subdir)

	args := []string{"clone", fmt.Sprintf("--branch=%v", gitBranch), gitURL, targetDirectory}
	if shallowClone {
		args = []string{"clone", fmt.Sprintf("--depth=%v", shallowCloneDepth), fmt.Sprintf("--branch=%v", gitBranch), "--no-tags", gitURL, targetDirectory}
	}

	if runtime.GOOS == "windows" {
		args = []string{"clone", fmt.Sprintf("--branch=%v", gitBranch), "--verbose", gitURL, targetDirectory}
		if shallowClone {
			args = []string{"clone", fmt.Sprintf("--depth=%v", shallowCloneDepth), fmt.Sprintf("--branch=%v", gitBranch), "--verbose", "--no-tags", gitURL, targetDirectory}
		}
	}

	err = foundation.RunCommandWithArgsExtended(ctx, "git", args)

	if err != nil {
		return
	}
	return
}

func gitCheckout(ctx context.Context, gitRevision string) (err error) {

	args := []string{"checkout", "--quiet", "--force", gitRevision}

	err = foundation.RunCommandWithArgsExtended(ctx, "git", args)
	if err != nil {
		return
	}
	return
}

func getTargetDir(subdir string) string {
	return filepath.Join("/estafette-work", subdir)
}
