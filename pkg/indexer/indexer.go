package indexer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hashicorp/go-set/v3"
	"github.com/pelletier/go-toml/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/logging"
	"github.com/tandemdude/proofman/pkg/config"
	"github.com/tandemdude/proofman/pkg/indexer/git"
	"github.com/tandemdude/proofman/pkg/isabelle"
	"github.com/tandemdude/proofman/pkg/parser"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var versionRegex = regexp.MustCompile(`(?m)^VERSION=(\d{4}(?:-\d+)*)$`)

var (
	ErrCannotReadVersion            = errors.New("failed reading /etc/version file")
	ErrCannotParseVersion           = errors.New("failed parsing /etc/version file")
	ErrCannotResolveBuiltinSessions = errors.New("failed to resolve isabelle builtin sessions")
	ErrCannotReadROOT               = errors.New("failed reading ROOT file")
	ErrCannotParseROOT              = errors.New("failed parsing ROOT file")
	ErrCannotReadROOTS              = errors.New("failed reading ROOTS file")
	ErrUnknownSession               = errors.New("unknown session required by package")
	ErrCannotSetupIndexRepo         = errors.New("failed setting up index repository")
	ErrCannotMakeBranch             = errors.New("cannot make new index repository branch")
	ErrCannotPopulateRepo           = errors.New("failed populating index repository with new files")
)

type AFPIndexer struct {
	afpVersion         string
	afpDirectoryPath   string
	indexRepositoryUrl string
}

func New(afpDirectoryPath, indexRepositoryUrl, versionOverride string) (*AFPIndexer, error) {
	version := versionOverride

	if version == "" {
		// check the directory exists, and try to read the AFP version from /etc/version
		// if we can't resolve the version or the directory then fail (we won't be able to index it)
		content, err := os.ReadFile(filepath.Join(afpDirectoryPath, "etc", "version"))
		if err != nil {
			return nil, errors.Join(err, ErrCannotReadVersion)
		}

		match := versionRegex.FindSubmatch(content)
		if match == nil {
			return nil, ErrCannotParseVersion
		}
		version = string(match[1])
	}
	logging.Unquiet("AFP directory matches version: %s", version)

	return &AFPIndexer{
		afpVersion:         version,
		afpDirectoryPath:   afpDirectoryPath,
		indexRepositoryUrl: indexRepositoryUrl,
	}, nil
}

func (a *AFPIndexer) theoriesPath() string {
	return filepath.Join(a.afpDirectoryPath, "thys")
}

type manifest struct {
	Name             string
	ProvidesSessions *set.Set[string]
	RequiresSessions *set.Set[string]
}

func (a *AFPIndexer) resolveManifest(thy string, builtinSessions []string) (*manifest, error) {
	logging.Verbose("parsing theory package %s", thy)

	pkgManifest := &manifest{
		Name:             thy,
		ProvidesSessions: set.New[string](0),
		RequiresSessions: set.New[string](0),
	}

	rootFile, err := os.ReadFile(filepath.Join(a.theoriesPath(), thy, "ROOT"))
	if err != nil {
		return nil, errors.Join(err, ErrCannotReadROOT)
	}

	// resolving the sessions provided by this package is easy, we just
	// need to parse the ROOT file and take all session definitions
	parsed, err := parser.ParseRootFile(bytes.NewReader(rootFile))
	if err != nil {
		return nil, errors.Join(err, ErrCannotParseROOT)
	}

	// most - if not all - AFP packages specify their sessions within the "AFP" chapter, but just in case
	// I am going to check all the chapters
	for _, chapter := range parsed.Chapters {
		for _, session := range chapter.Sessions {
			pkgManifest.ProvidesSessions.Insert(session.Name)

			// required sessions are those that are explicitly mentioned within the ROOT file (sessions block)
			// it is also possible to import sessions using the theory files, but I'm not sure if those also
			// need to be mentioned in the ROOT - for now I am assuming they do (so I don't have to parse all the .thy files)
			pkgManifest.RequiresSessions.InsertSlice(session.Sessions)
		}
	}

	// remove any Isabelle builtin sessions
	pkgManifest.RequiresSessions.RemoveSlice(builtinSessions)
	// remove any sessions that are provided by this package
	pkgManifest.RequiresSessions.RemoveSet(pkgManifest.ProvidesSessions)

	logging.Verbose("theory package %s - provides %d, requires %d", thy, pkgManifest.ProvidesSessions.Size(), pkgManifest.RequiresSessions.Size())

	return pkgManifest, nil
}

