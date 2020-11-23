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

Usage: go run build/ci.go <command> <command flags/arguments>

Available commands are:

   install    [ -arch architecture ] [ -cc compiler ] [ packages... ]                          -- builds packages and executables
   test       [ -coverage ] [ packages... ]                                                    -- runs the tests
   lint                                                                                        -- runs certain pre-selected linters
   archive    [ -arch architecture ] [ -type zip|tar ] [ -signer key-envvar ] [ -upload dest ] -- archives build artifacts
   importkeys                                                                                  -- imports signing keys from env
   debsrc     [ -signer key-id ] [ -upload dest ]                                              -- creates a debian source package
   nsis                                                                                        -- creates a Windows NSIS installer
   aar        [ -local ] [ -sign key-id ] [-deploy repo] [ -upload dest ]                      -- creates an Android archive
   xcode      [ -local ] [ -sign key-id ] [-deploy repo] [ -upload dest ]                      -- creates an iOS XCode framework
   xgo        [ -alltools ] [ options ]                                                        -- cross builds according to options
   purge      [ -store blobstore ] [ -days threshold ]                                         -- purges old archives from the blobstore

For all commands, -n prevents execution of external programs (dry run mode).

*/
package main

import (
	//"bufio"
	//"bytes"
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

	//"time"
	//"github.com/cespare/cp"
	"github.com/darcys22/godbledger/godbledger/version"
	"github.com/darcys22/godbledger/internal/build"
)

var (
	// NOTE: this could be inferred if these 'main' packages were moved
	//       into a ./cmd folder (see: https://github.com/golang-standards/project-layout#cmd)
	packagesToBuild = []string{
		"godbledger",
		"ledger_cli",
		"reporter",
	}

	// Files that end up in the godbledger*.zip archive.
	godbledgerArchiveFiles = []string{
		executablePath("godbledger"),
		executablePath("ledger_cli"),
		executablePath("reporter"),
	}

	allToolsArchiveFiles = godbledgerArchiveFiles

	// A debian package is created for all executables listed here.
	debExecutables = []debExecutable{
		{
			BinaryName:  "godbledger",
			Description: "Accounting server to manage financial databases and record double entry bookkeeping transactions",
		},
		{
			BinaryName:  "ledger_cli",
			Description: "Godbledger grpc client",
		},
		{
			BinaryName:  "reporter",
			Description: "basic reporting queries on SQL databases associated with godbledger servers",
		},
	}

	// A debian package is created for all executables listed here.

	debGoDBLedger = debPackage{
		Name:        "GoDBLedger",
		Version:     version.Version,
		Executables: debExecutables,
	}

	// Debian meta packages to build and push to Ubuntu PPA
	debPackages = []debPackage{
		debGoDBLedger,
	}

	// Distros for which packages are created.
	debDistroGoBoots = map[string]string{
		"xenial": "golang-go",
		"bionic": "golang-go",
		"disco":  "golang-go",
		"eoan":   "golang-go",
		"focal":  "golang-go",
	}

	debGoBootPaths = map[string]string{
		"golang-go": "/usr/lib/go",
	}
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
		//doArchive(os.Args[2:])
	case "debsrc":
		//doDebianSource(os.Args[2:])
	case "nsis":
		//doWindowsInstaller(os.Args[2:])
	case "xgo":
		doXgo(os.Args[2:])
	case "purge":
		//doPurge(os.Args[2:])
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
		ld = append(ld, "-X", "main.gitCommit="+env.Commit)
		ld = append(ld, "-X", "main.gitDate="+env.Date)
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
	if err := build.ExtractTarballArchive(archivePath, cachedir); err != nil {
		log.Fatal(err)
	}
	return filepath.Join(cachedir, base, "golangci-lint")
}

// Release Packaging
//func doArchive(cmdline []string) {
//var (
//arch   = flag.String("arch", runtime.GOARCH, "Architecture cross packaging")
//atype  = flag.String("type", "zip", "Type of archive to write (zip|tar)")
//signer = flag.String("signer", "", `Environment variable holding the signing key (e.g. LINUX_SIGNING_KEY)`)
//upload = flag.String("upload", "", `Destination to upload the archives (usually "gethstore/builds")`)
//ext    string
//)
//flag.CommandLine.Parse(cmdline)
//switch *atype {
//case "zip":
//ext = ".zip"
//case "tar":
//ext = ".tar.gz"
//default:
//log.Fatal("unknown archive type: ", atype)
//}

