package distance

import "progresso/units"

var DistanceMetric = units.Unit{
	Name:       "Distance",
	Size:       1,
	Multiplier: 1000,
	Names:      []string{"metre", "kilometre"},
	Shorts:     []string{"m", "km"},
}
