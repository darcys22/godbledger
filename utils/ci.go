// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// +build none

/*
The ci command is called from Continuous Integration scripts.

Usage: go run utils/ci.go <command> <command flags/arguments>

Available commands are:

   install    [ -arch architecture ] [ -cc compiler ] [ packages... ]                          -- builds packages and executables
   test       [ -coverage ] [ packages... ]                                                    -- runs the tests
   lint                                                                                        -- runs certain pre-selected linters
   archive                                                                                     -- creates github release
   importkeys                                                                                  -- imports signing keys from env
   debsrc     [ -signer key-id ] [ -upload dest ]                                              -- creates a debian source package
   xgo        [ -alltools ] [ options ]                                                        -- cross builds according to options

For all commands, -n prevents execution of external programs (dry run mode).

*/
package main

import (
	//"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	//"regexp"
	"runtime"
	"strings"

	"github.com/cespare/cp"
	"github.com/darcys22/godbledger/godbledger/version"
	"github.com/darcys22/godbledger/internal/build"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"time"
)

var (
	// Make Release iterates through these executables
	packagesToBuild = []string{
		"godbledger",
		"ledger-cli",
		"reporter",
	}

	// A debian package is created for all executables listed here.
	debExecutables = []debExecutable{
		{
			PackageName: "godbledger-core",
			BinaryName:  "godbledger",
			Description: "Accounting server to manage financial databases and record double entry bookkeeping transactions",
		},
		{
			BinaryName:  "ledger-cli",
			Description: "Godbledger grpc client",
		},
		{
			BinaryName:  "reporter",
			Description: "basic reporting queries on SQL databases associated with godbledger servers",
		},
	}

	// A debian package is created for all executables listed here.
	debGoDBLedger = debPackage{
		Name:        "godbledger",
		Version:     version.Version,
		Executables: debExecutables,
	}

	// Debian meta packages to build and push to Ubuntu PPA
	debPackages = []debPackage{
		debGoDBLedger,
	}

	// Distros for which packages are created.
	debDistroGoBoots = map[string]string{
		"xenial":  "golang-go",
		"bionic":  "golang-go",
		"focal":   "golang-go",
		"groovy":  "golang-go",
		"hirsute": "golang-go",
	}

	debGoBootPaths = map[string]string{
		"golang-go": "/usr/lib/go",
	}

	// This is the version of go that will be downloaded by
	//
	//     go run ci.go install -dlgo
	dlgoVersion = "1.16"
)

var GOBIN, _ = filepath.Abs(filepath.Join("build", "bin"))
var BUILDDIR, _ = filepath.Abs("build")

func archBinPath(goos string, arch string) string {
	if goos == runtime.GOOS && arch == runtime.GOARCH {
		return filepath.Join(GOBIN, "native")
	}
	return filepath.Join(GOBIN, fmt.Sprintf("%s-%s", goos, arch))
}

func cachePath() string {
	return filepath.Join(BUILDDIR, ".cache")
}

func distPath() string {
	return filepath.Join(BUILDDIR, "dist")
}

func executablePath(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(GOBIN, name)
}

func main() {
	log.SetFlags(log.Lshortfile)

	if _, err := os.Stat(filepath.Join("utils", "ci.go")); os.IsNotExist(err) {
		log.Fatal("this script must be run from the root of the repository")
	}
	if len(os.Args) < 2 {
		log.Fatal("need subcommand as first argument")
	}
	switch os.Args[1] {
	case "info":
		log.Printf("current system: %s/%s", runtime.GOOS, runtime.GOARCH)
	case "build":
		doBuild(os.Args[2:])
	case "test":
		doTest(os.Args[2:])
	case "lint":
		doLint(os.Args[2:])
	case "archive":
		doArchive(os.Args[2:])
	case "debsrc":
		doDebianSource(os.Args[2:])
	case "xgo":
		doXgo(os.Args[2:])
	default:
		log.Fatal("unknown command ", os.Args[1])
	}
}

// Compiling

// ensureMinimumGoVersion ensures that the current go version is compatible with
// our build requirements. People regularly open issues about compilation failure
// with outdated Go; this should save them the trouble.
func ensureMinimumGoVersion() {
	if !strings.Contains(runtime.Version(), "devel") {
		// Figure out the minor version number since we can't textually compare (1.10 < 1.9)
		var minor int
		fmt.Sscanf(strings.TrimPrefix(runtime.Version(), "go1."), "%d", &minor)

		if minor < 11 {
			log.Println("You have Go version", runtime.Version())
			log.Println("goDBLedger requires at least Go version 1.14 and cannot")
			log.Println("be compiled with an earlier version. Please upgrade your Go installation.")
			os.Exit(1)
		}
	}
}

