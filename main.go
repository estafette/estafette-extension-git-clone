package main

import (
	"encoding/json"
	"fmt"
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
	gitSource            = kingpin.Flag("git-source", "The source of the repository.").Envar("ESTAFETTE_GIT_SOURCE").Required().String()
	gitOwner             = kingpin.Flag("git-owner", "The owner of the repository.").Envar("ESTAFETTE_GIT_OWNER").Required().String()
	gitName              = kingpin.Flag("git-name", "The owner plus repository name.").Envar("ESTAFETTE_GIT_NAME").Required().String()
	gitBranch            = kingpin.Flag("git-branch", "The branch to clone.").Envar("ESTAFETTE_GIT_BRANCH").Required().String()
	gitRevision          = kingpin.Flag("git-revision", "The revision to check out.").Envar("ESTAFETTE_GIT_REVISION").String()
	shallowClone         = kingpin.Flag("shallow-clone", "Shallow clone git repository for improved clone time.").Default("true").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_SHALLOW").Bool()
	overrideRepo         = kingpin.Flag("override-repo", "Set other repository name to clone from same owner.").Envar("ESTAFETTE_EXTENSION_REPO").String()
	overrideBranch       = kingpin.Flag("override-branch", "Set other repository branch to clone from same owner.").Envar("ESTAFETTE_EXTENSION_BRANCH").String()
	overrideSubdirectory = kingpin.Flag("override-directory", "Set other repository directory to clone from same owner.").Envar("ESTAFETTE_EXTENSION_SUBDIR").String()

	bitbucketAPITokenJSON = kingpin.Flag("bitbucket-api-token", "Bitbucket api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_BITBUCKET_API_TOKEN").String()
	githubAPITokenJSON    = kingpin.Flag("github-api-token", "Github api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_GITHUB_API_TOKEN").String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// log to stdout and hide timestamp
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	// log startup message
	log.Printf("Starting estafette-extension-git-clone version %v...", version)

	// get api token from injected credentials
	bitbucketAPIToken := ""
	if *bitbucketAPITokenJSON != "" {
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*bitbucketAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal("Failed unmarshalling injected credentials: ", err)
		}
		if len(credentials) == 0 {
			log.Fatal("No credentials have been injected")
		}
		bitbucketAPIToken = credentials[0].AdditionalProperties.Token
	}

	githubAPIToken := ""
	if *githubAPITokenJSON != "" {
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*githubAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal("Failed unmarshalling injected credentials: ", err)
		}
		if len(credentials) == 0 {
			log.Fatal("No credentials have been injected")
		}
		githubAPIToken = credentials[0].AdditionalProperties.Token
	}

	if *overrideRepo != "" {
		if *overrideBranch == "" {
			*overrideBranch = "master"
		}
		if *overrideSubdirectory == "" {
			*overrideSubdirectory = *overrideRepo
		}

		overrideGitURL := ""
		if bitbucketAPIToken != "" {
			overrideGitURL = fmt.Sprintf("https://x-token-auth:%v@%v/%v/%v", bitbucketAPIToken, *gitSource, *gitOwner, *overrideRepo)
		}
		if githubAPIToken != "" {
			overrideGitURL = fmt.Sprintf("https://x-access-token:%v@%v/%v/%v", githubAPIToken, *gitSource, *gitOwner, *overrideRepo)
		}

		if overrideGitURL == "" {
			log.Fatalf("Failed generating url for cloning git repository %v to branch %v into subdir %v", *overrideRepo, *overrideBranch, *overrideSubdirectory)
		}

		// git clone the specified repository branch to the specific directory
		err := gitCloneOverride(*overrideRepo, overrideGitURL, *overrideBranch, *overrideSubdirectory, *shallowClone)
		if err != nil {
			log.Fatalf("Error cloning git repository %v to branch %v into subdir %v: %v", *overrideRepo, *overrideBranch, *overrideSubdirectory, err)
		}

		return
	}

	gitURL := ""
	if bitbucketAPIToken != "" {
		gitURL = fmt.Sprintf("https://x-token-auth:%v@%v/%v/%v", bitbucketAPIToken, *gitSource, *gitOwner, *gitName)
	}
	if githubAPIToken != "" {
		gitURL = fmt.Sprintf("https://x-access-token:%v@%v/%v/%v", githubAPIToken, *gitSource, *gitOwner, *gitName)
	}

	if gitURL == "" {
		log.Fatalf("Failed generating url for cloning repository %v to branch %v and revision %v with shallow clone is %v", *gitName, *gitBranch, *gitRevision, *shallowClone)
	}

	// git clone to specific branch and revision
	err := gitCloneRevision(*gitName, gitURL, *gitBranch, *gitRevision, *shallowClone)
	if err != nil {
		log.Fatalf("Error cloning git repository %v to branch %v and revision %v with shallow clone is %v: %v", *gitName, *gitBranch, *gitRevision, *shallowClone, err)
	}
}
