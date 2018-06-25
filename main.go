package main

import (
	"log"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin"
)

var (
	version   string
	branch    string
	revision  string
	buildDate string
	goVersion = runtime.Version()
)

var (
	// flags
	gitName      = kingpin.Flag("git-name", "The owner plus repository name.").Envar("ESTAFETTE_GIT_NAME").Required().String()
	gitURL       = kingpin.Flag("git-url", "The authenticated url to clone.").Envar("ESTAFETTE_GIT_URL").Required().String()
	gitBranch    = kingpin.Flag("git-branch", "The branch to clone.").Envar("ESTAFETTE_GIT_BRANCH").Required().String()
	gitRevision  = kingpin.Flag("git-revision", "The revision to check out.").Envar("ESTAFETTE_GIT_REVISION").Required().String()
	shallowClone = kingpin.Flag("shallow-clone", "Shallow clone git repository for improved clone time.").Default("true").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_SHALLOW").Bool()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// log to stdout
	log.SetOutput(os.Stdout)

	// log startup message
	log.Printf("Starting estafette-extension-git-clone version %v...", version)

	// git clone to specific branch and revision
	err := gitCloneRevision(*gitName, *gitURL, *gitBranch, *gitRevision, *shallowClone)
	if err != nil {
		log.Fatalf("Error cloning git repository %v to branch %v and revision %v with shallow clone is %v: %v", *gitName, *gitBranch, *gitRevision, *shallowClone, err)
	}
}
