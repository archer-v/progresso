package progresso

import (
	"progresso/units/bytes"
	"testing"
	"time"
)

func TestPrintNoSize(t *testing.T) {
	var s, expect string
	p := Progress{
		Unit:      bytes.BytesIEC,
		Percent:   0,
		Total:     0,
		Speed:     100 * bytes.KibiByte,
		SpeedAvg:  100 * bytes.KibiByte, // 100KiB/sec,
		Remaining: -1,
		Processed: bytes.MebiByte * 10,
		StartTime: time.Now().Add(time.Second * -5),
	}

	expect = "10.00MiB (Speed: 100.00KiB/s / AVG: 100.00KiB/s) (Time: 5 seconds)"
	s = p.String()
	if s != expect {
		t.Log("TestPrintNoSize: full failed:")
		t.Logf("   Got     : '%s'\n", s)
		t.Logf("   Expected: '%s'\n", expect)
		t.Fail()
	}
}

func TestPrintSize(t *testing.T) {
	var s, expect string
	p := Progress{
		Unit:      bytes.BytesIEC,
		Percent:   50.0,
		Total:     bytes.MebiByte * 20,
		Speed:     100 * bytes.KibiByte,
		SpeedAvg:  100 * bytes.KibiByte, // 100KiB/sec,
		Remaining: time.Second * 10,
		Processed: bytes.MebiByte * 10,
		StartTime: time.Now().Add(time.Second * -5),
	}

	expect = "[50.00%] (10.00MiB/20.00MiB) (Speed: 100.00KiB/s / AVG: 100.00KiB/s) (Time: 5 seconds / Remaining: 10 seconds)"
	s = p.String()
	if s != expect {
		t.Log("TestPrintSize: full failed:")
		t.Logf("   Got     : '%s'\n", s)
		t.Logf("   Expected: '%s'\n", expect)
		t.Fail()
	}
	// Test without p.SpeedAvg
	p.SpeedAvg = 0
	p.StartTime = time.Now().Add(time.Second * -5)
	expect = "[50.00%] (10.00MiB/20.00MiB) (Speed: 100.00KiB/s) (Time: 5 seconds / Remaining: 10 seconds)"
	s = p.String()
	if s != expect {
		t.Log("TestPrintSize: without p.SpeedAvg failed:")
		t.Logf("   Got     : '%s'\n", s)
		t.Logf("   Expected: '%s'\n", expect)
		t.Fail()
	}
	// Test without p.Remaining
	p.SpeedAvg = 100 * bytes.KibiByte
	p.Remaining = -1
	p.StartTime = time.Now().Add(time.Second * -5)
	expect = "[50.00%] (10.00MiB/20.00MiB) (Speed: 100.00KiB/s / AVG: 100.00KiB/s) (Time: 5 seconds)"
	s = p.String()
	if s != expect {
		t.Log("TestPrintSize: without p.Remaining failed:")
		t.Logf("   Got     : '%s'\n", s)
		t.Logf("   Expected: '%s'\n", expect)
		t.Fail()
	}
	// Test p.Remaining == 0
	p.Remaining = 0
	p.StartTime = time.Now().Add(time.Second * -5)
	expect = "[50.00%] (10.00MiB/20.00MiB) (Speed: 100.00KiB/s / AVG: 100.00KiB/s) (Time: 5 seconds / Remaining: 0 seconds)"
	s = p.String()
	if s != expect {
		t.Log("TestPrintSize: with p.Remaining == 0 failed:")
		t.Logf("   Got     : '%s'\n", s)
		t.Logf("   Expected: '%s'\n", expect)
		t.Fail()
	}
}

func TestIOProgress(t *testing.T) {
	iop := NewBytesProgressTracker().Size(100 * bytes.MebiByte)
	iop.progress = 50 * bytes.MebiByte
	iop.updatesW[1] = 40 * bytes.MebiByte
	iop.updatesT[1] = time.Now().Add(time.Second * -1)
	iop.startTime = time.Now().Add(time.Second * -10)
	go iop.Update(0)
	p := <-iop.Channel
	t.Logf("P: %p\n", &p)
	t.Logf("P: %s\n", p.String())
	//t.Fail()
}
