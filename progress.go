package progresso

import (
	"fmt"
	"github.com/archer-v/progresso/units"
	"time"
)

// Progress is the object sent back over the progress channel.
type Progress struct {
	Name        string        `json:"name"`           // The name of the tracker
	Processed   int64         `json:"processed"`      // The amount of work performed (bytes transferred, for example)
	Total       int64         `json:"total"`          // Total size of work (bytes to transfer for example). <= 0 if size is unknown.
	Percent     float64       `json:"percent"`        // If the size is known, the progress of the work in %
	SpeedAvg    int64         `json:"speed_avg"`      // Work/sec average over the entire work
	Speed       int64         `json:"speed"`          // Work/sec of the last few works
	Unit        units.Unit    `json:"unit"`           // The measurement unit system
	Remaining   time.Duration `json:"remaining"`      // Estimated time remaining, only available if the size is known.
	RemainingS  int64         `json:"remaining_s" `   // Estimated time remaining in seconds, only available if the size is known.
	StartTime   time.Time     `json:"start_time"`     // When the work was started
	EstStopTime time.Time     `json:"est_stop_time" ` // estimated stop time: when the work was stopped
	StopTime    time.Time     `json:"stop_time"`      // only specified when the work is completed: when the work was stopped
	Finished    bool          `json:"finished" `      // If the progress was stopped
	Completed   bool          `json:"completed" `     // If the progress was completed
	Data        any           `json:"data"`           // An additional user defined data associated with the progress
}

// String returns a string representation of the progress. It takes into account
// if the size was known, and only tries to display relevant data.
func (p *Progress) String() string {
	timeS := fmt.Sprintf(" (Time: %s", FormatDuration(time.Since(p.StartTime)))
	// Build the Speed string
	speedS := ""
	if p.Speed > 0 {
		speedS = fmt.Sprintf(" (Speed: %s", p.Unit.Format(p.Speed, true)) + "/s"
	}
	if p.SpeedAvg > 0 {
		if len(speedS) > 0 {
			speedS += " / AVG: "
		} else {
			speedS = " (Speed AVG: "
		}
		speedS += p.Unit.Format(p.SpeedAvg, true) + "/s"
	}
	if len(speedS) > 0 {
		speedS += ")"
	}

	if p.Total <= 0 {
		// No size was given, we can only show:
		// - Amount read/written
		// - average speed
		// - current speed
		return fmt.Sprintf("%s%s%s)",
			p.Unit.Format(p.Processed, true),
			speedS,
			timeS,
		)
	}
	// A size was given, we can add:
	// - Percentage
	// - Progress indicator
	// - Remaining time
	timeR := ""
	if p.Remaining >= time.Duration(0) {
		timeR = fmt.Sprintf(" / Remaining: %s", FormatDuration(p.Remaining))
	}

	return fmt.Sprintf("[%02.2f%%] (%s/%s)%s%s%s)",
		p.Percent,
		p.Unit.Format(p.Processed, true),
		p.Unit.Format(p.Total, true),
		speedS,
		timeS,
		timeR,
	)
}
