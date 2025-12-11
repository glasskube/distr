package tmpstream

import (
	"bytes"
	"fmt"
	"io"
)

type inMemoryStream struct{ data []byte }

// Destroy implements TmpStream.
func (i *inMemoryStream) Destroy() error { return nil }

// Get implements TmpStream.
func (i *inMemoryStream) Get() (io.ReadSeekCloser, error) {
	return &noopReadSeekCloser{bytes.NewReader(i.data)}, nil
}

func newInMemoryStream(src io.Reader) (result TmpStream, rerr error) {
	if data, err := io.ReadAll(src); err != nil {
		return nil, fmt.Errorf("failed to read source stream into memory: %w", err)
	} else {
		return &inMemoryStream{data: data}, nil
	}
}

type noopReadSeekCloser struct{ io.ReadSeeker }

func (n *noopReadSeekCloser) Close() error { return nil }
