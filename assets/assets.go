package assets

import (
	"net/http"
	"os"
	"time"

	"github.com/rakyll/statik/fs"

	_ "github.com/Qitmeer/qitmeer-wallet/assets/statik"
)

// GetStatic return statci binary fileSystem
func GetStatic() (http.FileSystem, error) {
	return fs.New()
}

//
//
//
//
//

// FilterFunc file func
type FilterFunc func() []byte

// MyStatic add filter
type MyStatic struct {
	fs      http.FileSystem
	filters map[string]FilterFunc
}

// NewMyStatic make
func NewMyStatic(fs http.FileSystem) *MyStatic {
	return &MyStatic{
		fs:      fs,
		filters: make(map[string]FilterFunc),
	}
}

// Open http.filestraem interface
func (s *MyStatic) Open(name string) (http.File, error) {
	if f, ok := s.filters[name]; ok {
		return NewHTTPFile(name, f()), nil
	}
	return s.fs.Open(name)
}

//AddFilter add name f
func (s *MyStatic) AddFilter(name string, f FilterFunc) {
	s.filters[name] = f
	return
}

// NewHTTPFile make
func NewHTTPFile(name string, data []byte) *HTTPFile {
	hf := &HTTPFile{}
	hf.name = name
	hf.data = data
	return hf
}

// HTTPFile http.File
type HTTPFile struct {
	FI
}

// Close err
func (h *HTTPFile) Close() error {
	return nil
}

// Read 6
func (h *HTTPFile) Read(p []byte) (n int, err error) {
	n = copy(p, h.data)
	return
}

// Seek no
func (h *HTTPFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

// Readdir no
func (h *HTTPFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

// Stat 1
func (h *HTTPFile) Stat() (os.FileInfo, error) {
	return h, nil
}

//--------

var fiTime = time.Now()

// FI os.FileInfo
type FI struct {
	name string
	data []byte
}

// Name file name path 3
func (f *FI) Name() string {
	return f.name
}

// Size data 5
func (f *FI) Size() int64 {
	return int64(len(f.data))
}

// Mode 0666
func (f *FI) Mode() os.FileMode {
	return os.ModePerm
}

// ModTime filetime 4
func (f *FI) ModTime() time.Time {
	return fiTime
}

//IsDir dir 2
func (f *FI) IsDir() bool {
	return false
}

// Sys s
func (f *FI) Sys() interface{} {
	return nil
}