//var (
////env = build.Env()

////basegeth = archiveBasename(*arch, params.ArchiveVersion(env.Commit))
////geth     = "geth-" + basegeth + ext
////alltools = "geth-alltools-" + basegeth + ext
//)
////maybeSkipArchive(env)
//if err := build.WriteArchive(geth, gethArchiveFiles); err != nil {
//log.Fatal(err)
//}
//if err := build.WriteArchive(alltools, allToolsArchiveFiles); err != nil {
//log.Fatal(err)
//}
//for _, archive := range []string{geth, alltools} {
//if err := archiveUpload(archive, *upload, *signer); err != nil {
//log.Fatal(err)
//}
//}
//}

//func archiveBasename(arch string, archiveVersion string) string {
//platform := runtime.GOOS + "-" + arch
//if arch == "arm" {
//platform += os.Getenv("GOARM")
//}
//if arch == "android" {
//platform = "android-all"
//}
//if arch == "ios" {
//platform = "ios-all"
//}
//return platform + "-" + archiveVersion
//}

//func archiveUpload(archive string, blobstore string, signer string) error {
//If signing was requested, generate the signature files
//if signer != "" {
//key := getenvBase64(signer)
//if err := build.PGPSignFile(archive, archive+".asc", string(key)); err != nil {
//return err
//}
//}
//If uploading to Azure was requested, push the archive possibly with its signature
//if blobstore != "" {
//auth := build.AzureBlobstoreConfig{
//Account:   strings.Split(blobstore, "/")[0],
//Token:     os.Getenv("AZURE_BLOBSTORE_TOKEN"),
//Container: strings.SplitN(blobstore, "/", 2)[1],
//}
//if err := build.AzureBlobstoreUpload(archive, filepath.Base(archive), auth); err != nil {
//return err
//}
//if signer != "" {
//if err := build.AzureBlobstoreUpload(archive+".asc", filepath.Base(archive+".asc"), auth); err != nil {
//return err
//}
//}
//}
//return nil
//}

// skips archiving for some build configurations.
//func maybeSkipArchive(env build.Environment) {
//if env.IsPullRequest {
//log.Printf("skipping because this is a PR build")
//os.Exit(0)
//}
//if env.IsCronJob {
//log.Printf("skipping because this is a cron job")
//os.Exit(0)
//}
//if env.Branch != "master" && !strings.HasPrefix(env.Tag, "v1.") {
//log.Printf("skipping because branch %q, tag %q is not on the whitelist", env.Branch, env.Tag)
//os.Exit(0)
//}
//}