func doBuild(cmdline []string) {
	var (
		goos = flag.String("os", runtime.GOOS, "OS to (cross) build for")
		arch = flag.String("arch", runtime.GOARCH, "Architecture to (cross) build for")
		cc   = flag.String("cc", "", "C compiler to (cross) build with")
	)
	flag.CommandLine.Parse(cmdline)
	env := build.Env()
	log.Printf("build targeting: %s/%s", *goos, *arch)

	// Compile packages given as arguments, or everything if there are no arguments.
	packages := []string{"./..."}
	if flag.NArg() > 0 {
		packages = flag.Args()
	}

	ensureMinimumGoVersion()

	// ensure our output path exists so we can use the -o flag to dump build output there
	os.MkdirAll(archBinPath(*goos, *arch), os.ModePerm)

	// native build can be done with plain go tools
	if *goos == runtime.GOOS && *arch == runtime.GOARCH {
		goinstall := goTool("build", buildFlags(env)...)
		if runtime.GOARCH == "arm64" {
			goinstall.Args = append(goinstall.Args, "-p", "1")
		}
		goinstall.Args = append(goinstall.Args, []string{"-o", archBinPath(*goos, *arch)}...)
		goinstall.Args = append(goinstall.Args, "-v")
		goinstall.Args = append(goinstall.Args, packages...)
		build.MustRun(goinstall)
		return
	}

	// Seems we are cross compiling, work around forbidden GOBIN
	goinstall := goToolArch(*arch, *cc, "install", buildFlags(env)...)
	goinstall.Args = append(goinstall.Args, "-v")
	goinstall.Args = append(goinstall.Args, []string{"-buildmode", "archive"}...)
	goinstall.Args = append(goinstall.Args, packages...)
	build.MustRun(goinstall)

	if cmds, err := ioutil.ReadDir("cmd"); err == nil {
		for _, cmd := range cmds {
			pkgs, err := parser.ParseDir(token.NewFileSet(), filepath.Join(".", "cmd", cmd.Name()), nil, parser.PackageClauseOnly)
			if err != nil {
				log.Fatal(err)
			}
			for name := range pkgs {
				if name == "main" {
					gobuild := goToolArch(*arch, *cc, "build", buildFlags(env)...)
					gobuild.Args = append(gobuild.Args, "-v")
					gobuild.Args = append(gobuild.Args, []string{"-o", executablePath(cmd.Name())}...)
					gobuild.Args = append(gobuild.Args, "."+string(filepath.Separator)+filepath.Join("cmd", cmd.Name()))
					build.MustRun(gobuild)
					break
				}
			}
		}
	}
}

func buildFlags(env build.Environment) (flags []string) {
	var ld []string
	if env.Commit != "" {
		ld = append(ld, "-X", "github.com/darcys22/godbledger/godbledger/version.gitCommit="+env.Commit)
		ld = append(ld, "-X", "github.com/darcys22/godbledger/godbledger/version.gitDate="+env.Date)
		ld = append(ld, "-X", "github.com/darcys22/godbledger/godbledger/version.gitBranch="+env.Branch)
	}
	if runtime.GOOS == "darwin" {
		ld = append(ld, "-s")
	}

	if len(ld) > 0 {
		flags = append(flags, "-ldflags", strings.Join(ld, " "))
	}
	return flags
}

func goTool(subcmd string, args ...string) *exec.Cmd {
	return goToolArch(runtime.GOARCH, os.Getenv("CC"), subcmd, args...)
}

func goToolArch(arch string, cc string, subcmd string, args ...string) *exec.Cmd {
	cmd := build.GoTool(subcmd, args...)
	if arch == "" || arch == runtime.GOARCH {
		cmd.Env = append(cmd.Env, "GOBIN="+GOBIN)
	} else {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
		cmd.Env = append(cmd.Env, "GOARCH="+arch)
	}
	if cc != "" {
		cmd.Env = append(cmd.Env, "CC="+cc)
	}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOBIN=") {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
	return cmd
}

