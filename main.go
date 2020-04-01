package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

var (
	appgroup  string
	app       string
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
	shallowCloneDepth    = kingpin.Flag("shallow-clone-depth", "Depth for shallow clone git repository for improved clone time.").Default("50").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_DEPTH").Int()
	overrideRepo         = kingpin.Flag("override-repo", "Set other repository name to clone from same owner.").Envar("ESTAFETTE_EXTENSION_REPO").String()
	overrideBranch       = kingpin.Flag("override-branch", "Set other repository branch to clone from same owner.").Envar("ESTAFETTE_EXTENSION_BRANCH").String()
	overrideSubdirectory = kingpin.Flag("override-directory", "Set other repository directory to clone from same owner.").Envar("ESTAFETTE_EXTENSION_SUBDIR").String()

	bitbucketAPITokenJSON   = kingpin.Flag("bitbucket-api-token", "Bitbucket api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_BITBUCKET_API_TOKEN").String()
	githubAPITokenJSON      = kingpin.Flag("github-api-token", "Github api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_GITHUB_API_TOKEN").String()
	cloudsourceAPITokenJSON = kingpin.Flag("cloudsource-api-token", "Cloud Source api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_CLOUDSOURCE_API_TOKEN").String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// init log format from envvar ESTAFETTE_LOG_FORMAT
	foundation.InitLoggingFromEnv(appgroup, app, version, branch, revision, buildDate)

	// create context to cancel commands on sigterm
	ctx := foundation.InitCancellationContext(context.Background())

	// get api token from injected credentials
	bitbucketAPIToken := ""
	if *bitbucketAPITokenJSON != "" {
		log.Info().Msg("Unmarshalling injected bitbucket api token credentials")
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*bitbucketAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed unmarshalling injected bitbucket api token credentials")
		}
		if len(credentials) == 0 {
			log.Fatal().Msg("No bitbucket api token credentials have been injected")
		}
		bitbucketAPIToken = credentials[0].AdditionalProperties.Token
	}

	githubAPIToken := ""
	if *githubAPITokenJSON != "" {
		log.Info().Msg("Unmarshalling injected github api token credentials")
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*githubAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed unmarshalling injected github api token credentials")
		}
		if len(credentials) == 0 {
			log.Fatal().Msg("No github api token credentials have been injected")
		}
		githubAPIToken = credentials[0].AdditionalProperties.Token
	}

	cloudsourceAPIToken := ""
	if *cloudsourceAPITokenJSON != "" {
		log.Info().Msg("Unmarshalling injected cloud source api token credentials")
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*cloudsourceAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed unmarshalling injected cloud source api token credentials")
		}
		if len(credentials) == 0 {
			log.Fatal().Msg("No cloud source api token credentials have been injected")
		}
		cloudsourceAPIToken = credentials[0].AdditionalProperties.Token
	}

	if *overrideRepo != "" {
		if *overrideBranch == "" {
			*overrideBranch = "master"
		}
		if *overrideSubdirectory == "" {
			*overrideSubdirectory = *overrideRepo
		}

		overrideGitURL := fmt.Sprintf("https://%v/%v/%v", *gitSource, *gitOwner, *overrideRepo)
		if strings.HasPrefix(*overrideRepo, "https://") {
			// this allows for any public repo to be cloned
			overrideGitURL = *overrideRepo
		} else {
			if bitbucketAPIToken != "" {
				overrideGitURL = fmt.Sprintf("https://x-token-auth:%v@%v/%v/%v", bitbucketAPIToken, *gitSource, *gitOwner, *overrideRepo)
			}
			if githubAPIToken != "" {
				overrideGitURL = fmt.Sprintf("https://x-access-token:%v@%v/%v/%v", githubAPIToken, *gitSource, *gitOwner, *overrideRepo)
			}
			if cloudsourceAPIToken != "" {
				overrideGitURL = fmt.Sprintf("https://estafette:%v@%v/p/%v/r/%v", cloudsourceAPIToken, *gitSource, *gitOwner, *overrideRepo)
			}
		}

		// git clone the specified repository branch to the specific directory
		err := gitCloneOverride(ctx, *overrideRepo, overrideGitURL, *overrideBranch, *overrideSubdirectory, *shallowClone, *shallowCloneDepth)
		if err != nil {
			log.Fatal().Err(err).Msgf("Error cloning git repository %v to branch %v into subdir %v", *overrideRepo, *overrideBranch, *overrideSubdirectory)
		}

		return
	}

	gitURL := fmt.Sprintf("https://%v/%v/%v", *gitSource, *gitOwner, *gitName)
	if bitbucketAPIToken != "" {
		gitURL = fmt.Sprintf("https://x-token-auth:%v@%v/%v/%v", bitbucketAPIToken, *gitSource, *gitOwner, *gitName)
	}
	if githubAPIToken != "" {
		gitURL = fmt.Sprintf("https://x-access-token:%v@%v/%v/%v", githubAPIToken, *gitSource, *gitOwner, *gitName)
	}
	if cloudsourceAPIToken != "" {
		gitURL = fmt.Sprintf("https://estafette:%v@%v/p/%v/r/%v", cloudsourceAPIToken, *gitSource, *gitOwner, *gitName)
	}

	// git clone to specific branch and revision
	err := gitCloneRevision(ctx, *gitName, gitURL, *gitBranch, *gitRevision, *shallowClone, *shallowCloneDepth)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error cloning git repository %v to branch %v and revision %v with shallow clone is %v", *gitName, *gitBranch, *gitRevision, *shallowClone)
	}
}
