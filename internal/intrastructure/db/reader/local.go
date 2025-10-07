package reader

import (
	"io"
	"os"
	"strings"
)

type LocalFileReader struct{}

func (LocalFileReader) Open(path string) (io.ReadCloser, error) {
	if after, ok := strings.CutPrefix(path, "file://"); ok  {
		path = after
	}
	return os.Open(path)
}