// Running The Tests
//
// "tests" also includes static analysis tools such as vet.

func doTest(cmdline []string) {
	coverage := flag.Bool("coverage", false, "Whether to record code coverage")
	verbose := flag.Bool("v", false, "Whether to log verbosely")
	integration := flag.Bool("integration", false, "Whether to run integration tests")
	mysql := flag.Bool("mysql", false, "Whether to run mysql integration tests (not currently supported)")
	secure := flag.Bool("secure", false, "Whether to run secure integration tests")
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}

	// Run the actual tests.
	// Test a single package at a time. CI builders are slow
	// and some tests run into timeouts under load.
	gotest := goTool("test", buildFlags(env)...)
	gotest.Args = append(gotest.Args, "-p", "1")
	if *coverage {
		gotest.Args = append(gotest.Args, "-covermode=atomic", "-cover")
	}
	if *integration {
		gotest.Args = append(gotest.Args, "-tags=integration")
	}
	if *mysql {
		gotest.Args = append(gotest.Args, "-tags=integration,mysql")
	}
	if *secure {
		gotest.Args = append(gotest.Args, "-tags=integration,secure")
	}
	if *verbose {
		gotest.Args = append(gotest.Args, "-v")
	}

	gotest.Args = append(gotest.Args, packages...)
	build.MustRun(gotest)
}

// doLint runs golangci-lint on requested packages.
func doLint(cmdline []string) {
	var (
		cachedir = flag.String("cachedir", cachePath(), "directory for caching golangci-lint binary.")
	)
	flag.CommandLine.Parse(cmdline)
	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}

	linter := downloadLinter(*cachedir)
	lflags := []string{"run", "--config", ".golangci.yml"}
	build.MustRunCommand(linter, append(lflags, packages...)...)
	fmt.Println("You have achieved perfection.")
}

//downloadLinter downloads and unpacks golangci-lint.
func downloadLinter(cachedir string) string {
	const version = "1.27.0"

	csdb := build.MustLoadChecksums("utils/checksums.txt")
	base := fmt.Sprintf("golangci-lint-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/golangci/golangci-lint/releases/download/v%s/%s.tar.gz", version, base)
	archivePath := filepath.Join(cachedir, base+".tar.gz")
	if err := csdb.DownloadFile(url, archivePath); err != nil {
		log.Fatal(err)
	}
	if err := build.ExtractArchive(archivePath, cachedir); err != nil {
		log.Fatal(err)
	}
	return filepath.Join(cachedir, base, "golangci-lint")
}

