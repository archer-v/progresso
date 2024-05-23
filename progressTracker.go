package progresso

import (
	"progresso/units"
	"progresso/units/bytes"
	"time"
)

const (
	// DefaultUpdateFreq defines frequency of the updates over the channels
	DefaultUpdateFreq    = 100 * time.Millisecond
	DefaultUpdateGranule = 1
	// timeSlots defines the number of slots in the time slice
	// used to calculate the speed and estimated time to completion
	timeSlots = 5
)

type ProgressTracker struct {
	size          int64
	progress      int64
	unit          units.Unit
	Channel       chan Progress
	closed        bool
	startTime     time.Time
	lastSent      time.Time
	updatesW      []int64
	updatesT      []time.Time
	updateFreq    time.Duration
	updateGranule int64
	ts            int
}

// NewProgressTracker creates a new progress tracker with the given measurement unit
func NewProgressTracker(unit units.Unit) *ProgressTracker {
	return &ProgressTracker{
		Channel:       make(chan Progress),
		unit:          unit,
		size:          -1,
		updateFreq:    DefaultUpdateFreq,
		updateGranule: DefaultUpdateGranule,
		startTime:     time.Time{},
		lastSent:      time.Time{},
		updatesW:      make([]int64, timeSlots),
		updatesT:      make([]time.Time, timeSlots),
	}
}

// NewBytesProgressTracker creates a new progress tracker with bytes unit
func NewBytesProgressTracker() *ProgressTracker {
	return NewProgressTracker(bytes.BytesMetric)
}

// Update updates the progress tracker
// with the given amount of work processed and fires the channel
func (p *ProgressTracker) Update(written int64) {
	if p.closed && p.Channel == nil {
		// Nothing to do
		return
	}
	if written > 0 {
		p.progress += written
	}
	// Throttle sending updated, limit to UpdateFreq - which should be 100ms
	// Always send when finished
	if (time.Since(p.lastSent) < DefaultUpdateFreq) && ((p.size > 0) && (p.progress != p.size)) {
		return
	}
	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}

	prog := Progress{
		Unit:      p.unit,
		StartTime: p.startTime,
		Processed: p.progress,
		Total:     p.size,
	}

	// Calculate current speed based on the last `timeSlots` updates sent
	p.updatesW[p.ts%timeSlots] = p.progress
	p.updatesT[p.ts%timeSlots] = time.Now()
	p.ts++
	if !p.updatesT[p.ts%timeSlots].IsZero() {
		// Calculate the average speed of the last ~2 seconds
		prog.Speed = int64((float64(p.progress-p.updatesW[p.ts%timeSlots]) / float64(time.Since(p.updatesT[p.ts%timeSlots]))) * float64(time.Second))

		// Calculate the average speed since starting the transfer
		tp := time.Since(p.startTime)
		if tp > 0 {
			prog.SpeedAvg = int64((float64(p.progress) / float64(tp)) * float64(time.Second))
		} else {
			prog.SpeedAvg = -1
		}
		if p.size > 0 && prog.SpeedAvg > 0 {
			prog.Remaining = time.Duration((float64(p.size-p.progress) / float64(prog.SpeedAvg)) * float64(time.Second))
		} else {
			prog.Remaining = -1
		}
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
			prog.StopTime = time.Now()
			p.Channel <- prog
			p.cleanup()
		}
		return
	}

	if p.updateGranule > 1 &&
		(p.progress-written)/p.updateGranule == p.progress/p.updateGranule {
		// skip updating the progress if the granule is the same as the previous one
		return
	}

	// Don't force send, only send when it would not block, the chan is non-buffered
	select {
	case p.Channel <- prog:
		// update last sent values
		p.lastSent = time.Now()
	default:
	}

}

func (p *ProgressTracker) cleanup() {
	p.closed = true
	if p.Channel != nil {
		close(p.Channel)
		p.Channel = nil
	}
}

// Stop stops the progress tracker, and sends the last message
func (p *ProgressTracker) Stop() {
	p.closed = true
	p.Update(-1)
}

// Size sets the total size of the work to be done
func (p *ProgressTracker) Size(size int64) *ProgressTracker {
	p.size = size
	return p
}

// UpdateFreq sets the frequency at which to send updates
func (p *ProgressTracker) UpdateFreq(freq time.Duration) *ProgressTracker {
	p.updateFreq = freq
	return p
}

// UpdateGranule sets size of the granule of work at which to send updates
func (p *ProgressTracker) UpdateGranule(granule int64) *ProgressTracker {
	p.updateGranule = granule
	return p
}
