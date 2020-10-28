package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baulk/bulk/base"
	"github.com/baulk/bulk/netutils"
)

// rind golang base download utils

type rindOptions struct {
	opt netutils.Options
	eu  netutils.DownloadEntity
}

// version info
var (
	VERSION     = "1.0"
	BUILDTIME   string
	BUILDCOMMIT string
	BUILDBRANCH string
	GOVERSION   string
)

func version() {
	fmt.Fprintf(os.Stdout, `rind - Interesting command line download tool
version:       %s
build branch:  %s
build commit:  %s
build time:    %s
go version:    %s
`, VERSION, BUILDBRANCH, BUILDCOMMIT, BUILDTIME, GOVERSION)

}

func usage() {
	fmt.Fprintf(os.Stdout, `rind - Interesting command line download tool
usage: %s <option> url
  -h|--help        Show usage text and quit
  -v|--version     Show version number and quit
  -V|--verbose     Make the operation more talkative
  -F|--force       Turn on force mode. such as overwrite exists file
  -k|--insecure    Allow insecure server connections when using SSL
  -C|--checksum    Specify the checksum of the downloaded file
  -W|--workdir     Specify download file save directory. (default $PWD)
  -O|--out         Write file to specified path
  -S|--show        Show the checksum of the downloaded file (-S, -Sblake3 -Ssha256 default: sha256)
  -A|--user-agent  Send User-Agent <name> to server
  --https-proxy    Use this proxy. Equivalent to setting the environment variable 'HTTPS_PROXY'
  --direct         Download files without automatic proxy        

`, os.Args[0])
}

//-A|--user-agent  Send User-Agent <name> to server

func (ro *rindOptions) Invoke(val int, oa, raw string) error {
	switch val {
	case 'h':
		usage()
		os.Exit(0)
	case 'v':
		version()
		os.Exit(0)
	case 'F':
		ro.eu.OverwriteExisting = true
	case 'k':
		ro.opt.InsecureSkipVerify = true
	case 'V':
		base.IsDebugMode = true
	case 'A': // --user-agent
		ro.opt.UserAgent = oa
	case 'C': //--checksum
		ro.eu.HashValue = oa
	case 'W':
		outdir, err := filepath.Abs(oa)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[31munable resolve outdir absPath %s\x1b[0m\n", oa)
			os.Exit(1)
		}
		ro.opt.DestinationPath = outdir
	case 'O':
		absPath, err := filepath.Abs(oa)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[31munable resolve output absPath %s\x1b[0m\n", oa)
			os.Exit(1)
		}
		ro.eu.Destination = absPath
	case 'S': // --show checksum
		if len(oa) == 0 {
			ro.eu.Algorithm = "SHA256"
		} else {
			ro.eu.Algorithm = oa
		}
	case 1001: // --https-proxy
		ro.opt.ProxyURL = oa
	case 1002: // --noproxy
		ro.opt.DisableAutoProxy = true
	}
	return nil
}

func (ro *rindOptions) ParseArgv() error {
	var pa base.ParseArgs
	pa.Add("help", base.NOARG, 'h')
	pa.Add("version", base.NOARG, 'v')
	pa.Add("verbose", base.NOARG, 'V')
	pa.Add("out", base.REQUIRED, 'O')
	pa.Add("user-agent", base.REQUIRED, 'A')
	pa.Add("checksum", base.REQUIRED, 'C')
	pa.Add("force", base.NOARG, 'F')
	pa.Add("insecure", base.NOARG, 'k')
	pa.Add("show", base.OPTIONAL, 'S')
	pa.Add("workdir", base.REQUIRED, 'W')
	pa.Add("https-proxy", base.NOARG, 1001)
	pa.Add("direct", base.NOARG, 1002)
	if err := pa.Execute(os.Args, ro); err != nil {
		return err
	}
	ua := pa.Unresolved()
	if len(ua) == 0 {
		return errors.New("missing URL")
	}
	ro.eu.URL = ua[0]
	return nil
}

func main() {
	var ro rindOptions
	if err := ro.ParseArgv(); err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[31mrind: %v\x1b[0m\n", err)
		os.Exit(1)
	}
	e := netutils.NewExecutor(&ro.opt)
	filename, err := e.WebGet(&ro.eu)
	if err != nil {
		fmt.Fprintf(os.Stderr, "download \x1b[31m%v\x1b[0m\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "download %s success\n", filename)
}
