package progresso

import (
	"github.com/archer-v/progresso/units"
	"github.com/archer-v/progresso/units/bytes"
	"io"
	"sync"
	"time"
)

const (
	// DefaultUpdateFreq defines frequency of the updates over the channels
	DefaultUpdateFreq    = 100 * time.Millisecond
	DefaultUpdateGranule = 1
	// DefaultTimeSlots defines the number of slots in the time slice
	// used to calculate an instant speed
	DefaultTimeSlots = 5
)

type ProgressTracker struct {
	name                 string
	size                 int64
	progress             int64
	unit                 units.Unit
	Channel              chan Progress
	closed               bool
	startTime            time.Time
	lastSent             time.Time
	updatesW             []int64     // list of last work updates
	updatesT             []time.Time // list of last time updates
	timeSlots            int
	updateFreq           time.Duration
	updateGranule        int64
	updateGranulePercent int
	updatesCounter       int // counter of updates
	sync.Mutex
}

// NewProgressTracker creates a new progress tracker with the given measurement unit
func NewProgressTracker() (p *ProgressTracker) {
	p = &ProgressTracker{
		Channel:       make(chan Progress),
		size:          -1,
		updateFreq:    DefaultUpdateFreq,
		updateGranule: DefaultUpdateGranule,
		timeSlots:     DefaultTimeSlots,
	}
	p.Reset()
	return
}

// NewBytesProgressTracker creates a new progress tracker with bytes unit
func NewBytesProgressTracker() *ProgressTracker {
	return NewProgressTracker().SetUnit(bytes.BytesMetric)
}

// Increment increments the progress tracker
// at the given amount of work processed and fires the channel
// data is optional and will be exposed as the Data field in the progress object
func (p *ProgressTracker) Increment(progress int64, data ...any) {
	p.Lock()
	defer p.Unlock()

	p.increment(progress, data...)
}

// Update updates the tracker with new progress value
// data is optional and will be exposed as the Data field in the progress object
func (p *ProgressTracker) Update(progress int64, data ...any) {
	p.Lock()
	defer p.Unlock()
	if progress > p.progress {
		p.increment(progress-p.progress, data...)
		return
	}
	// Updates in the past isn't allowed now
}

func (p *ProgressTracker) increment(progress int64, data ...any) {

	if p.closed && p.Channel == nil {
		// Nothing to do
		return
	}

	if progress > 0 {
		p.progress += progress
	}

	// Throttle sending updated, limit to updateFreq
	// Always send when finished
	if time.Since(p.lastSent) < p.updateFreq && !p.closed {
		if (p.size <= 0) || (p.size > 0 && p.progress < p.size) {
			return
		}
	}

	curTime := time.Now()
	if p.startTime.IsZero() {
		p.startTime = curTime
	}

	prog := Progress{
		Name:      p.name,
		Unit:      p.unit,
		StartTime: p.startTime,
		Processed: p.progress,
		Total:     p.size,
	}

	if data != nil && len(data) > 0 {
		prog.Data = data[0]
	}

	if p.updatesW == nil {
		p.updatesW = make([]int64, p.timeSlots)
		p.updatesT = make([]time.Time, p.timeSlots)
	}

	// saves update data to the current slot
	p.updatesW[p.updatesCounter%p.timeSlots] = p.progress
	p.updatesT[p.updatesCounter%p.timeSlots] = curTime
	p.updatesCounter++

	if !p.updatesT[p.updatesCounter%p.timeSlots].IsZero() {

		// Calculate the average speed since starting the transfer
		tp := time.Since(p.startTime)
		if tp > 0 {
			prog.SpeedAvg = int64((float64(p.progress) / float64(tp)) * float64(time.Second))
		} else {
			prog.SpeedAvg = -1
		}

		// Calculate the remaining time
		if p.size > 0 && prog.SpeedAvg > 0 {
			prog.Remaining = time.Duration((float64(p.size-p.progress) / float64(prog.SpeedAvg)) * float64(time.Second))
		} else {
			prog.Remaining = -1
		}

		// Calculate the average speed of the last updateFreq * p.timeSlots seconds
		prog.Speed = int64(
			(float64(p.progress-p.updatesW[p.updatesCounter%p.timeSlots]) /
				float64(time.Since(p.updatesT[p.updatesCounter%p.timeSlots]))) *
				float64(time.Second))

	} else {
		prog.Speed = -1
		prog.SpeedAvg = -1
		prog.Remaining = -1
	}

	// Calculate the percentage only if we have a size
	if p.size > 0 {
		prog.Percent = float64(int64((float64(p.progress)/float64(p.size))*10000.0)) / 100.0
	}

	if p.closed || (p.size >= 0 && p.progress >= p.size) {
		// EOF or closed, we have to send this last message, and then close the chan
		// Prevent sending the last message multiple times
		if p.Channel != nil {
			prog.StopTime = curTime
			p.send(prog)
			p.cleanup()
		}
		return
	}

	// filter updates except the first one
	if p.updatesCounter > 1 {
		pp := p.updatesW[(p.updatesCounter-2)%p.timeSlots]

		// do not send updates if the progress is the same as
		// the previous one
		if p.progress == pp {
			return
		}

		// skip updating the progress if the granule is the same as the previous one
		if p.updateGranule > 1 {
			// previous progress
			if pp/p.updateGranule == p.progress/p.updateGranule {
				return
			}
		}

		// skip updating the progress if the granule in percent is the same as the previous one
		if p.size > 0 && p.updateGranulePercent > 0 {
			// prev percent
			ppt := int(float64(int64((float64(pp)/float64(p.size))*10000.0)) / 100.0)
			if ppt/p.updateGranulePercent == int(prog.Percent)/p.updateGranulePercent {
				return
			}
		}
	}
	p.send(prog)
}

