# progresso

Go library to track progress of long living operation

Based on progressio library [github.com/bartmeuris/progressio] 

## About

This package was created since most progress packages out-there seem to
directly want to output stuff to a console mixing UI and work logic, 
work in a non-Go way (callback functions), or can only be used for 
specific scenario's like file downloads or Read/Write operations.

This library can be used to track the progress of an operation, 
the amount of work performed of which can be measured as 
an integer number.

The package provides two methods to count the amount of work:
*  wrappers around standard io.Reader and io.Writer objects, so anything that uses standard io.Reader/io.Writer objects can give you progress feedback
*  Increment(int) method to inform the tracker about amount of work performed

ProgressTracker sends back a Progress struct over a channel, 
so a subscriber can receive the progress feedback.

It attempts to do all the heavy lifting for you:

* updates are throttled to a configurable value (by time intervals or volume of work processed)
* formatting the value to configurable measurement unit (bytes, distance, etc)
* Precalculates things like:
  * Speed in unit/sec of the last few operations
  * Average speed in unit/sec since the start of the operation
  * Remaining time
  * Percentage

Some of these statistics are not available if the size was not specified up front.

## API

### ProgressTracker struct

```
type ProgressTracker struct {
	   Channel       chan Progress	
}
```

The  progresso.ProgressTracker object has methods: 
* ```Increment(int64, any)``` - increments the progress at the given amount 
* ```Update(int64, any)``` - updates the tracker with new progress value
* ```Reset()``` - resets the progress tracker to an initial state
* ```Stop()``` - stops the tracker, and sends the last message
* ```GetWriter``` - returns a ProgressTrackerWriter for the progress tracker
* ```GetReader``` - returns a ProgressTrackerReader for the progress tracker

and several setters to set configurable options:
* ```SetSize(size int64)```
* ```SetUpdateFreq(freq time.Duration)``` - sets the frequency of the updates over the channels
* ```SetUpdateGranule(granule int64)``` - sets updates interval in units of work at which to send updates
* ```SetUpdateGranulePercent``` - sets updates interval in percent of work at which to send updates
* ```SetUnit``` - sets the measurement unit of the progress tracker
* ```SetBlock``` - sets blocking write to the Channel to prevent possible lost of messages if channel isn't reading state
* ```SetName``` - sets the name of the progress tracker
* ```SetTimeSlots``` - sets the number of time slots used to calculate an instant speed (default 5)

#### Constructors

* ```NewProgressTracker(units.Unit)``` - creates a new progress tracker with the given measurement unit
* ```NewBytesProgressTracker()``` - creates a new progress tracker with bytes unit
* ```NewProgressTrackerReader(size)``` - creates a new ProgressTracker impelementing io.Reader interface. Specify a size <= 0 if you don't know the size.
* ```NewProgressTrackerWriter(size)``` - creates a new ProgressTracker impelementing io.Writer interface. Specify a size <= 0 if you don't know the size.


### Progress struct

```
type Progress struct {
    Name        string        // The name of the tracker  
    Processed   int64         // The amount of work performed (bytes transfered, for example)
    Total       int64         // Total size of work (bytes to transfer for example). <= 0 if size is unknown.
    Percent     float64       // If the size is known, the progress of the transfer in %
    SpeedAvg    int64         // Bytes/sec average over the entire transfer
    Speed       int64         // Bytes/sec of the last few reads/writes
    Unit        units.Unit    // The unit system is used to format the value (for example to bytes, kilobytes, megabytes, etc)
    Remaining   time.Duration // Estimated time remaining, only available if the size is known.
    StartTime   time.Time     // When the transfer was started
    StopTime    time.Time     // only specified when the transfer is completed: when the transfer was stopped
    Data        any  		  // An additional user defined data associated with the progress
}

```
The progresso.Progress object has at the moment only one method, the
String() function to return the `string` representation of the object.

### Unit struct

Unit struct represents the unit of measure of operation progress

```
type Unit struct {
   Size       int64    // The size of one unit
   Name       string   // The name of the unit standard
   Multiplier int64    // The multiplier used by the unit standard
   Names      []string // The names used by the unit standard
   Shorts     []string // The shortened names used by the unit standard
}
```

Several common units already defined: 
* units.BytesMetric
* units.BytesIEC
* units.DistanceMetric

See units.bytes and unit.distance how to define your own units  

## Example

Copying data with progress tracking using ProgressTrackerWriter implementing io.Writer/Reader interface

```
import (
  "io"
  "github.com/archer-v/progresso"
)

// io.Copy wrapper to specify the size and show copy progress.
func copyProgress(w io.Writer, r io.Reader, size int64) (written int64, err error) {
  
  // Wrap your io.Writer:
  pw, ch := progresso.NewProgressTrackerWriter(w, size)
  defer pw.Close()
  
  // Launch a Go-Routine reading from the progress channel
  go func() {
    for p := range ch {
      fmt.Printf("\rProgress: %s", p.String())
    }
    fmt.Printf("\nDone\n")
  }
  
  // Copy the data from the reader to the new writer
  return io.Copy(pw, r)
}
```

Example of tracking object movement process

```
import (
  "io"
  "github.com/archer-v/progresso"
)

func movement(distance int64) {
  r := NewProgressTracker(distance.DistanceMetric)
  r.SetUpdateGranule(100).SetSize(distance)
  
  // Launch a Go-Routine reading from the progress channel
  go func() {
    for p := range ch {
      fmt.Printf("\rProgress: %s", p.String())
    }
    fmt.Printf("\nDone\n")
  }
  
  for i := 0; i < 200; i++ {
	time.Sleep(20 * time.Millisecond)
	r.Increment(10) // move at 10 meters
  }
}
```
