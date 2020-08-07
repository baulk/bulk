package netutils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/baulk/bulk/base"
	"github.com/baulk/bulk/progressbar"
	"github.com/mattn/go-runewidth"
	"golang.org/x/net/http2"
)

// Options todo
type Options struct {
	InsecureSkipVerify bool
	IsDebugMode        bool
	DisableAutoProxy   bool
	DestinationPath    string
	ProxyURL           string
}

// DefaultOptions default
var DefaultOptions = &Options{}

// GetProxyURL get proxy url
func (opt *Options) GetProxyURL() string {
	if len(opt.ProxyURL) != 0 {
		return canonicalProxyURL(opt.ProxyURL)
	}
	if ps, err := ResolveProxy(); err == nil {
		return canonicalProxyURL(ps.ProxyServer)
	}
	return ""
}

// Executor download executor
type Executor struct {
	client          *http.Client
	DestinationPath string
	IsDebugMode     bool
}

// EnhanceURL metadata
type EnhanceURL struct {
	URL               string
	Destination       string
	HashValue         string
	Algorithm         string
	DisplayHash       bool // Calculate the hash checksum of the downloaded file
	OverwriteExisting bool
}

func isTrue(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "on" || s == "yes" || s == "1"
}

// NewExecutor new executor
func NewExecutor(opt *Options) *Executor {
	if opt == nil {
		opt = DefaultOptions
	}
	e := &Executor{DestinationPath: opt.DestinationPath, IsDebugMode: opt.IsDebugMode}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opt.InsecureSkipVerify,
		},
	}
	if !opt.DisableAutoProxy {
		if proxyURL := opt.GetProxyURL(); len(proxyURL) != 0 {
			if u, err := url.Parse(proxyURL); err == nil {
				transport.Proxy = http.ProxyURL(u)
				e.DbgPrint("Use ProxyURL %s", proxyURL)
			}
		}
	}
	http2.ConfigureTransport(transport)
	e.client = &http.Client{
		Transport: transport,
	}
	return e
}

// DbgPrint todo
func (e *Executor) DbgPrint(format string, a ...interface{}) {
	if e.IsDebugMode {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(base.StrCat("\x1b[33m* ", ss, "\x1b[0m\n"))
	}
}

func isCorrectName(name string) bool {
	return name != "" && name != "." && name != "/"
}

// ResolveFileName resolve filename from response and rawurl
func (e *Executor) ResolveFileName(resp *http.Response, rawurl string) string {
	if disp := resp.Header.Get("Content-Disposition"); disp != "" {
		if _, params, err := mime.ParseMediaType(disp); err == nil {
			if filename := params["filename"]; len(filename) > 0 {
				if !base.PathIsSlipVulnerability(filename) {
					e.DbgPrint("Resolve '%s' from Content-Disposition", filename)
					return path.Base(filename)
				}
			}
		}
	}
	if u, err := url.Parse(rawurl); err == nil {
		if filename := path.Base(u.Path); isCorrectName(filename) {
			return filename
		}
	}
	return "index.html"
}

func tlsVersionName(i uint16) string {
	switch int(i) {
	case tls.VersionTLS13:
		return "TLSv1.3"
	case tls.VersionTLS12:
		return "TLSv1.2"
	case tls.VersionTLS11:
		return "TLSv1.1"
	}
	return "unsupported version"
}

func flatAddress(addrs []net.IPAddr) string {
	if len(addrs) == 0 {
		return "<empty>"
	}
	ss := make([]string, 0, len(addrs))
	for _, s := range addrs {
		ss = append(ss, s.String())
	}
	return strings.Join(ss, "|")
}

func (e *Executor) traceRequest(req *http.Request) *http.Request {
	trace := &httptrace.ClientTrace{
		DNSStart: func(di httptrace.DNSStartInfo) {
			fmt.Fprintf(os.Stderr, "\x1b[33mResolve %s\x1b[0m", di.Host)
		},
		DNSDone: func(dnsinfo httptrace.DNSDoneInfo) {
			if dnsinfo.Err == nil {
				fmt.Fprintf(os.Stderr, "\x1b[33m to %s\x1b[0m\n", flatAddress(dnsinfo.Addrs))
			}
		},
		ConnectDone: func(network, addr string, err error) {
			if err == nil {
				fmt.Fprintf(os.Stderr, "\x1b[33mConnecting to %s connected\x1b[0m\n", addr)
			}
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err == nil {
				fmt.Fprintf(os.Stderr, "\x1b[33mSSL connection using %s/%s\x1b[0m\n", tlsVersionName(state.Version), tls.CipherSuiteName(state.CipherSuite))
			}
		},
		WroteHeaderField: func(key string, value []string) {
			fmt.Fprintf(os.Stderr, "\x1b[33m> %s: %s\x1b[0m\n", key, strings.Join(value, "; "))
		},
	}
	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
}

