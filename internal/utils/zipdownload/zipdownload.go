package zipdownload

import (
	"archive/zip"
	"bytes"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/shared/httpClient"
)

// DownloadZipped downloads a zip archive and returns all contained files
func DownloadZipped(url string) (map[string][]byte, error) {
	resp, err := httpClient.Do().R().Get(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := resp.Body()
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	allFiles := make(map[string][]byte)
	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			return allFiles, err
		}
		allFiles[zipFile.Name] = unzippedFileBytes
	}
	return allFiles, nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	return data, errors.WithStack(err)
}