// Release Packaging
func doArchive(cmdline []string) {
	baseURLStr := "https://api.github.com/"

	var (
		owner      = flag.String("owner", "darcys22", `github user who is owner of the repo`)
		repo       = flag.String("repo", "godbledger", `github repo to upload releases to`)
		path       = flag.String("path", "build/dist/", `folder containing assets to upload`)
		tag        = flag.String("tag", version.Version, `version of assets being uploaded`)
		name       = flag.String("name", fmt.Sprintf("v%s", version.Version), `name of the release`)
		commitish  = flag.String("commitish", "", `commit hash`)
		draft      = flag.Bool("draft", true, `whether to create the release as a draft`)
		prerelease = flag.Bool("prerelease", false, `whether to create the release as a prerelease`)
		soft       = flag.Bool("soft", false, `fail if the tag already exists`)
		replace    = flag.Bool("replace", false, `whether the upload will replace all assets on an already existing release`)
		recreate   = flag.Bool("recrease", false, `whether the upload will enforce the replacement of an already existing release`)
		parallel   = flag.Int("parallel", 1, `uploads the designated assets in parallel`)
	)
	flag.CommandLine.Parse(cmdline)

	log.Printf("Name: %s", *name)
	if len(*commitish) == 0 {
		env := build.Env()
		*commitish = env.Commit
	}
	log.Printf("Commit: %s", *commitish)

	localAssets, err := build.LocalAssets(*path)
	if err != nil {
		log.Fatalf("Failed to find assets from %s: %s\n", path, err)
	}

	log.Printf("Number of file to upload: %d", len(localAssets))

	//Create the checksums
	checksums, err := build.SHA256Assets(localAssets)
	var b bytes.Buffer
	b.WriteString("**sha256sum**\n\n")
	for i, localAsset := range localAssets {
		fmt.Fprintf(&b, "%s %s\n", checksums[i], filepath.Base(localAsset))
	}
	body := b.String()

	// Create a GitHub client
	token := os.Getenv("GH_ACCESS_TOKEN")
	if len(token) == 0 {
		log.Fatal("Failed to get GitHub access token")
	}
	gitHubClient, err := build.NewGitHubClient(*owner, *repo, token, baseURLStr)
	if err != nil {
		log.Fatalf("Failed to construct GitHub client: %s\n", err)
	}

	ghr := GHR{
		GitHub: gitHubClient,
	}

	// Prepare create release request
	req := &github.RepositoryRelease{
		Name:            github.String(*name),
		TagName:         github.String(*tag),
		Prerelease:      github.Bool(*prerelease),
		Draft:           github.Bool(*draft),
		TargetCommitish: github.String(*commitish),
		Body:            github.String(body),
	}

	ctx := context.TODO()

	if *soft {
		_, err := ghr.GitHub.GetRelease(ctx, *req.TagName)

		if err == nil {
			log.Fatalf("ghr aborted since tag `%s` already exists\n", *req.TagName)
		}

		if err != nil {
			log.Fatalf("Failed to get GitHub release: %s\n", err)
		}
	}

	release, err := ghr.CreateRelease(ctx, req, *recreate)
	if err != nil {
		log.Fatalf("Failed to create GitHub release page: %s\n", err)
	}

	if *replace {
		err := ghr.DeleteAssets(ctx, *release.ID, localAssets, *parallel)
		if err != nil {
			log.Fatalf("Failed to delete existing assets: %s\n", err)
		}
	}

	err = ghr.UploadAssets(ctx, *release.ID, localAssets, *parallel)
	if err != nil {
		log.Fatalf("Failed to upload one of assets: %s\n", err)
	}

	if !*draft {
		_, err := ghr.GitHub.EditRelease(ctx, *release.ID, &github.RepositoryRelease{
			Draft: github.Bool(false),
		})
		if err != nil {
			log.Fatalf("Failed to publish release: %s\n", err)
		}
	}
}

// GHR contains the top level GitHub object
// https://github.com/tcnksm/ghr/blob/master/ghr.go
type GHR struct {
	GitHub build.GitHub
}

// CreateRelease creates (or recreates) a new package release
func (g *GHR) CreateRelease(ctx context.Context, req *github.RepositoryRelease, recreate bool) (*github.RepositoryRelease, error) {

	// When draft release creation is requested,
	// create it without any check (it can).
	if *req.Draft {
		log.Printf("Create a draft release")
		return g.GitHub.CreateRelease(ctx, req)
	}

	// Always create release as draft first. After uploading assets, turn off
	// draft unless the `-draft` flag is explicitly specified.
	// It is to prevent users from seeing empty release.
	req.Draft = github.Bool(true)

	// Check release exists.
	// If release is not found, then create a new release.
	release, err := g.GitHub.GetRelease(ctx, *req.TagName)
	if err != nil {
		if err != build.ErrReleaseNotFound {
			return nil, errors.Wrap(err, "failed to get release")
		}
		log.Printf("Release (with tag %s) not found: create a new one",
			*req.TagName)

		if recreate {
			log.Printf("WARNING: '-recreate' is specified but release (%s) not found", *req.TagName)
		}

		log.Println("==> Create a new release")
		return g.GitHub.CreateRelease(ctx, req)
	}

	// recreate is not true. Then use that existing release.
	if !recreate {
		log.Printf("Release (with tag %s) exists: use existing one",
			*req.TagName)

		log.Printf("found release (%s). Use existing one.\n",
			*req.TagName)
		return release, nil
	}

	// When recreate is requested, delete existing release and create a
	// new release.
	log.Printf("Recreate a release")
	if err := g.DeleteRelease(ctx, *release.ID, *req.TagName); err != nil {
		return nil, err
	}

	return g.GitHub.CreateRelease(ctx, req)
}

// DeleteRelease removes an existing release, if it exists. If it does not exist,
// DeleteRelease returns an error
func (g *GHR) DeleteRelease(ctx context.Context, ID int64, tag string) error {

	err := g.GitHub.DeleteRelease(ctx, ID)
	if err != nil {
		return err
	}

	err = g.GitHub.DeleteTag(ctx, tag)
	if err != nil {
		return err
	}

	// This is because sometimes the process of creating a release on GitHub
	// is faster than deleting a tag.
	time.Sleep(5 * time.Second)

	return nil
}