func (e *Executor) traceHeader(resp *http.Response) {
	if e.IsDebugMode {
		for k, v := range resp.Header {
			fmt.Fprintf(os.Stderr, "\x1b[33m* %s: %s\x1b[0m\n", k, strings.Join(v, "; "))
		}
	}
}

func (e *Executor) traceResponseError(resp *http.Response) {
	if e.IsDebugMode {
		_, _ = io.Copy(os.Stderr, resp.Body)
		_, _ = os.Stderr.Write([]byte("\n"))
	}
}

// ResolvePath todo
func (e *Executor) ResolvePath(resp *http.Response, eu *EnhanceURL) (string, error) {
	// When output name is set OverwriteExisting is default
	if eu.Destination != "" {
		destination, err := filepath.Abs(eu.Destination)
		if err != nil {
			return "", err
		}
		destdir := filepath.Dir(destination)
		if err := os.MkdirAll(destdir, 0755); err != nil {
			return "", err
		}
		return destination, nil
	}
	filename := e.ResolveFileName(resp, eu.URL)
	destinationPath, err := filepath.Abs(e.DestinationPath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(destinationPath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(destinationPath, 0755); err != nil {
			return "", err
		}
	}
	dest := filepath.Join(destinationPath, filename)
	if eu.OverwriteExisting {
		return dest, nil
	}
	if _, err := os.Stat(dest); err != nil && os.IsNotExist(err) {
		return dest, nil
	}
	name, ext := stripExtension(filename)
	for i := 1; i < 1001; i++ {
		dest := filepath.Join(destinationPath, fmt.Sprintf("%s-(%d)%s", name, i, ext))
		if _, err := os.Stat(dest); err != nil && os.IsNotExist(err) {
			return dest, nil
		}
	}
	return "", base.ErrorCat("'", filename, "' already exists")
}

func fileNameOrDescription(filename string) string {
	if runewidth.StringWidth(filename) > 20 {
		return "downloading"
	}
	return filename
}

// WebGet get file from network
func (e *Executor) WebGet(eu *EnhanceURL) (string, error) {
	if eu == nil {
		return "", errors.New("incorrect WebGet param")
	}
	req, err := http.NewRequest("GET", eu.URL, nil)
	if err != nil {
		return "", err
	}
	if e.IsDebugMode {
		req = e.traceRequest(req) // register trace info
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		e.traceResponseError(resp)
		return "", base.ErrorCat("response ", resp.Status)
	}
	e.DbgPrint("%s %s", resp.Proto, resp.Status)
	e.traceHeader(resp)
	dest, err := e.ResolvePath(resp, eu)
	if err != nil {
		return "", err
	}
	e.DbgPrint("Resolve save path %s", dest)
	if FileHashEqual(dest, eu) {
		e.DbgPrint("Found '%s' hash equal '%s'", dest, eu.HashValue)
		return dest, nil
	}
	filename := filepath.Base(dest)
	part := dest + ".part"
	verifier := NewVerifier(eu)
	fd, err := os.Create(part)
	if err != nil {
		return "", err
	}
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		fileNameOrDescription(filename),
	)
	var w io.Writer
	if verifier != nil {
		w = io.MultiWriter(fd, bar, verifier.H)
	} else {
		w = io.MultiWriter(fd, bar)
	}
	if _, err := io.Copy(w, resp.Body); err != nil && err != io.EOF {
		fd.Close()
		return "", err
	}
	if verifier != nil {
		if err := verifier.IsMatch(filename); err != nil {
			fd.Close()
			return "", err
		}
	}
	fd.Close()
	if err := MoveFile(part, dest); err != nil {
		return "", base.ErrorCat("unable move ", part, "to ", dest, " error: ", err.Error())
	}
	return dest, nil
}
