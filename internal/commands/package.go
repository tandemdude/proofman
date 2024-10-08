package commands

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/hashicorp/go-set/v3"
	"github.com/proofman-dev/commons/constants"
	genproto "github.com/proofman-dev/commons/protos/generated"
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/logging"
	"github.com/tandemdude/proofman/pkg/config"
	"github.com/tandemdude/proofman/pkg/isabelle"
	"github.com/tandemdude/proofman/pkg/parser"
	"github.com/tandemdude/proofman/pkg/proofbank"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func resolveUnknownVersion() string {
	// Mercurial:
	// hg log -l 1 -I "glob:**" --template "{date|isodate}\n"
	// Output: '2023-07-14 17:26 +0200'
	logging.Verbose("trying to resolve package version using mercurial")
	cmd := exec.Command("hg", "log", "-l", "1", "-I", "glob:**", "--template", "{date|isodate}\n")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out))[0:10]
	}

	// Git:
	// git log -1 --format=%cd --date=short -- .
	// Output: '2024-04-28'
	logging.Verbose("trying to resolve package version using git")
	cmd = exec.Command("git", "log", "-1", "--format=%cd", "--date=short", "--", ".")
	out, err = cmd.CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	// fall back to today's date
	logging.Verbose("falling back to today's date as the package version")
	return time.Now().Format(time.DateOnly)
}

type resolvedManifest struct {
	manifest *genproto.PackageManifest
	includes []string
	excludes []string
}

func manifestFromScratch(pwd string) (*resolvedManifest, error) {
	// indexing a project that already exists - probably safe to assume
	// that the name will be the same as the current working directory
	dirName := filepath.Base(pwd)

	// parse the ROOT file to figure out what sessions are provided
	rootContent, err := os.ReadFile("ROOT")
	if err != nil {
		return nil, err
	}
	rootStructure, err := parser.ParseRootFile(bytes.NewReader(rootContent))
	if err != nil {
		return nil, err
	}

	manifest := &genproto.PackageManifest{
		Name:             dirName,
		Version:          resolveUnknownVersion(),
		ProvidesSessions: make([]string, 0),
	}

	for _, chapter := range rootStructure.Chapters {
		for _, session := range chapter.Sessions {
			manifest.ProvidesSessions = append(manifest.ProvidesSessions, session.Name)
		}
	}

	// Requirements for this package must be inferred from the session definitions in the
	// ROOT file - an indexing implementation of the AFP should make sure that the order
	// the different theories are indexed in is correct so that the correct packages for each
	// theory are available from the package index
	//
	// Because we don't know anything about required theory versions for what we are currently
	// trying to package, this script assumes that it requires the latest version of all dependencies
	//
	// IF you are indexing multiple versions of the AFP, it would be recommended to start with
	// the oldest version and work forward in time, so that each theory has the correct dependency
	// requirements
	rawRequiresSessions := make([]string, 0)

	// A required session is either the parent to a session, or a session mentioned within the
	// 'sessions' block within one of the defined sessions. We can later query the package index
	// to ask which dependencies provide the sessions that we need.
	for _, chapter := range rootStructure.Chapters {
		for _, session := range chapter.Sessions {
			if len(session.SystemName) > 0 {
				rawRequiresSessions = append(rawRequiresSessions, session.SystemName)
			}

			rawRequiresSessions = append(rawRequiresSessions, session.Sessions...)
		}
	}

	// Remove any builtin sessions (HOL, Main, etc...) and any sessions that are provided by this
	// ROOT file from the requirements.
	// TODO - move this to the server?
	builtinSessions, err := isabelle.FetchBuiltinSessions()
	if err != nil {
		return nil, err
	}

	finalRequiresSessions := make([]string, 0)
	for _, req := range rawRequiresSessions {
		// If the session is provided by this ROOT file
		if slices.Contains(manifest.ProvidesSessions, req) {
			continue
		}

		// If the sessions is an Isabelle builtin
		if slices.Contains(builtinSessions, req) {
			continue
		}

		// This is an actual requirement
		finalRequiresSessions = append(finalRequiresSessions, req)
	}

	// We need to resolve the versions and actual package names for the sessions that are required
	// by making a request to the package index
	client, err := proofbank.NewUnauthenticatedClient(internal.ProofbankBaseUrl)
	if err != nil {
		return nil, err
	}

	resp, err := client.QueryPackagesForSessions(finalRequiresSessions)
	if err != nil {
		return nil, err
	}

	// Store package name and version in a map while we figure out which packages we need
	// We can normalise them to version strings later
	rawRequiresPackages := make(map[string]*set.Set[string])

	for _, session := range rawRequiresSessions {
		data, ok := resp.SessionsResult[session]
		if !ok {
			// TODO - the package index doesn't know about this session - fatal error
			return nil, errors.New("session " + session + " not provided by any package")
		}

		var packageVersions *set.Set[string]
		packageVersions, ok = rawRequiresPackages[data.Name]
		if !ok {
			packageVersions = set.From(data.Versions)
			rawRequiresPackages[data.Name] = packageVersions
			continue
		}

		// remove version elements from the set where this session isn't available
		packageVersions.RemoveSet(packageVersions.Difference(set.From(data.Versions)))
		// if packageVersions is now empty then this dependency is not satisfiable
		if packageVersions.Size() == 0 {
			// TODO - better
			return nil, errors.New("package " + data.Name + " not satisfiable due to version conflicts")
		}
	}

	requiresPackages := make([]string, 0)
	for packageName, availableVersions := range rawRequiresPackages {
		largest := ""
		for item := range availableVersions.Items() {
			if item > largest {
				largest = item
			}
		}
		requiresPackages = append(requiresPackages, fmt.Sprintf("%s @ %s", packageName, largest))
	}

	manifest.DependsOn = requiresPackages

	return &resolvedManifest{
		manifest: manifest,
		includes: []string{},
		excludes: []string{},
	}, nil
}

