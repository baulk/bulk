package basics

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/baulk/bulk/netutils"
)

// File compress file
type File struct {
	URL         string `json:"url,omitempty"`
	Hash        string `json:"hash,omitempty"`
	Path        string `json:"path,omitempty"`
	Destination string `json:"destination"`
	Name        string `json:"name,omitempty"`        // if not exists use filepath.Base
	Executabled bool   `json:"executabled,omitempty"` // when mark executabled. a script create under windows can run linux
}

func isTrue(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "on" || s == "yes" || s == "1"
}

func outTempDir() string {
	if bulkoutdir := os.Getenv("BULK_DOWNLOAD_OUTDIR"); len(bulkoutdir) != 0 {
		return bulkoutdir

	}
	return os.ExpandEnv("${TEMP}/bulk_download_out")
}

func newExecutor() *netutils.Executor {
	opt := &netutils.Options{
		IsDebugMode:        isTrue(os.Getenv("BULK_DEBUG_TRACE")),
		InsecureSkipVerify: isTrue(os.Getenv("BULK_INSECURE_SKIP_VERIFY")),
		DisableAutoProxy:   isTrue(os.Getenv("BULK_DISABLE_AUTO_PROXY")),
		DestinationPath:    outTempDir(),
	}
	return netutils.NewExecutor(opt)
}

var fileexecutor = newExecutor()

// Prepare check file
func (file *File) Prepare() error {
	if file.Path != "" {
		if _, err := os.Stat(file.Path); err != nil {
			return err
		}
		return nil
	}
	if file.URL == "" {
		return ErrResponseFilesField
	}
	fullpath, err := fileexecutor.WebGet(&netutils.EnhanceURL{URL: file.URL, HashValue: file.Hash})
	if err != nil {
		return err
	}
	file.Path = fullpath
	return nil
}

// BuildPath todo
func (file *File) BuildPath() string {
	destination := filepath.ToSlash(file.Destination)
	if len(file.Name) == 0 {
		return path.Join(destination, filepath.Base(file.Path))
	}
	return path.Join(destination, file.Name)
}

// ResponseFile todo
type ResponseFile struct {
	cwd           string
	Destination   string   `json:"destination"`
	CompressLevel int      `json:"level,omitempty"`
	Method        string   `json:"method,omitempty"`
	Files         []File   `json:"files,omitempty"`
	Dirs          []string `json:"dirs,omitempty"`
}

// DefaultResponseFile default response file
const DefaultResponseFile = "compress.rsp.json"

// NewResponseFile response file
func NewResponseFile(src string) (*ResponseFile, error) {
	fd, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	rsp := ResponseFile{cwd: filepath.Dir(src)}
	if err := json.NewDecoder(fd).Decode(&rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}
