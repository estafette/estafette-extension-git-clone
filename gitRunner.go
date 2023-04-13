package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"regexp"
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

func gitCloneOverride(ctx context.Context, gitName, gitURL, gitBranch, gitRevision, subdir string, shallowClone bool, shallowCloneDepth int) (err error) {

	log.Info().Msgf("Cloning git repository %v to branch %v into subdir %v with shallow clone is %v and depth %v...", gitName, gitBranch, subdir, shallowClone, shallowCloneDepth)

	// git clone
	err = gitCloneWithRetry(ctx, gitName, gitURL, gitBranch, shallowClone, shallowCloneDepth, subdir, 3)
	if err != nil {
		return
	}

	// checkout specific revision
	if gitRevision != "" {
		err = gitCheckout(ctx, gitRevision)
		if err != nil {
			log.Info().Msgf("Finished cloning git repository %v to revision %v into subdir %v with shallow clone is %v and depth %v", gitName, gitRevision, subdir, shallowClone, shallowCloneDepth)
			return
		}
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
		args = []string{"clone", fmt.Sprintf("--branch=%v", gitBranch), "--verbose", "--progress", gitURL, targetDirectory}
		if shallowClone {
			args = []string{"clone", fmt.Sprintf("--depth=%v", shallowCloneDepth), fmt.Sprintf("--branch=%v", gitBranch), "--verbose", "--progress", "--no-tags", gitURL, targetDirectory}
		}
	}

	err = foundation.RunCommandWithArgsExtended(ctx, "git", args)

	if err != nil {
		return
	}

	gitModules := filepath.Join(targetDirectory, ".gitmodules")
	if foundation.FileExists(gitModules) {
		log.Info().Msg("Found .gitmodules")
		// update .gitmodules with git credentials
		var data []byte
		data, err = ioutil.ReadFile(gitModules)
		if err != nil {
			return
		}
		gitCreds := ""
		switch ctx.Value("source") {
		case "bitbucket":
			gitCreds = "https://x-token-auth:" + ctx.Value("token").(string) + "@"
		case "github":
			gitCreds = "https://x-access-token:" + ctx.Value("token").(string) + "@"
		case "cloudsource":
			gitCreds = "https://estafette:" + ctx.Value("token").(string) + "@"
		default:
			return errors.New("invalid git source expected bitbucket, github or cloudsource")
		}
		data = regexp.MustCompile(`:`).ReplaceAll(data, []byte("/"))
		data = regexp.MustCompile(`https://`).ReplaceAll(data, []byte(gitCreds))
		data = regexp.MustCompile(`git@`).ReplaceAll(data, []byte(gitCreds))
		err = ioutil.WriteFile(gitModules, data, 0644)
		if err != nil {
			return
		}

		log.Info().Msg("Initializing submodules")
		err = foundation.RunCommandInDirectoryWithArgsExtended(ctx, targetDirectory, "git", []string{"submodule", "init"})
		if err != nil {
			return
		}

		log.Info().Msg("Updating submodules")
		err = foundation.RunCommandInDirectoryWithArgsExtended(ctx, targetDirectory, "git", []string{"submodule", "update"})
		if err != nil {
			return
		}

		// restore .gitmodules file
		err = foundation.RunCommandInDirectoryWithArgsExtended(ctx, targetDirectory, "git", []string{"checkout", ".gitmodules"})
		if err != nil {
			return
		}
	}
	return
}

func gitCheckout(ctx context.Context, gitRevision string) (err error) {
	args := []string{"fetch", "origin", fmt.Sprintf("HEAD:refs/heads/%s", gitRevision)}
	err = foundation.RunCommandWithArgsExtended(ctx, "git", args)
	if err != nil {
		return
	}

	args = []string{"checkout", "--quiet", "--force", gitRevision}
	err = foundation.RunCommandWithArgsExtended(ctx, "git", args)
	if err != nil {
		return
	}

	return
}

func getTargetDir(subdir string) string {
	return filepath.Join("/estafette-work", subdir)
}