func manifestFromConfig(pwd string) (*resolvedManifest, error) {
	// configuration file exists, we can use it to provide the dependencies
	cfg, err := config.FromFile(pwd)
	if err != nil {
		return nil, err
	}

	// parse the ROOT file to figure out what sessions are provided
	rootContent, err := os.ReadFile("ROOT")
	if err != nil {
		return nil, err
	}
	rootStructure, err := parser.ParseRootFile(bytes.NewReader(rootContent))
	if err != nil {
		return nil, err
	}

	manifest := &genproto.PackageManifest{
		Name:             cfg.Project.Name,
		Version:          cfg.Project.Version,
		ProvidesSessions: make([]string, 0),
		DependsOn:        make([]string, 0),
	}

	for _, chapter := range rootStructure.Chapters {
		for _, session := range chapter.Sessions {
			manifest.ProvidesSessions = append(manifest.ProvidesSessions, session.Name)
		}
	}

	manifest.DependsOn = cfg.Project.Requires

	return &resolvedManifest{
		manifest: manifest,
		includes: cfg.Package.Include,
		excludes: cfg.Package.Exclude,
	}, nil
}

func package_(_ *cli.Context) error {
	// look for an existing configuration file
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	existing, err := config.FileExistsIn(pwd)

	var manifestInfo *resolvedManifest

	if err != nil || !existing {
		manifestInfo, err = manifestFromScratch(pwd)
	} else {
		manifestInfo, err = manifestFromConfig(pwd)
	}

	if err != nil {
		return fmt.Errorf("failed extracting package metadata: %s", err)
	}

	// we have the package metadata and the list of excludes, now we need to pack
	// the project into a zip file so that it can be distributed
	fname := fmt.Sprintf("%s.%s.ipkg", manifestInfo.manifest.Name, manifestInfo.manifest.Version)
	pkgFile, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("failed creating package file: %s", err)
	}
	defer pkgFile.Close()

	writer := zip.NewWriter(pkgFile)
	defer writer.Close()

	walker := func(path string, info os.DirEntry, err error) error {
		logging.Verbose("processing item '%s'", path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// excludes take preference over includes
		for _, exclude := range manifestInfo.excludes {
			matches, err := doublestar.PathMatch(exclude, path)
			if err != nil {
				return err
			}

			if matches {
				logging.Verbose("item '%s' explicitly excluded - matches expression '%s'", path, exclude)
				return nil
			}
		}

		// check if this file should be explicitly or implicitly included
		matchesDefaultInclude, _ := doublestar.Match("*.{thy,ML}", info.Name())
		if info.Name() != "ROOT" &&
			info.Name() != "README.md" &&
			info.Name() != internal.ConfigFileName &&
			!matchesDefaultInclude {
			matchesAnyInclude := false
			for _, include := range manifestInfo.includes {
				matches, err := doublestar.PathMatch(include, path)
				if err != nil {
					return err
				}

				if matches {
					matchesAnyInclude = true
					logging.Verbose("item '%s' explicitly included - matches expression '%s'", path, include)
					break
				}
			}

			if !matchesAnyInclude {
				return nil
			}
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := writer.Create(filepath.Join(manifestInfo.manifest.Name, path))
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		logging.Unquiet("wrote file '%s' to archive", path)
		return nil
	}

	err = filepath.WalkDir(".", walker)
	if err != nil {
		return err
	}

	// write package metadata file
	f, err := writer.Create(constants.ManifestFileName)
	if err != nil {
		return err
	}

	dumped, err := protojson.Marshal(manifestInfo.manifest)
	if err != nil {
		return err
	}

	_, err = f.Write(dumped)
	if err != nil {
		return err
	}
	logging.Unquiet("wrote file 'manifest.ipkg.json' to archive")
	logging.Unquiet("packaging complete - created file '%s'", fname)

	return nil
}

var PackageCommand = &cli.Command{
	Name:   "package",
	Usage:  "Packages the current project to be suitable for upload to ProofBank",
	Action: package_,
}
