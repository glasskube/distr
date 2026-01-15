package tmpstream

import (
	"io"

	"github.com/distr-sh/distr/internal/env"
)

// TmpStream represents a resource that can be accessed via io interfaces and destroyed if no longer needed.
type TmpStream interface {
	Get() (io.ReadSeekCloser, error)
	Destroy() error
}

// New creates a new TmpStream that makes an [io.Reader] seekable by either buffering it in memory or writing it to a
// temporary file, depending on whether [env.RegistryScratchDir] is set.
func New(src io.Reader) (TmpStream, error) {
	if dir := env.RegistryScratchDir(); dir == nil {
		return newInMemoryStream(src)
	} else {
		return newTempFileStream(*dir, src)
	}
}
