package netutils

import (
	"crypto/tls"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/baulk/bulk/base"
	"github.com/baulk/bulk/progressbar"
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
	client *http.Client
	OutDir string
}

func isTrue(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "on" || s == "yes" || s == "1"
}

// NewExecutor new executor
func NewExecutor(outdir string) *Executor {
	ps, err := ResolveProxy()
	transport := &http.Transport{}
	if err == nil {
		proxyurl := ps.ProxyServer
		if !strings.Contains(proxyurl, "://") {
			proxyurl = "http://" + proxyurl // avoid proxy url parse failed
		}
		if u, err := url.Parse(proxyurl); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	_ = os.MkdirAll(outdir, 0755)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: isTrue(os.Getenv("BULK_INSECURE_TLS"))} //set ssl
	return &Executor{
		OutDir: outdir,
		client: &http.Client{
			Transport: transport,
		},
	}
}

func resolveFileName(resp *http.Response, rawurl string) string {
	if disp := resp.Header.Get("Content-Disposition"); disp != "" {
		if _, params, err := mime.ParseMediaType(disp); err == nil {
			if filename := params["filename"]; len(filename) > 0 {
				if !base.PathIsSlipVulnerability(filename) {
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

// IsAlreadyExistsFile file exists
func IsAlreadyExistsFile(fullname, checksum string) bool {
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

// Get get file from network
func (e *Executor) Get(rawurl, checksum string) (string, error) {
	req, err := http.NewRequest("GET", rawurl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "bulk/1.0")
	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	filename := resolveFileName(resp, rawurl)
	fullname := filepath.Join(e.OutDir, filename)
	if IsAlreadyExistsFile(fullname, checksum) {
		return fullname, nil
	}
	hc := NewHashComparator(checksum)
	fd, err := os.Create(fullname)
	if err != nil {
		return "", err
	}
	defer fd.Close()
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
		return "", err
	}
	if hc != nil {
		if err := hc.IsMatch(); err != nil {
			return "", err
		}
	}
	return fullname, nil
}
