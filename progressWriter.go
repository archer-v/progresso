package progresso

import "io"

// Copy functionality of io.NopCloser, but for Writers
type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
func getNopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{w}
}

// ProgressTrackerWriter is a struct representing an io.WriterCloser, which sends back progress
// feedback over a channel
type ProgressTrackerWriter struct {
	w io.WriteCloser
	ProgressTracker
}

// NewProgressTrackerWriter creates a new ProgressTrackerWriter object based on the io.Writer and the
// size you specified. Specify a size <= 0 if you don't know the size.
func NewProgressTrackerWriter(w io.Writer, size int64) (*ProgressTrackerWriter, <-chan Progress) {
	if w == nil {
		return nil, nil
	}
	wc, ok := w.(io.WriteCloser)
	if !ok {
		wc = getNopWriteCloser(w)
	}
	ret := &ProgressTrackerWriter{wc, *NewBytesProgressTracker().Size(size)}
	return ret, ret.Channel
}

// Write wraps the io.Writer Write function to also update the progress.
func (p *ProgressTrackerWriter) Write(b []byte) (n int, err error) {
	n, err = p.w.Write(b[0:])
	p.Update(int64(n))
	return
}

// Close wraps the io.WriterCloser Close function to clean up everything. ProgressTrackerWriter
// objects should always be closed to make sure everything is cleaned up.
func (p *ProgressTrackerWriter) Close() (err error) {
	err = p.w.Close()
	p.Stop()
	return
}
