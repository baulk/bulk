package netutils

import (
	"crypto/tls"
	"encoding/hex"
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
	"golang.org/x/net/http2"
)

// error
var (
	ErrProxyNotConfigured = errors.New("Proxy is not configured correctly")
)

// ProxySettings todo
type ProxySettings struct {
	ProxyServer   string
	ProxyOverride string // aka no proxy
	sep           string
}

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}

// Executor download executor
type Executor struct {
	client      *http.Client
	Destination string
	IsDebugMode bool
}

func isTrue(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "on" || s == "yes" || s == "1"
}

// NewExecutor new executor
func NewExecutor(destdir string) *Executor {
	ps, err := ResolveProxy()
	InsecureSkipVerify := isTrue(os.Getenv("BULK_INSECURE_TLS"))
	var transport http.RoundTripper
	if err == nil {
		h1transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: InsecureSkipVerify,
			},
		}
		proxyurl := ps.ProxyServer
		if !strings.Contains(proxyurl, "://") {
			proxyurl = "http://" + proxyurl // avoid proxy url parse failed
		}
		if u, err := url.Parse(proxyurl); err == nil {
			h1transport.Proxy = http.ProxyURL(u)
		}
		transport = h1transport
	} else {
		transport = &http2.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: InsecureSkipVerify,
			},
		}
	}
	_ = os.MkdirAll(destdir, 0755)
	return &Executor{
		Destination: destdir,
		client: &http.Client{
			Transport: transport,
		},
	}
}

func (e *Executor) resolveFileName(resp *http.Response, rawurl string) string {
	if disp := resp.Header.Get("Content-Disposition"); disp != "" {
		if _, params, err := mime.ParseMediaType(disp); err == nil {
			if filename := params["filename"]; len(filename) > 0 {
				if !base.PathIsSlipVulnerability(filename) {
					e.DbgPrint("Resolve Content-Disposition to '%s'", filename)
					return path.Base(filename)
				}
			}
		}
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return "index.html"
	}
	if filename := path.Base(u.Path); filename != "" && filename != "." {
		return filename
	}
	return "index.html"
}

// FileHashEqual file exists
func FileHashEqual(fullname, checksum string) bool {
	if checksum == "" {
		return false
	}
	if _, err := os.Stat(fullname); err != nil {
		return false
	}
	hc := NewHashComparator(checksum)
	if hc == nil {
		return false
	}
	fd, err := os.Open(fullname)
	if err != nil {
		return false
	}
	if _, err := io.Copy(hc.H, fd); err != nil && err != io.EOF {
		return false
	}
	if hsx2 := hex.EncodeToString(hc.H.Sum(nil)); hsx2 != hc.S {
		return false
	}
	return true
}

// DbgPrint todo
func (e *Executor) DbgPrint(format string, a ...interface{}) {
	if e.IsDebugMode {
		ss := fmt.Sprintf(format, a...)
		_, _ = os.Stderr.WriteString(base.StrCat("\x1b[33m* ", ss, "\x1b[0m\n"))
	}
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

// WebGet get file from network
func (e *Executor) WebGet(rawurl, checksum string) (string, error) {
	req, err := http.NewRequest("GET", rawurl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "bulk/1.0")
	if e.IsDebugMode {
		req = e.traceRequest(req)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		if e.IsDebugMode {
			_, _ = io.Copy(os.Stderr, resp.Body)
			_, _ = os.Stderr.Write([]byte("\n"))
		}
		return "", base.ErrorCat("response ", resp.Status)
	}
	e.DbgPrint("%s %s", resp.Proto, resp.Status)
	if e.IsDebugMode {
		for k, v := range resp.Header {
			fmt.Fprintf(os.Stderr, "\x1b[33m* %s: %s\x1b[0m\n", k, strings.Join(v, "; "))
		}
	}
	filename := e.resolveFileName(resp, rawurl)
	fullname := filepath.Join(e.Destination, filename)
	if FileHashEqual(fullname, checksum) {
		e.DbgPrint("Found '%s' hash equal '%s'", filename, checksum)
		return fullname, nil
	}
	part := filename + ".part"
	hc := NewHashComparator(checksum)
	fd, err := os.Create(part)
	if err != nil {
		return "", err
	}
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		filename,
	)
	var w io.Writer
	if hc != nil {
		w = io.MultiWriter(fd, bar, hc.H)
	} else {
		w = io.MultiWriter(fd, bar)
	}
	if _, err := io.Copy(w, resp.Body); err != nil && err != io.EOF {
		fd.Close()
		return "", err
	}
	if hc != nil {
		if err := hc.IsMatch(); err != nil {
			fd.Close()
			return "", err
		}
	}
	fd.Close()
	if err := MoveFile(part, fullname); err != nil {
		return "", base.ErrorCat("unable move ", part, "to ", fullname, " error: ", err.Error())
	}
	return fullname, nil
}