// UploadAssets uploads the designated assets in parallel (determined by parallelism setting)
func (g *GHR) UploadAssets(ctx context.Context, releaseID int64, localAssets []string, parallel int) error {
	start := time.Now()
	defer func() {
		log.Printf("UploadAssets: time: %d ms", int(time.Since(start).Seconds()*1000))
	}()

	eg, ctx := errgroup.WithContext(ctx)
	semaphore := make(chan struct{}, parallel)
	for _, localAsset := range localAssets {
		localAsset := localAsset
		eg.Go(func() error {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			log.Printf("Uploading: %15s\n", filepath.Base(localAsset))
			_, err := g.GitHub.UploadAsset(ctx, releaseID, localAsset)
			if err != nil {
				return errors.Wrapf(err,
					"failed to upload asset: %s", localAsset)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "one of the goroutines failed")
	}

	return nil
}

// DeleteAssets removes uploaded assets for a given release
func (g *GHR) DeleteAssets(ctx context.Context, releaseID int64, localAssets []string, parallel int) error {
	start := time.Now()
	defer func() {
		log.Printf("DeleteAssets: time: %d ms", int(time.Since(start).Seconds()*1000))
	}()

	eg, ctx := errgroup.WithContext(ctx)

	assets, err := g.GitHub.ListAssets(ctx, releaseID)
	if err != nil {
		return errors.Wrap(err, "failed to list assets")
	}

	semaphore := make(chan struct{}, parallel)
	for _, localAsset := range localAssets {
		for _, asset := range assets {
			// https://golang.org/doc/faq#closures_and_goroutines
			localAsset, asset := localAsset, asset

			// Uploaded asset name is same as basename of local file
			if *asset.Name == filepath.Base(localAsset) {
				eg.Go(func() error {
					semaphore <- struct{}{}
					defer func() {
						<-semaphore
					}()

					log.Printf("Deleting: %15s\n", *asset.Name)
					if err := g.GitHub.DeleteAsset(ctx, *asset.ID); err != nil {
						return errors.Wrapf(err,
							"failed to delete asset: %s", *asset.Name)
					}
					return nil
				})
			}
		}
	}

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "one of the goroutines failed")
	}

	return nil
}

// skips archiving for some build configurations.
func maybeSkipArchive(env build.Environment) {
	if env.IsPullRequest {
		log.Printf("skipping because this is a PR build")
		os.Exit(0)
	}
	if env.IsCronJob {
		log.Printf("skipping because this is a cron job")
		os.Exit(0)
	}
	//if env.Branch != "master" && !strings.HasPrefix(env.Tag, "v1.") {
	//log.Printf("skipping because branch %q, tag %q is not on the whitelist", env.Branch, env.Tag)
	//os.Exit(0)
	//}
}

