package units

import "fmt"

const (
	MetricMultiplier = 1000 // Metric uses 1 10^3 multiplier
	IECMultiplier    = 1024 // IEC Standard multiplier, 1024 based
)

// Unit is a structure representing a unit standard
type Unit struct {
	Size       int64    // The size of one unit
	Name       string   // The name of the unit standard
	Multiplier int64    // The multiplier used by the unit standard
	Names      []string // The names used by the unit standard
	Shorts     []string // The shortened names used by the unit standard
}

func (ss Unit) getUnit(size int64) (divider int64, name, short string) {
	if size < 0 {
		size = -size
	}
	if size == 0 {
		return 1, ss.Names[0], ss.Shorts[0]
	}
	div := ss.Size
	if div == 0 {
		div = 1
	}
	for i := 0; i < len(ss.Names); i++ {
		//fmt.Printf("TEST[%d]: %d / DIV: %d | %s / %s\n", i, size, div, ss.Names[i], ss.Shorts[i])
		if div <= size {
			div *= ss.Multiplier
			continue
		}
		return div / ss.Multiplier, ss.Names[i-1], ss.Shorts[i-1]
	}
	return div / ss.Multiplier, ss.Names[len(ss.Names)-1], ss.Shorts[len(ss.Shorts)-1]
}

// Format formats a number of bytes using the given unit standard system.
// If the 'short' flag is set to true, it uses the shortened names.
func (ss Unit) Format(size int64, short bool) string {
	div, name, shortnm := ss.getUnit(size)
	ds := float64(size) / float64(div)
	numfm := "%.2f"
	if div == 1 {
		numfm = "%.0f"
	}
	if short {
		return fmt.Sprintf(numfm+"%s", ds, shortnm)
	}
	return fmt.Sprintf(numfm+" %s", ds, name)
}
