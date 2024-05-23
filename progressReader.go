package progresso

import (
	"io"
	"os"
)

// ProgressTrackerReader is a struct representing an io.ReaderCloser, which sends back progress
// feedback over a channel
type ProgressTrackerReader struct {
	r io.ReadCloser
	ProgressTracker
}

// NewProgressTrackerFileReader creates a new ProgressTrackerReader based on a file. It teturns a
// ProgressTrackerReader object and a channel on success, or an error on failure.
func NewProgressTrackerFileReader(file string) (*ProgressTrackerReader, <-chan Progress, error) {
	f, ferr := os.Open(file)
	if ferr != nil {
		return nil, nil, ferr
	}
	// Get the filesize by seeking to the end of the file, and back to offset 0
	fsize, err := f.Seek(0, 2)
	if err != nil {
		return nil, nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, nil, err
	}
	io, ch := NewProgressTrackerReader(f, fsize)
	return io, ch, nil
}

// NewProgressTrackerReader creates a new ProgressTrackerReader object based on the io.Reader and the
// size you specified. Specify a size <= 0 if you don't know the size.
func NewProgressTrackerReader(r io.Reader, size int64) (*ProgressTrackerReader, <-chan Progress) {
	if r == nil {
		return nil, nil
	}
	rc, ok := r.(io.ReadCloser)
	if !ok {
		rc = io.NopCloser(r)
	}
	ret := &ProgressTrackerReader{rc, *NewBytesProgressTracker().Size(size)}
	return ret, ret.Channel
}

// Read wraps the io.Reader Read function to also update the progress.
func (p *ProgressTrackerReader) Read(b []byte) (n int, err error) {
	n, err = p.r.Read(b)
	p.Update(int64(n))
	return
}

// Close wraps the io.ReaderCloser Close function to clean up everything. ProgressTrackerReader
// objects should always be closed to make sure everything is cleaned up.
func (p *ProgressTrackerReader) Close() (err error) {
	err = p.r.Close()
	p.Stop()
	return
}
