package progresso

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

const throttleTime = 15 * time.Millisecond
const bufSize = 6000000

type throttleWriter struct {
	lastTime time.Time
}

func (t *throttleWriter) Write(b []byte) (n int, err error) {
	sleep := time.Since(t.lastTime)
	if t.lastTime.IsZero() {
		sleep = 0
	}
	if sleep < throttleTime {
		<-time.After(throttleTime - sleep)
	}
	t.lastTime = time.Now()
	return len(b), nil
}

type throttleReader struct {
	i        io.Reader
	lastTime time.Time
}

func (t *throttleReader) Read(b []byte) (n int, err error) {
	sleep := time.Since(t.lastTime)
	if t.lastTime.IsZero() {
		sleep = 0
	}
	if sleep < throttleTime {
		<-time.After(throttleTime - sleep)
	}
	t.lastTime = time.Now()
	return t.i.Read(b)
}

func getWriter() io.Writer {
	return &throttleWriter{}
}
func getReader(size int) io.Reader {
	ibuf := make([]byte, size)
	return &throttleReader{
		i: bytes.NewBuffer(ibuf),
	}
}

func readProgress(t *testing.T, msg string, ch <-chan Progress) {
	cs := ""
	p := Progress{}
	for p = range ch {
		ps := msg + ": " + p.String()

		if len(cs) < len(ps) {
			cs = strings.Repeat(" ", len(ps))
		}
		t.Logf("\r%s\r%s", cs, ps)
	}
	t.Logf("\n%s\n", p.String())
}

// TestProgressWriter is an example of using the progresso package with an
// io.Writer without knowing the amount of bytes to be processed.
func TestProgressWriter(t *testing.T) {
	r := getReader(bufSize)
	w, ch := NewProgressTrackerWriter(getWriter(), -1)

	go readProgress(t, "TestWriter", ch)

	io.Copy(w, r)
	t.Logf("Copy done\n")
}

// TestProgressWriterSize is an example of using the progresso package with an
// io.Writer while knowing the expected amount of bytes to be processed.
func TestProgressWriterSize(t *testing.T) {
	r := getReader(bufSize)
	w, ch := NewProgressTrackerWriter(getWriter(), bufSize)

	go readProgress(t, "TestWriterSize", ch)

	io.Copy(w, r)
	t.Logf("Copy done\n")
}

// TestProgressReader is an example of using the progresso package with an
// io.Reader without knowing the amount of bytes to be processed.
func TestProgressReader(t *testing.T) {
	r, ch := NewProgressTrackerReader(getReader(bufSize), -1)
	w := getWriter()

	go readProgress(t, "TestReader", ch)

	io.Copy(w, r)
	t.Logf("Copy done\n")
}

// TestProgressReaderSize is an example of using the progresso package with an
// io.Reader while knowing the expected amount of bytes to be processed.
func TestProgressReaderSize(t *testing.T) {
	r, ch := NewProgressTrackerReader(getReader(bufSize), bufSize)
	w := getWriter()
	r.UpdateGranule(bufSize / 10)
	go readProgress(t, "TestReaderSize", ch)

	io.Copy(w, r)
	t.Logf("Copy done\n")
}