// Debian Packaging
func doDebianSource(cmdline []string) {
	//var (
	//goversion = flag.String("goversion", "", `Go version to build with (will be included in the source package)`)
	//cachedir  = flag.String("cachedir", "./build/cache", `Filesystem path to cache the downloaded Go bundles at`)
	//signer    = flag.String("signer", "", `Signing key name, also used as package author`)
	//upload    = flag.String("upload", "", `Where to upload the source package (usually "ethereum/ethereum")`)
	//sshUser   = flag.String("sftp-user", "", `Username for SFTP upload (usually "geth-ci")`)
	//workdir   = flag.String("workdir", "", `Output directory for packages (uses temp dir if unset)`)
	//now       = time.Now()
	//)
	//flag.CommandLine.Parse(cmdline)
	//*workdir = makeWorkdir(*workdir)
	//env := build.Env()
	//maybeSkipArchive(env)

	//// Import the signing key.
	//if key := getenvBase64("PPA_SIGNING_KEY"); len(key) > 0 {
	//gpg := exec.Command("gpg", "--import")
	//gpg.Stdin = bytes.NewReader(key)
	//build.MustRun(gpg)
	//}

	//// Download and verify the Go source package.
	//gobundle := downloadGoSources(*goversion, *cachedir)

	//// Download all the dependencies needed to build the sources and run the ci script
	//srcdepfetch := goTool("install", "-n", "./...")
	//srcdepfetch.Env = append(os.Environ(), "GOPATH="+filepath.Join(*workdir, "modgopath"))
	//build.MustRun(srcdepfetch)

	//cidepfetch := goTool("run", "./build/ci.go")
	//cidepfetch.Env = append(os.Environ(), "GOPATH="+filepath.Join(*workdir, "modgopath"))
	//cidepfetch.Run() // Command fails, don't care, we only need the deps to start it

	//// Create Debian packages and upload them.
	//for _, pkg := range debPackages {
	//for distro, goboot := range debDistroGoBoots {
	//// Prepare the debian package with the go-ethereum sources.
	//meta := newDebMetadata(distro, goboot, *signer, env, now, pkg.Name, pkg.Version, pkg.Executables)
	//pkgdir := stageDebianSource(*workdir, meta)

	//// Add Go source code
	//if err := build.ExtractTarballArchive(gobundle, pkgdir); err != nil {
	//log.Fatalf("Failed to extract Go sources: %v", err)
	//}
	//if err := os.Rename(filepath.Join(pkgdir, "go"), filepath.Join(pkgdir, ".go")); err != nil {
	//log.Fatalf("Failed to rename Go source folder: %v", err)
	//}
	//// Add all dependency modules in compressed form
	//os.MkdirAll(filepath.Join(pkgdir, ".mod", "cache"), 0755)
	//if err := cp.CopyAll(filepath.Join(pkgdir, ".mod", "cache", "download"), filepath.Join(*workdir, "modgopath", "pkg", "mod", "cache", "download")); err != nil {
	//log.Fatalf("Failed to copy Go module dependencies: %v", err)
	//}
	//// Run the packaging and upload to the PPA
	//debuild := exec.Command("debuild", "-S", "-sa", "-us", "-uc", "-d", "-Zxz", "-nc")
	//debuild.Dir = pkgdir
	//build.MustRun(debuild)

	//var (
	//basename = fmt.Sprintf("%s_%s", meta.Name(), meta.VersionString())
	//source   = filepath.Join(*workdir, basename+".tar.xz")
	//dsc      = filepath.Join(*workdir, basename+".dsc")
	//changes  = filepath.Join(*workdir, basename+"_source.changes")
	//)
	//if *signer != "" {
	//build.MustRunCommand("debsign", changes)
	//}
	//if *upload != "" {
	//ppaUpload(*workdir, *upload, *sshUser, []string{source, dsc, changes})
	//}
	//}
	//}
}

//func downloadGoSources(version string, cachedir string) string {
//csdb := build.MustLoadChecksums("build/checksums.txt")
//file := fmt.Sprintf("go%s.src.tar.gz", version)
//url := "https://dl.google.com/go/" + file
//dst := filepath.Join(cachedir, file)
//if err := csdb.DownloadFile(url, dst); err != nil {
//log.Fatal(err)
//}
//return dst
//}

func ppaUpload(workdir, ppa, sshUser string, files []string) {
	//p := strings.Split(ppa, "/")
	//if len(p) != 2 {
	//log.Fatal("-upload PPA name must contain single /")
	//}
	//if sshUser == "" {
	//sshUser = p[0]
	//}
	//incomingDir := fmt.Sprintf("~%s/ubuntu/%s", p[0], p[1])
	//// Create the SSH identity file if it doesn't exist.
	//var idfile string
	//if sshkey := getenvBase64("PPA_SSH_KEY"); len(sshkey) > 0 {
	//idfile = filepath.Join(workdir, "sshkey")
	//if _, err := os.Stat(idfile); os.IsNotExist(err) {
	//ioutil.WriteFile(idfile, sshkey, 0600)
	//}
	//}
	//// Upload
	//dest := sshUser + "@ppa.launchpad.net"
	//if err := build.UploadSFTP(idfile, dest, incomingDir, files); err != nil {
	//log.Fatal(err)
	//}
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
		wdflag, err = ioutil.TempDir("", "geth-build-")
	}
	if err != nil {
		log.Fatal(err)
	}
	return wdflag
}

//func isUnstableBuild(env build.Environment) bool {
//if env.Tag != "" {
//return false
//}
//return true
//}

type debPackage struct {
	Name        string          // the name of the Debian package to produce, e.g. "ethereum"
	Version     string          // the clean version of the debPackage, e.g. 1.8.12, without any metadata
	Executables []debExecutable // executables to be included in the package
}