func (p *ProgressTracker) cleanup() {
	p.closed = true
	if p.Channel != nil {
		close(p.Channel)
		p.Channel = nil
	}
}

func (p *ProgressTracker) send(prog Progress) {
	// Don't force send, only send when it would not block, the chan is non-buffered
	select {
	case p.Channel <- prog:
		// update last sent values
		p.lastSent = time.Now()
	default:
	}
}

// Reset resets the progress tracker to an initial state
func (p *ProgressTracker) Reset() {
	p.Lock()
	defer p.Unlock()
	p.progress = 0 // reset progress
	p.startTime = time.Time{}
	p.lastSent = time.Time{}
	p.updatesW = nil
	p.updatesT = nil
	p.updatesCounter = 0
}

// Stop stops the progress tracker, and sends the last message
func (p *ProgressTracker) Stop() {
	p.Lock()
	defer p.Unlock()
	p.closed = true
	p.increment(-1)
}

// GetWriter returns a ProgressTrackerWriter for the progress tracker
func (p *ProgressTracker) GetWriter(w io.Writer, size int64) *ProgressTrackerWriter {
	t, _ := newProgressTrackerWriter(w, size, p)
	return t
}

// GetReader returns a ProgressTrackerReader for the progress tracker
func (p *ProgressTracker) GetReader(r io.Reader, size int64) *ProgressTrackerReader {
	t, _ := newProgressTrackerReader(r, size, p)
	return t
}

// SetSize sets the total size of the work to be done
func (p *ProgressTracker) SetSize(size int64) *ProgressTracker {
	p.Lock()
	defer p.Unlock()
	p.size = size
	return p
}

// SetUpdateFreq sets the frequency at which to send updates
func (p *ProgressTracker) SetUpdateFreq(freq time.Duration) *ProgressTracker {
	p.Lock()
	defer p.Unlock()
	p.updateFreq = freq
	return p
}

// SetUpdateGranule sets size of the granule of work at which to send updates
func (p *ProgressTracker) SetUpdateGranule(granule int64) *ProgressTracker {
	p.Lock()
	defer p.Unlock()
	p.updateGranule = granule
	return p
}

// SetUpdateGranulePercent sets updates interval in percent of work at which to send updates
func (p *ProgressTracker) SetUpdateGranulePercent(percent int) *ProgressTracker {
	p.Lock()
	defer p.Unlock()
	p.updateGranulePercent = percent
	return p
}

// SetTimeSlots sets the number of time slots used to calculate an instant speed
func (p *ProgressTracker) SetTimeSlots(slots int) *ProgressTracker {
	p.Lock()
	defer p.Unlock()
	p.timeSlots = slots
	p.updatesW = nil
	p.updatesT = nil
	return p
}

// SetName sets the name of the progress tracker
func (p *ProgressTracker) SetName(name string) *ProgressTracker {
	p.Lock()
	defer p.Unlock()
	p.name = name
	return p
}

// SetUnit sets the measurement unit of the progress tracker
func (p *ProgressTracker) SetUnit(u units.Unit) *ProgressTracker {
	p.Lock()
	defer p.Unlock()
	p.unit = u
	return p
}