// Debian Packaging
func doDebianSource(cmdline []string) {
	var (
		cachedir = flag.String("cachedir", "./build/cache", `Filesystem path to cache the downloaded Go bundles at`)
		signer   = flag.String("signer", "", `Signing key name, also used as package author`)
		upload   = flag.String("upload", "", `Where to upload the source package (usually "darcys22/godbledger")`)
		sshUser  = flag.String("sftp-user", "", `Username for SFTP upload (usually "darcys22")`)
		workdir  = flag.String("workdir", "", `Output directory for packages (uses temp dir if unset)`)
		now      = time.Now()
	)
	flag.CommandLine.Parse(cmdline)
	*workdir = makeWorkdir(*workdir)
	env := build.Env()
	maybeSkipArchive(env)

	// Import the signing key.
	//gpg --export-secret-key sean@darcyfinancial.com  | base64 | paste -s -d '' > secret-key-base64-encoded.gpg
	if key := getenvBase64("PPA_SIGNING_KEY"); len(key) > 0 {
		gpg := exec.Command("gpg", "--import", "--no-tty", "--batch", "--yes")
		gpg.Stdin = bytes.NewReader(key)
		build.MustRun(gpg)
	}

	// Download and verify the Go source package.
	gobundle := downloadGoSources(*cachedir)

	// Download all the dependencies needed to build the sources and run the ci script
	srcdepfetch := goTool("mod", "download")
	gopath, _ := filepath.Abs(filepath.Join(*workdir, "modgopath"))
	srcdepfetch.Env = append(os.Environ(), "GOPATH="+gopath)
	build.MustRun(srcdepfetch)

	cidepfetch := goTool("run", "./build/ci.go")
	cidepfetch.Env = append(os.Environ(), "GOPATH="+filepath.Join(*workdir, "modgopath"))
	cidepfetch.Run() // Command fails, don't care, we only need the deps to start it

	// Create Debian packages and upload them.
	for _, pkg := range debPackages {
		for distro, goboot := range debDistroGoBoots {
			// Prepare the debian package with the go-ethereum sources.
			meta := newDebMetadata(distro, goboot, *signer, env, now, pkg.Name, pkg.Version, pkg.Executables)
			fmt.Println("Building debian package in: " + *workdir)
			pkgdir := stageDebianSource(*workdir, meta)

			// Add Go source code
			if err := build.ExtractArchive(gobundle, pkgdir); err != nil {
				log.Fatalf("Failed to extract Go sources: %v", err)
			}
			if err := os.Rename(filepath.Join(pkgdir, "go"), filepath.Join(pkgdir, ".go")); err != nil {
				log.Fatalf("Failed to rename Go source folder: %v", err)
			}
			// Add all dependency modules in compressed form
			os.MkdirAll(filepath.Join(pkgdir, ".mod", "cache"), 0755)
			if err := cp.CopyAll(filepath.Join(pkgdir, ".mod", "cache", "download"), filepath.Join(*workdir, "modgopath", "pkg", "mod", "cache", "download")); err != nil {
				log.Fatalf("Failed to copy Go module dependencies: %v", err)
			}
			// Run the packaging and upload to the PPA
			debuild := exec.Command("debuild", "-S", "-sa", "-us", "-uc", "-d", "-Zxz", "-nc")
			debuild.Dir = pkgdir
			build.MustRun(debuild)

			var (
				basename  = fmt.Sprintf("%s_%s", meta.Name(), meta.VersionString())
				source    = filepath.Join(*workdir, basename+".tar.xz")
				dsc       = filepath.Join(*workdir, basename+".dsc")
				changes   = filepath.Join(*workdir, basename+"_source.changes")
				buildinfo = filepath.Join(*workdir, basename+"_source.buildinfo")
			)
			if *signer != "" {
				debsign := exec.Command("debsign", changes)
				build.MustRun(debsign)
			}
			if *upload != "" {
				ppaUpload(*workdir, *upload, *sshUser, []string{source, dsc, changes, buildinfo})
			}
		}
	}
}

// downloadGoSources downloads the Go source tarball.
func downloadGoSources(cachedir string) string {
	csdb := build.MustLoadChecksums("utils/checksums.txt")
	file := fmt.Sprintf("go%s.src.tar.gz", dlgoVersion)
	url := "https://dl.google.com/go/" + file
	dst := filepath.Join(cachedir, file)
	if err := csdb.DownloadFile(url, dst); err != nil {
		log.Fatal(err)
	}
	return dst
}

func ppaUpload(workdir, ppa, sshUser string, files []string) {
	p := strings.Split(ppa, "/")
	if len(p) != 2 {
		log.Fatal("-upload PPA name must contain single /")
	}
	if sshUser == "" {
		sshUser = p[0]
	}
	incomingDir := fmt.Sprintf("~%s/ubuntu/%s", p[0], p[1])
	// Create the SSH identity file if it doesn't exist.
	var idfile string
	if sshkey := getenvBase64("PPA_SSH_KEY"); len(sshkey) > 0 {
		idfile = filepath.Join(workdir, "sshkey")
		if _, err := os.Stat(idfile); os.IsNotExist(err) {
			ioutil.WriteFile(idfile, sshkey, 0600)
		}
	}
	// Upload
	dest := sshUser + "@ppa.launchpad.net"
	if err := build.UploadSFTP(idfile, dest, incomingDir, files); err != nil {
		log.Fatal(err)
	}
}

func getenvBase64(variable string) []byte {
	dec, err := base64.StdEncoding.DecodeString(os.Getenv(variable))
	if err != nil {
		log.Fatal("invalid base64 " + variable)
	}
	return []byte(dec)
}