type debMetadata struct {
	//Env           build.Environment
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

//func newDebMetadata(distro, goboot, author string, env build.Environment, t time.Time, name string, version string, exes []debExecutable) debMetadata {
//if author == "" {
//// No signing key, use default author.
//author = "Ethereum Builds <fjl@ethereum.org>"
//}
//return debMetadata{
//GoBootPackage: goboot,
//GoBootPath:    debGoBootPaths[goboot],
//PackageName:   name,
//Env:           env,
//Author:        author,
//Distro:        distro,
//Version:       version,
//Time:          t.Format(time.RFC1123Z),
//Executables:   exes,
//}
//}

// Name returns the name of the metapackage that depends
// on all executable packages.
//func (meta debMetadata) Name() string {
//if isUnstableBuild(meta.Env) {
//return meta.PackageName + "-unstable"
//}
//return meta.PackageName
//}

// VersionString returns the debian version of the packages.
//func (meta debMetadata) VersionString() string {
//vsn := meta.Version
//if meta.Env.Buildnum != "" {
//vsn += "+build" + meta.Env.Buildnum
//}
//if meta.Distro != "" {
//vsn += "+" + meta.Distro
//}
//return vsn
//}

// ExeList returns the list of all executable packages.
//func (meta debMetadata) ExeList() string {
//names := make([]string, len(meta.Executables))
//for i, e := range meta.Executables {
//names[i] = meta.ExeName(e)
//}
//return strings.Join(names, ", ")
//}

// ExeName returns the package name of an executable package.
//func (meta debMetadata) ExeName(exe debExecutable) string {
//if isUnstableBuild(meta.Env) {
//return exe.Package() + "-unstable"
//}
//return exe.Package()
//}

// ExeConflicts returns the content of the Conflicts field
// for executable packages.
//func (meta debMetadata) ExeConflicts(exe debExecutable) string {
//if isUnstableBuild(meta.Env) {
//// Set up the conflicts list so that the *-unstable packages
//// cannot be installed alongside the regular version.
////
//// https://www.debian.org/doc/debian-policy/ch-relationships.html
//// is very explicit about Conflicts: and says that Breaks: should
//// be preferred and the conflicting files should be handled via
//// alternates. We might do this eventually but using a conflict is
//// easier now.
//return "ethereum, " + exe.Package()
//}
//return ""
//}

//func stageDebianSource(tmpdir string, meta debMetadata) (pkgdir string) {
//pkg := meta.Name() + "-" + meta.VersionString()
//pkgdir = filepath.Join(tmpdir, pkg)
//if err := os.Mkdir(pkgdir, 0755); err != nil {
//log.Fatal(err)
//}
//// Copy the source code.
//build.MustRunCommand("git", "checkout-index", "-a", "--prefix", pkgdir+string(filepath.Separator))

//// Put the debian build files in place.
//debian := filepath.Join(pkgdir, "debian")
//build.Render("build/deb/"+meta.PackageName+"/deb.rules", filepath.Join(debian, "rules"), 0755, meta)
//build.Render("build/deb/"+meta.PackageName+"/deb.changelog", filepath.Join(debian, "changelog"), 0644, meta)
//build.Render("build/deb/"+meta.PackageName+"/deb.control", filepath.Join(debian, "control"), 0644, meta)
//build.Render("build/deb/"+meta.PackageName+"/deb.copyright", filepath.Join(debian, "copyright"), 0644, meta)
//build.RenderString("8\n", filepath.Join(debian, "compat"), 0644, meta)
//build.RenderString("3.0 (native)\n", filepath.Join(debian, "source/format"), 0644, meta)
//for _, exe := range meta.Executables {
//install := filepath.Join(debian, meta.ExeName(exe)+".install")
//docs := filepath.Join(debian, meta.ExeName(exe)+".docs")
//build.Render("build/deb/"+meta.PackageName+"/deb.install", install, 0644, exe)
//build.Render("build/deb/"+meta.PackageName+"/deb.docs", docs, 0644, exe)
//}
//return pkgdir
//}

//// Windows installer
//func doWindowsInstaller(cmdline []string) {
//// Parse the flags and make skip installer generation on PRs
//var (
//arch    = flag.String("arch", runtime.GOARCH, "Architecture for cross build packaging")
//signer  = flag.String("signer", "", `Environment variable holding the signing key (e.g. WINDOWS_SIGNING_KEY)`)
//upload  = flag.String("upload", "", `Destination to upload the archives (usually "gethstore/builds")`)
//workdir = flag.String("workdir", "", `Output directory for packages (uses temp dir if unset)`)
//)
//flag.CommandLine.Parse(cmdline)
//*workdir = makeWorkdir(*workdir)
//env := build.Env()
//maybeSkipArchive(env)

//// Aggregate binaries that are included in the installer
//var (
//devTools []string
//allTools []string
//gethTool string
//)
//for _, file := range allToolsArchiveFiles {
//if file == "COPYING" { // license, copied later
//continue
//}
//allTools = append(allTools, filepath.Base(file))
//if filepath.Base(file) == "geth.exe" {
//gethTool = file
//} else {
//devTools = append(devTools, file)
//}
//}

//// Render NSIS scripts: Installer NSIS contains two installer sections,
//// first section contains the geth binary, second section holds the dev tools.
//templateData := map[string]interface{}{
//"License":  "COPYING",
//"Geth":     gethTool,
//"DevTools": devTools,
//}
//build.Render("build/nsis.geth.nsi", filepath.Join(*workdir, "geth.nsi"), 0644, nil)
//build.Render("build/nsis.install.nsh", filepath.Join(*workdir, "install.nsh"), 0644, templateData)
//build.Render("build/nsis.uninstall.nsh", filepath.Join(*workdir, "uninstall.nsh"), 0644, allTools)
//build.Render("build/nsis.pathupdate.nsh", filepath.Join(*workdir, "PathUpdate.nsh"), 0644, nil)
//build.Render("build/nsis.envvarupdate.nsh", filepath.Join(*workdir, "EnvVarUpdate.nsh"), 0644, nil)
//if err := cp.CopyFile(filepath.Join(*workdir, "SimpleFC.dll"), "build/nsis.simplefc.dll"); err != nil {
//log.Fatal("Failed to copy SimpleFC.dll: %v", err)
//}
//if err := cp.CopyFile(filepath.Join(*workdir, "COPYING"), "COPYING"); err != nil {
//log.Fatal("Failed to copy copyright note: %v", err)
//}
// Build the installer. This assumes that all the needed files have been previously
// built (don't mix building and packaging to keep cross compilation complexity to a
// minimum).
//version := strings.Split(params.Version, ".")
//if env.Commit != "" {
//version[2] += "-" + env.Commit[:8]
//}
//installer, _ := filepath.Abs("geth-" + archiveBasename(*arch, params.ArchiveVersion(env.Commit)) + ".exe")
//build.MustRunCommand("makensis.exe",
//"/DOUTPUTFILE="+installer,
//"/DMAJORVERSION="+version[0],
//"/DMINORVERSION="+version[1],
//"/DBUILDVERSION="+version[2],
//"/DARCH="+*arch,
//filepath.Join(*workdir, "geth.nsi"),
//)
// Sign and publish installer.
//if err := archiveUpload(installer, *upload, *signer); err != nil {
//log.Fatal(err)
//}
//}

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
			"GOPATH="+goPath,
		}...)
	}
	return cmd
}

