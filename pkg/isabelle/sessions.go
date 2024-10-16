package isabelle

import (
	"bytes"
	"fmt"
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/logging"
	"github.com/tandemdude/proofman/pkg/localcache"
	"github.com/tandemdude/proofman/pkg/parser"
	"io"
	"net/http"
	"strings"
)

const (
	isabelleRootsUrl = "https://isabelle.in.tum.de/repos/isabelle/raw-file/%s/ROOTS"
	isabelleRootUrl  = "https://isabelle.in.tum.de/repos/isabelle/raw-file/%s/%s/ROOT"
)

func FetchBuiltinSessions(versionOverride string) ([]string, error) {
	var version string
	var err error

	if versionOverride != "" {
		version = versionOverride
	} else {
		version, err = Version()
		if err != nil {
			return nil, err
		}
	}

	logging.Verbose("checking local cache for list of builtin sessions")
	existing := localcache.ReadFile(fmt.Sprintf("%s.builtin_sessions", version))

	builtinSessions := make([]string, 0)
	if len(existing) > 0 {
		logging.Verbose("found cached list of builtin sessions")
		for _, elem := range strings.Split(existing, "\n") {
			builtinSessions = append(builtinSessions, strings.TrimSpace(elem))
		}
		return builtinSessions, nil
	}

	// download and cache the ROOTS file from the isabelle repository
	logging.Verbose("fetching ROOTS file from the Isabelle repository")
	res, err := http.Get(fmt.Sprintf(isabelleRootsUrl, version))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	logging.Verbose("ROOTS file fetched successfully")

	parsedRootLocations := strings.Split(string(rawBody), "\n")
	allRootLocations := make([]string, 0)
	for _, loc := range parsedRootLocations {
		loc = strings.TrimSpace(loc)
		if len(loc) > 0 {
			allRootLocations = append(allRootLocations, loc)
		}
	}

	logging.Verbose("parsed ROOTS file - contains %d entries", len(allRootLocations))
	for _, rootLocation := range allRootLocations {
		logging.Verbose("fetching ROOT file for %s", rootLocation)
		res, err = http.Get(fmt.Sprintf(isabelleRootUrl, version, rootLocation))
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(res.Body)
		if err != nil {
			logging.Unquiet("error reading %s ROOT file - %s", rootLocation, err)
			return nil, err
		}
		res.Body.Close()

		parsed, err := parser.ParseRootFile(bytes.NewReader(content))
		if err != nil {
			logging.Unquiet("parsing %s ROOT file failed - %s", rootLocation, err)
			internal.WriteFile(".failed.ROOT", string(content), false)

			return nil, err
		}

		for _, chapter := range parsed.Chapters {
			for _, session := range chapter.Sessions {
				builtinSessions = append(builtinSessions, session.Name)
			}
		}
	}

	logging.Verbose("found %d builtin sessions successfully - saving to cache", len(builtinSessions))
	err = localcache.WriteFile(fmt.Sprintf("%s.builtin_sessions", version), strings.Join(builtinSessions, "\n"))
	if err != nil {
		return nil, err
	}

	return builtinSessions, nil
}