func (a *AFPIndexer) prepareIndexRepo() error {
	err := git.SetupRemote(a.indexRepositoryUrl, a.afpDirectoryPath)
	if err != nil {
		return errors.Join(err, ErrCannotSetupIndexRepo)
	}

	// create a new branch for this AFP version
	err = git.MakeBranch(a.theoriesPath(), a.afpVersion)
	if err != nil {
		return errors.Join(err, ErrCannotMakeBranch)
	}

	logging.Unquiet("cloned and prepared index repository branch successfully")

	return nil
}

func (a *AFPIndexer) Index() error {
	// find all theory packages within the AFP by parsing the ROOTS file
	contents, err := os.ReadFile(filepath.Join(a.theoriesPath(), "ROOTS"))
	if err != nil {
		return errors.Join(err, ErrCannotReadROOTS)
	}

	builtinSessions, err := isabelle.FetchBuiltinSessions()
	if err != nil {
		return errors.Join(err, ErrCannotResolveBuiltinSessions)
	}

	manifests := make(map[string]*manifest)

	// parse the ROOT file for all the packages to resolve required and provided sessions
	rawTheoryPackages := strings.Split(string(contents), "\n")
	theoryPackages := make([]string, 0)
	for _, rawPkg := range rawTheoryPackages {
		trimmed := strings.TrimSpace(rawPkg)
		if trimmed == "" {
			continue
		}

		theoryPackages = append(theoryPackages, trimmed)
	}

	var pbar *progressbar.ProgressBar
	if internal.LogLevel == internal.LogLvlQuiet {
		pbar = progressbar.DefaultSilent(int64(len(rawTheoryPackages)))
	} else {
		pbar = progressbar.Default(int64(len(theoryPackages)), "parsing")
	}

	for _, pkg := range theoryPackages {
		if strings.TrimSpace(pkg) == "" {
			continue
		}

		m, err := a.resolveManifest(strings.TrimSpace(pkg), builtinSessions)
		if err != nil {
			return err
		}

		manifests[m.Name] = m

		_ = pbar.Add(1)
	}

	// create a map to allow us to resolve the package that provides a given session
	sessionsToPackage := make(map[string]string)
	for k, v := range manifests {
		for session := range v.ProvidesSessions.Items() {
			sessionsToPackage[session] = k
		}
	}

	pbar.Reset()
	pbar.Describe("resolving")

	// resolve which packages are required by each package
	packageRequires := make(map[string]*set.Set[string])
	for name, m := range manifests {
		packageRequires[name] = set.New[string](0)

		for s := range m.RequiresSessions.Items() {
			elem, ok := sessionsToPackage[s]
			if !ok {
				return errors.Join(ErrUnknownSession, errors.New(s))
			}

			packageRequires[name].Insert(elem)
		}

		_ = pbar.Add(1)
	}

	pbar.Reset()
	pbar.Describe("writing")

	// copy the files from the AFP directory into the index repo for each of the packages
	for pkgName := range manifests {
		logging.Verbose("creating proofman config file for %s", pkgName)

		// create a proofman.toml file for this package
		requiresPkgs := make([]string, 0)
		if reqs, ok := packageRequires[pkgName]; ok {
			for req := range reqs.Items() {
				requiresPkgs = append(requiresPkgs, req+" @ "+a.afpVersion)
			}
		}

		marshalled, err := toml.Marshal(config.ProofmanConfig{
			Project: config.Project{
				Name:        pkgName,
				Description: pkgName + " from the Archive of Formal Proofs",
				Version:     a.afpVersion,
				Requires:    requiresPkgs,
			},
		})
		if err != nil {
			return errors.Join(err, ErrCannotPopulateRepo)
		}

		err = os.WriteFile(
			filepath.Join(a.theoriesPath(), pkgName, internal.ConfigFileName),
			marshalled,
			0666,
		)
		if err != nil {
			return errors.Join(err, ErrCannotPopulateRepo)
		}

		_ = pbar.Add(1)
	}

	// push the changes to the upstream
	err = a.prepareIndexRepo()
	if err != nil {
		return err
	}

	logging.Unquiet("committing and pushing changes")
	if err = git.AddAll(a.afpDirectoryPath); err != nil {
		return err
	}
	if err = git.Commit(a.afpDirectoryPath, fmt.Sprintf("[proofman] auto index AFP (version %s)", a.afpVersion)); err != nil {
		return err
	}
	if err = git.Push(a.afpDirectoryPath); err != nil {
		return err
	}
	logging.Quiet("indexing complete")

	return nil
}
