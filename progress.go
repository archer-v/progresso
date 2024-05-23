/*
Package progresso contains io.Reader and io.Writer wrappers to easily get
progress feedback, including speed/sec, average speed, %, time remaining,
size, transferred size, ...  over a channel in a progressio.Progress object.

Important note is that the returned object implements the io.Closer interface
and you have to close the progresso.ProgressTrackerReader and
progresso.ProgressTrackerWriter objects in order to clean everything up.

Usage is pretty simple:

	preader, pchan := progressio.NewProgressTrackerReader(myreader, -1)
	defer preader.Close()
	go func() {
		for p := range pchan {
			fmt.Printf("Progress: %s\n", p.String())
		}
	}
	// read from your new reader object
	io.Copy(mywriter, preader)

A helper function is available that opens a file, determines it's size, and
wraps it's os.File io.Reader object:

	if pr, pc, err := progressio.NewProgressFileReader(myfile); err != nil {
		return err
	} else {
		defer pr.Close()
		go func() {
			for p := range pc{
				fmt.Printf("Progress: %s\n", p.String())
			}
		}
		// read from your new reader object
		io.Copy(mywriter, pr)
	}

A wrapper for an io.WriterCloser is available too, but no helper function
is available to write to an os.File since the target size is not known.
Usually, wrapping the io.Writer is more accurate, since writing potentially
takes up more time and happens last. Useage is similar to wrapping the
io.Reader:

	pwriter, pchan := progressio.NewProgressTrackerWriter(mywriter, -1)
	defer pwriter.Close()
	go func() {
		for p := range pchan {
			fmt.Printf("Progress: %s\n", p.String())
		}
	}
	// write to your new writer object
	io.Copy(pwriter, myreader)

Note that you can also implement your own formatting. See the String() function
implementation or consult the Progress struct layout and documentation
*/
package progresso

import (
	"fmt"
	"progresso/units"
	"time"
)

// Progress is the object sent back over the progress channel.
type Progress struct {
	Processed int64         // The amount of work performed (bytes transfered, for example)
	Total     int64         // Total size of work (bytes to transfer for example). <= 0 if size is unknown.
	Percent   float64       // If the size is known, the progress of the transfer in %
	SpeedAvg  int64         // Bytes/sec average over the entire transfer
	Speed     int64         // Bytes/sec of the last few reads/writes
	Unit      units.Unit    // The unit system
	Remaining time.Duration // Estimated time remaining, only available if the size is known.
	StartTime time.Time     // When the transfer was started
	StopTime  time.Time     // only specified when the transfer is completed: when the transfer was stopped
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