package main

import (
	"fmt"
	"math"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

func gitCloneRevision(gitName, gitURL, gitBranch, gitRevision string) (err error) {

	log.Info().
		Str("name", gitName).
		Str("url", gitURL).
		Str("branch", gitBranch).
		Str("revision", gitRevision).
		Msgf("Cloning git repository %v to branch %v and revision %v...", gitName, gitBranch, gitRevision)

	// git clone
	err = gitCloneWithRetry(gitName, gitURL, gitBranch, 3)
	if err != nil {
		return
	}

	// checkout specific revision
	if gitRevision != "" {
		err = gitCheckout(gitRevision)
		if err != nil {
			return
		}
	}

	log.Info().
		Str("name", gitName).
		Str("url", gitURL).
		Str("branch", branch).
		Str("revision", revision).
		Msgf("Finished cloning git repository %v to branch %v and revision %v", gitName, gitBranch, gitRevision)

	return
}

func gitCloneWithRetry(gitName, gitURL, gitBranch string, retries int) (err error) {

	attempt := 0

	for attempt == 0 || (err != nil && attempt < retries) {

		err = gitClone(gitName, gitURL, gitBranch)
		if err != nil {
			log.Debug().Err(err).Msgf("Attempt %v cloning git repository %v to branch %v and revision %v failed", attempt, gitName, gitBranch, gitRevision)
		}

		// wait with exponential backoff
		<-time.After(time.Duration(math.Pow(2, float64(attempt))) * time.Second)

		attempt++
	}

	return
}

func gitClone(gitName, gitURL, gitBranch string) (err error) {

	args := []string{"clone", "--depth=50", fmt.Sprintf("--branch=%v", gitBranch), gitURL, "/estafette-work"}
	gitCloneCommand := exec.Command("git", args...)
	gitCloneCommand.Stdout = log.Logger
	gitCloneCommand.Stderr = log.Logger
	err = gitCloneCommand.Run()
	if err != nil {
		return
	}
	return
}

func gitCheckout(gitRevision string) (err error) {

	args := []string{"checkout", "--quiet", "--force", gitRevision}
	checkoutCommand := exec.Command("git", args...)
	checkoutCommand.Dir = "/estafette-work"
	checkoutCommand.Stdout = log.Logger
	checkoutCommand.Stderr = log.Logger
	err = checkoutCommand.Run()
	if err != nil {
		return
	}
	return
}