func makeWorkdir(wdflag string) string {
	var err error
	if wdflag != "" {
		err = os.MkdirAll(wdflag, 0744)
	} else {
		wdflag, err = ioutil.TempDir("", "godbledger-build-")
	}
	if err != nil {
		log.Fatal(err)
	}
	return wdflag
}

func isUnstableBuild(env build.Environment) bool {
	if env.Tag != "" {
		return false
	}
	return true
}

type debPackage struct {
	Name        string          // the name of the Debian package to produce, e.g. "godbledger"
	Version     string          // the clean version of the debPackage, e.g. 1.8.12, without any metadata
	Executables []debExecutable // executables to be included in the package
}

type debMetadata struct {
	Env           build.Environment
	GoBootPackage string
	GoBootPath    string

	PackageName string

	// go-ethereum version being built. Note that this
	// is not the debian package version. The package version
	// is constructed by VersionString.
	Version string

	Author       string // "name <email>", also selects signing key
	Distro, Time string
	Executables  []debExecutable
}

type debExecutable struct {
	PackageName string
	BinaryName  string
	Description string
}

// Package returns the name of the package if present, or
// fallbacks to BinaryName
func (d debExecutable) Package() string {
	if d.PackageName != "" {
		return d.PackageName
	}
	return d.BinaryName
}

func newDebMetadata(distro, goboot, author string, env build.Environment, t time.Time, name string, version string, exes []debExecutable) debMetadata {
	if author == "" {
		// No signing key, use default author.
		author = "Sean Darcy <sean@darcyfinanical.com>"
	}
	return debMetadata{
		GoBootPackage: goboot,
		GoBootPath:    debGoBootPaths[goboot],
		PackageName:   name,
		Env:           env,
		Author:        author,
		Distro:        distro,
		Version:       version,
		Time:          t.Format(time.RFC1123Z),
		Executables:   exes,
	}
}

// Name returns the name of the metapackage that depends
// on all executable packages.
func (meta debMetadata) Name() string {
	if isUnstableBuild(meta.Env) {
		return meta.PackageName + "-unstable"
	}
	return meta.PackageName
}

// VersionString returns the debian version of the packages.
func (meta debMetadata) VersionString() string {
	vsn := meta.Version
	if meta.Env.Buildnum != "" {
		vsn += "+build" + meta.Env.Buildnum
	}
	if meta.Distro != "" {
		vsn += "+" + meta.Distro
	}
	return vsn
}

// ExeList returns the list of all executable packages.
func (meta debMetadata) ExeList() string {
	names := make([]string, len(meta.Executables))
	for i, e := range meta.Executables {
		names[i] = meta.ExeName(e)
	}
	return strings.Join(names, ", ")
}

// ExeName returns the package name of an executable package.
func (meta debMetadata) ExeName(exe debExecutable) string {
	if isUnstableBuild(meta.Env) {
		return exe.Package() + "-unstable"
	}
	return exe.Package()
}

// ExeConflicts returns the content of the Conflicts field
// for executable packages.
func (meta debMetadata) ExeConflicts(exe debExecutable) string {
	if isUnstableBuild(meta.Env) {
		// Set up the conflicts list so that the *-unstable packages
		// cannot be installed alongside the regular version.
		//
		// https://www.debian.org/doc/debian-policy/ch-relationships.html
		// is very explicit about Conflicts: and says that Breaks: should
		// be preferred and the conflicting files should be handled via
		// alternates. We might do this eventually but using a conflict is
		// easier now.
		return "godbledger, " + exe.Package()
	}
	return ""
}

