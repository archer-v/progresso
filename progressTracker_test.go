package progresso

import (
	"bytes"
	"github.com/archer-v/progresso/units/distance"
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

func readProgress(t *testing.T, msg string, ch <-chan Progress, done chan<- struct{}) {
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

	close(done)
}

// TestProgressWriter is an example of using the progresso package with an
// io.Writer without knowing the amount of bytes to be processed.
func TestProgressWriter(t *testing.T) {
	r := getReader(bufSize)
	w, ch := NewProgressTrackerWriter(getWriter(), -1)

	done := make(chan struct{})
	go readProgress(t, "TestWriter", ch, done)

	io.Copy(w, r)
	w.Close()
	<-done
	t.Logf("Copy done\n")
}

// TestProgressWriterFromTracker is an example of using the progresso package with an
// io.Writer without knowing the amount of bytes to be processed.
func TestProgressWriterFromTracker(t *testing.T) {
	r := getReader(bufSize)
	p := NewBytesProgressTracker()
	w := p.GetWriter(getWriter(), -1)

	done := make(chan struct{})
	go readProgress(t, "TestWriter", w.Channel, done)

	io.Copy(w, r)
	w.Close()
	<-done
	t.Logf("Copy done\n")
}

// TestProgressWriterSize is an example of using the progresso package with an
// io.Writer while knowing the expected amount of bytes to be processed.
func TestProgressWriterSize(t *testing.T) {
	r := getReader(bufSize)
	w, ch := NewProgressTrackerWriter(getWriter(), bufSize)

	done := make(chan struct{})
	go readProgress(t, "TestWriterSize", ch, done)

	io.Copy(w, r)
	<-done
	t.Logf("Copy done\n")
}

// TestProgressReader is an example of using the progresso package with an
// io.Reader without knowing the amount of bytes to be processed.
func TestProgressReader(t *testing.T) {
	r, ch := NewProgressTrackerReader(getReader(bufSize), -1)
	w := getWriter()

	done := make(chan struct{})
	go readProgress(t, "TestReader", ch, done)

	io.Copy(w, r)
	r.Close()
	<-done
	t.Logf("Copy done\n")
}

// TestProgressReaderSize is an example of using the progresso package with an
// io.Reader while knowing the expected amount of bytes to be processed.
func TestProgressReaderSize(t *testing.T) {
	r, ch := NewProgressTrackerReader(getReader(bufSize), bufSize)
	w := getWriter()
	r.SetUpdateGranule(bufSize / 10)
	done := make(chan struct{})
	go readProgress(t, "TestReaderSize", ch, done)

	io.Copy(w, r)
	<-done
	t.Logf("Copy done\n")
}

func TestProgressTrackerDistance(t *testing.T) {
	r := NewProgressTracker().SetUnit(distance.DistanceMetric)

	r.SetUpdateGranule(100).SetSize(2000)
	done := make(chan struct{})
	go readProgress(t, "TestDistance", r.Channel, done)

	for i := 0; i < 200; i++ {
		time.Sleep(throttleTime)
		r.Increment(10) // 10 meters
	}
	<-done
	t.Logf("done\n")
}

func TestProgressTrackerDistancePercentGranule(t *testing.T) {
	r := NewProgressTracker().SetUnit(distance.DistanceMetric)

	r.SetUpdateGranulePercent(10).SetSize(2000)
	done := make(chan struct{})
	go readProgress(t, "TestDistance", r.Channel, done)

	for i := 0; i < 200; i++ {
		time.Sleep(throttleTime)
		r.Increment(10) // 10 meters
	}
	<-done
	t.Logf("done\n")
}
