package fileio

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/oidc-mytoken/server/internal/utils/mulerrors"
)

// SearcherFilesystem is a http.Filesystem that looks through multiple http.
// Filesystems for a file returning the first one found
type SearcherFilesystem []http.FileSystem

type SearcherFile []http.File

func (sf SearcherFile) collectErrors(callback func(file http.File) error) error {
	err := mulerrors.Errors{}
	for _, f := range sf {
		e := callback(f)
		if e != nil {
			err = append(err, e)
		}
	}
	if len(err) == 0 {
		err = nil
	}
	return err
}

// Close implements the http.File interface
func (sf SearcherFile) Close() error {
	return sf.collectErrors(
		func(file http.File) error {
			return file.Close()
		},
	)
}

// Read implements the http.File interface
func (sf SearcherFile) Read(p []byte) (n int, err error) {
	for _, f := range sf {
		buffer := make([]byte, len(p))
		n, err = f.Read(buffer)
		if err == nil {
			copy(p, buffer)
			return
		}
	}
	return
}

// Seek implements the http.File interface
func (sf SearcherFile) Seek(offset int64, whence int) (i int64, e error) {
	for _, f := range sf {
		i, e = f.Seek(offset, whence)
		if e == nil {
			return
		}
	}
	return
}

// Readdir implements the http.File interface
func (sf SearcherFile) Readdir(count int) (infos []fs.FileInfo, err error) {
	for _, f := range sf {
		if count > 0 && len(infos) >= count {
			break
		}
		info, e := f.Readdir(count)
		if e != nil {
			continue
		}
		for _, i := range info {
			if count > 0 && len(infos) >= count {
				break
			}
			found := false
			for _, ii := range infos {
				if i.Name() == ii.Name() {
					found = true
					break
				}
			}
			if !found {
				infos = append(infos, i)
			}
		}
	}
	return
}

// Stat implements the http.File interface
func (sf SearcherFile) Stat() (i fs.FileInfo, e error) {
	for _, f := range sf {
		i, e = f.Stat()
		if e == nil {
			return
		}
	}
	return
}

// Open implements the http.FileSystem interface
func (lfs SearcherFilesystem) Open(name string) (http.File, error) {
	sf := SearcherFile{}
	for _, ffs := range lfs {
		if ffs == nil {
			continue
		}
		f, err := ffs.Open(name)
		if err != nil {
			continue
		}
		stat, err := f.Stat()
		if err != nil {
			continue
		}
		if stat.IsDir() {
			sf = append(sf, f)
		} else {
			return f, nil
		}
	}
	var err error
	if len(sf) == 0 {
		err = os.ErrNotExist
	}
	return sf, err
}

// NewLocalAndOtherSearcherFilesystem creates a SearcherFilesystem from a local basePath (
// if not empty) and other http.FileSystems. The local fs will be the first one.
func NewLocalAndOtherSearcherFilesystem(basePath string, other ...http.FileSystem) SearcherFilesystem {
	if basePath != "" {
		other = append([]http.FileSystem{http.Dir(basePath)}, other...)
	}
	return other
}