func stageDebianSource(tmpdir string, meta debMetadata) (pkgdir string) {
	pkg := meta.Name() + "-" + meta.VersionString()
	pkgdir = filepath.Join(tmpdir, pkg)
	if err := os.Mkdir(pkgdir, 0755); err != nil {
		log.Fatal(err)
	}
	// Copy the source code.
	build.MustRunCommand("git", "checkout-index", "-a", "--prefix", pkgdir+string(filepath.Separator))

	// Put the debian build files in place.
	debian := filepath.Join(pkgdir, "debian")
	build.Render("utils/deb/deb.rules", filepath.Join(debian, "rules"), 0755, meta)
	build.Render("utils/deb/deb.changelog", filepath.Join(debian, "changelog"), 0644, meta)
	build.Render("utils/deb/deb.control", filepath.Join(debian, "control"), 0644, meta)
	build.Render("utils/deb/deb.copyright", filepath.Join(debian, "copyright"), 0644, meta)
	build.RenderString("8\n", filepath.Join(debian, "compat"), 0644, meta)
	build.RenderString("3.0 (native)\n", filepath.Join(debian, "source/format"), 0644, meta)
	for _, exe := range meta.Executables {
		install := filepath.Join(debian, meta.ExeName(exe)+".install")
		build.Render("utils/deb/deb.install", install, 0644, exe)

		docs := filepath.Join(debian, meta.ExeName(exe)+".docs")
		build.Render("utils/deb/deb.docs", docs, 0644, exe)

		if exe.PackageName == "godbledger-core" {
			preinst := filepath.Join(debian, meta.ExeName(exe)+".preinst")
			build.Render("utils/deb/godbledger-core.preinst", preinst, 0644, meta)

			postinst := filepath.Join(debian, meta.ExeName(exe)+".postinst")
			build.Render("utils/deb/godbledger-core.postinst", postinst, 0644, meta)

			prerm := filepath.Join(debian, meta.ExeName(exe)+".prerm")
			build.Render("utils/deb/godbledger-core.prerm", prerm, 0644, meta)

			postrm := filepath.Join(debian, meta.ExeName(exe)+".postrm")
			build.Render("utils/deb/godbledger-core.postrm", postrm, 0644, meta)

			servicefile := filepath.Join(debian, meta.ExeName(exe)+".service")
			build.Render("utils/deb/godbledger-core.service", servicefile, 0644, meta)
		}
	}
	return pkgdir
}

//// Cross compilation

func doXgo(cmdline []string) {
	var (
		xtarget = flag.String("target", "", "cross-compile target")
	)
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	if *xtarget == "" || strings.Contains(*xtarget, "*") {
		// TODO: not sure about this, limiting xgo to a single target, but it lets us manage the output to a target-based folder
		log.Println("must supply a single xgo build target for cross-compliation")
		os.Exit(1)
	}

	targetSuffix := strings.ReplaceAll(*xtarget, "/", "-")
	outDir := filepath.Join(distPath(), targetSuffix)
	os.MkdirAll(outDir, os.ModePerm)

	log.Printf("xgo target [%s] --> %s\n", *xtarget, outDir)

	// Make sure xgo is available for cross compilation
	gogetxgo := goTool("get", "src.techknowlogick.com/xgo")
	build.MustRun(gogetxgo)

	for _, cmd := range packagesToBuild {
		xgoArgs := append(buildFlags(env), flag.Args()...)
		xgoArgs = append(xgoArgs, []string{"--targets", *xtarget}...)
		xgoArgs = append(xgoArgs, []string{"--dest", outDir}...)
		xgoArgs = append(xgoArgs, "-v")
		xgoArgs = append(xgoArgs, "./"+cmd) // relative package name (assumes we are inside GOPATH)
		xgo := xgoTool(xgoArgs)
		build.MustRun(xgo)

		// strip the suffix out of the binary name
		// TODO: add this ability into xgo
		filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil // skip
			}

			suffix := filepath.Base(filepath.Dir(path))
			if strings.HasPrefix(info.Name(), cmd) && strings.Contains(info.Name(), suffix) {
				newName := strings.Replace(info.Name(), "-"+suffix, "", 1)
				newPath := filepath.Join(filepath.Dir(path), newName)
				log.Println("renaming:", path)
				log.Println("      to:", newPath)
				os.Rename(path, newPath)
			}
			return nil
		})
	}
}

func xgoTool(args []string) *exec.Cmd {
	cmd := exec.Command(filepath.Join(GOBIN, "xgo"), args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, []string{
		"GOBIN=" + GOBIN,
	}...)
	if os.Getenv("GOPATH") == "" {
		// xgo requires that $GOPATH be set
		homeRel := os.Getenv("HOME")
		if runtime.GOOS == "windows" {
			homeRel = os.Getenv("USERPROFILE")
		}
		if homeRel == "" {
			log.Println("GOPATH undefined and cannot determine homedir")
			os.Exit(1)
		}
		homeAbs, _ := filepath.Abs(homeRel)
		goPath := filepath.Join(homeAbs, "go")
		log.Printf("GOPATH undefined but required by xgo; injecting %s\n", goPath)
		cmd.Env = append(cmd.Env, []string{
			"GOPATH=" + goPath,
		}...)
	}
	return cmd
}
