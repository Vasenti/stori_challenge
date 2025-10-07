package ports

import "io"

type Reader interface {
	Open(path string) (io.ReadCloser, error)
}