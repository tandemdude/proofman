package commands

import (
	"github.com/tandemdude/proofman/internal/logging"
	"github.com/tandemdude/proofman/pkg/indexer"
	"github.com/urfave/cli/v2"
)

func index(cCtx *cli.Context) error {
	afpPath := cCtx.String("afp-path")
	repoUrl := cCtx.String("repo-url")

	if afpPath == "" || repoUrl == "" {
		logging.Quiet("afp-path and repo-url are required flags")
		return nil
	}

	idx, err := indexer.New(afpPath, repoUrl, cCtx.String("use-version"))
	if err != nil {
		return err
	}

	err = idx.Index()
	if err != nil {
		return err
	}

	return nil
}

var IndexAfpCommand = &cli.Command{
	Name:   "index-afp",
	Usage:  "Indexes the AFP and uploads it to the index Git repository",
	Action: index,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "afp-path",
			Usage:    "The `PATH` to the local AFP download directory",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "repo-url",
			Usage:    "The `INDEX_REPO_URL` to upload the indexed packages to",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "use-version",
			Usage: "The `VERSION` of the AFP that is being indexed. If unspecified it will be inferred from the /etc/version file",
		},
	},
}
