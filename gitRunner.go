package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"time"
)

func gitCloneRevision(gitName, gitURL, gitBranch, gitRevision string, shallowClone bool) (err error) {

	log.Printf("Cloning git repository %v to branch %v and revision %v with shallow clone is %v...", gitName, gitBranch, gitRevision, shallowClone)

	// git clone
	err = gitCloneWithRetry(gitName, gitURL, gitBranch, shallowClone, 3)
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

	log.Printf("Finished cloning git repository %v to branch %v and revision %v with shallow clone is %v", gitName, gitBranch, gitRevision, shallowClone)

	return
}

func gitCloneWithRetry(gitName, gitURL, gitBranch string, shallowClone bool, retries int) (err error) {

	attempt := 0

	for attempt == 0 || (err != nil && attempt < retries) {

		err = gitClone(gitName, gitURL, gitBranch, shallowClone)
		if err != nil {
			log.Printf("Attempt %v cloning git repository %v to branch %v and revision %v failed: %v", attempt, gitName, gitBranch, gitRevision, err)
		}

		// wait with exponential backoff
		<-time.After(time.Duration(math.Pow(2, float64(attempt))) * time.Second)

		attempt++
	}

	return
}

func gitClone(gitName, gitURL, gitBranch string, shallowClone bool) (err error) {

	args := []string{"clone", fmt.Sprintf("--branch=%v", gitBranch), gitURL, "/estafette-work"}
	if shallowClone {
		args = []string{"clone", "--depth=50", fmt.Sprintf("--branch=%v", gitBranch), gitURL, "/estafette-work"}
	}
	gitCloneCommand := exec.Command("git", args...)
	gitCloneCommand.Stdout = os.Stdout
	gitCloneCommand.Stderr = os.Stderr
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
	checkoutCommand.Stdout = os.Stdout
	checkoutCommand.Stderr = os.Stderr
	err = checkoutCommand.Run()
	if err != nil {
		return
	}
	return
}
