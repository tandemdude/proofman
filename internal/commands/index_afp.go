package commands

import (
	"github.com/tandemdude/proofman/pkg/indexer"
	"github.com/urfave/cli/v2"
)

func index(cCtx *cli.Context) error {
	afpPath := cCtx.String("afp-path")
	repoUrl := cCtx.String("repo-url")

	idx, err := indexer.New(afpPath, repoUrl)
	if err != nil {
		return err
	}

	err = idx.Index()
	if err != nil {
		return err
	}

	return nil
}

// git@github.com:proofman-dev/afp-index.git
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
	},
}
