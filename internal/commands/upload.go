package commands

import (
	"bytes"
	"errors"
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/logging"
	"github.com/tandemdude/proofman/pkg/proofbank"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func upload(cCtx *cli.Context) error {
	fname := cCtx.Args().First()
	if fname == "" {
		return errors.New("FILE is a required argument")
	}
	if !strings.HasSuffix(fname, ".ipkg") {
		return errors.New("FILE must be a package archive (*.ipkg)")
	}

	logging.Unquiet("reading archive file")
	handle, err := os.ReadFile(fname)
	if err != nil {
		return err
	}

	client, err := proofbank.NewAuthenticatedClient(internal.ProofbankBaseUrl, internal.ProofbankApiToken)
	if err != nil {
		return err
	}

	logging.Unquiet("uploading archive to index")
	err = client.UploadPackage(bytes.NewReader(handle))
	if err != nil {
		return err
	}

	logging.Unquiet("upload completed")

	return nil
}

var UploadCommand = &cli.Command{
	Name:      "upload",
	Usage:     "Uploads a package to ProofBank",
	Action:    upload,
	Args:      true,
	ArgsUsage: "FILE",
}