// Binary distribution cleanups

//func doPurge(cmdline []string) {
//var (
//store = flag.String("store", "", `Destination from where to purge archives (usually "gethstore/builds")`)
//limit = flag.Int("days", 30, `Age threshold above which to delete unstable archives`)
//)
//flag.CommandLine.Parse(cmdline)

//if env := build.Env(); !env.IsCronJob {
//log.Printf("skipping because not a cron job")
//os.Exit(0)
//}
//// Create the azure authentication and list the current archives
//auth := build.AzureBlobstoreConfig{
//Account:   strings.Split(*store, "/")[0],
//Token:     os.Getenv("AZURE_BLOBSTORE_TOKEN"),
//Container: strings.SplitN(*store, "/", 2)[1],
//}
//blobs, err := build.AzureBlobstoreList(auth)
//if err != nil {
//log.Fatal(err)
//}
//fmt.Printf("Found %d blobs\n", len(blobs))

// Iterate over the blobs, collect and sort all unstable builds
//for i := 0; i < len(blobs); i++ {
//if !strings.Contains(blobs[i].Name, "unstable") {
//blobs = append(blobs[:i], blobs[i+1:]...)
//i--
//}
//}
//for i := 0; i < len(blobs); i++ {
//for j := i + 1; j < len(blobs); j++ {
//if blobs[i].Properties.LastModified.After(blobs[j].Properties.LastModified) {
//blobs[i], blobs[j] = blobs[j], blobs[i]
//}
//}
//}
// Filter out all archives more recent that the given threshold
//for i, blob := range blobs {
//if time.Since(blob.Properties.LastModified) < time.Duration(*limit)*24*time.Hour {
//blobs = blobs[:i]
//break
//}
//}
//fmt.Printf("Deleting %d blobs\n", len(blobs))
// Delete all marked as such and return
//if err := build.AzureBlobstoreDelete(auth, blobs); err != nil {
//log.Fatal(err)
//}
//}
