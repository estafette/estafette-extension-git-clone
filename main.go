package main

import (
	stdlog "log"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	gitName     = kingpin.Flag("git-name", "The owner plus repository name.").Envar("ESTAFETTE_GIT_NAME").Required().String()
	gitURL      = kingpin.Flag("git-url", "The authenticated url to clone.").Envar("ESTAFETTE_GIT_URL").Required().String()
	gitBranch   = kingpin.Flag("git-branch", "The branch to clone.").Envar("ESTAFETTE_GIT_BRANCH").Required().String()
	gitRevision = kingpin.Flag("estafette-build-status", "The revision to check out.").Envar("ESTAFETTE_GIT_REVISION").Required().String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// log as severity for stackdriver logging to recognize the level
	zerolog.LevelFieldName = "severity"

	// set some default fields added to all logs
	log.Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "estafette-extension-git-clone").
		Str("version", version).
		Str("gitName", *gitName).
		Str("gitBranch", *gitBranch).
		Str("gitRevision", *gitRevision).
		Logger()

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)

	// log startup message
	log.Info().
		Str("branch", branch).
		Str("revision", revision).
		Str("buildDate", buildDate).
		Str("goVersion", goVersion).
		Msg("Starting estafette-extension-git-clone...")

	// git clone to specific branch and revision
	err := gitCloneRevision(*gitName, *gitURL, *gitBranch, *gitRevision)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error cloning git repository %v to branch %v and revision %v...", *gitName, *gitBranch, *gitRevision)
	}

	log.Info().
		Msg("Finished estafette-extension-git-clone...")
}
