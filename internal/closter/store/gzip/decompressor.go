package gzip

import (
	"compress/gzip"
	"io"

	"github.com/vinceanalytics/vince/internal/logger"
)

// Decompressor is a wrapper around a gzip.Reader that reads from an io.Reader
// and decompresses the data.
type Decompressor struct {
	*gzip.Reader
}

// NewDecompressor returns an instantied Decompressor that reads from r and
// decompresses the data using gzip.
func NewDecompressor(r io.Reader) *Decompressor {
	gr, err := gzip.NewReader(r)
	if err != nil {
		logger.Fail("Failed creating gzip reader", "err", err)
	}
	gr.Multistream(false)
	return &Decompressor{
		Reader: gr,
	}
}
