package bytes

import "github.com/archer-v/progresso/units"

// Various constants related to the units
const (
	Byte int64 = 1 // Byte is the representation of a single byte

	KiloByte = Byte * units.MetricMultiplier     // Metric unit "KiloByte" constant
	MegaByte = KiloByte * units.MetricMultiplier // Metric unit MegaByte constant
	GigaByte = MegaByte * units.MetricMultiplier // Metric unit GigaByte constant
	TeraByte = GigaByte * units.MetricMultiplier // Metric unit TerraByte constant
	PetaByte = TeraByte * units.MetricMultiplier // Metric unit PetaByte constant

	KibiByte = Byte * units.IECMultiplier     // IEC standard unit KibiByte constant
	MebiByte = KibiByte * units.IECMultiplier // IEC standard unit MebiByte constant
	GibiByte = MebiByte * units.IECMultiplier // IEC standard unit GibiByte constant
	TebiByte = GibiByte * units.IECMultiplier // IEC standard unit TebiByte constant
	PebiByte = TebiByte * units.IECMultiplier // IEC standard unit PebiByte constant

	JEDECKiloByte = KibiByte // JEDEC uses IEC multipliers, but Metric names, JEDEC KiloByte constant
	JEDECMegaByte = MebiByte // JEDEC uses IEC multipliers, but Metric names, JEDEC MegaByte constant
	JEDECGigaByte = GibiByte // JEDEC uses IEC multipliers, but Metric names, JEDEC GigaByte constant
)

// IECNames is an array containing the unit names for the IEC standards
var _IECNames = []string{
	"byte",
	"kibibyte",
	"mebibyte",
	"gibibyte",
	"tebibyte",
	"pebibyte",
}

// IECShorts is an array containing the shortened unit names for the IEC standard
var _IECShorts = []string{
	"B",
	"KiB",
	"MiB",
	"GiB",
	"TiB",
	"PiB",
}

// JEDECNames is an array containing the unit names for the JEDEC standard
var _JEDECNames = []string{
	"byte",
	"kilobyte",
	"megabyte",
	"gigabyte",
}

// JEDECShorts is an array containing the shortened unit names for the JEDEC standard
var _JEDECShorts = []string{
	"B",
	"KB",
	"MB",
	"GB",
}

// MetricNames is an array containing the unit names for the metric units
var _MetricNames = []string{
	"byte",
	"kilobyte",
	"megabyte",
	"gigabyte",
	"terabyte",
	"petabyte",
}

// MetricShorts is an array containing the shortened unit names for the metric units
var _MetricShorts = []string{
	"B",
	"kB",
	"MB",
	"GB",
	"TB",
	"PB",
}

// BytesMetric is a Unit instance representing bytes in metric system
var BytesMetric = units.Unit{
	Name:       "BytesMetric",
	Size:       Byte,
	Multiplier: units.MetricMultiplier,
	Names:      _MetricNames,
	Shorts:     _MetricShorts,
}

// BytesIEC is a Unit instance representing bytes in IEC standard
var BytesIEC = units.Unit{
	Name:       "BytesIEC",
	Size:       Byte,
	Multiplier: units.IECMultiplier,
	Names:      _IECNames,
	Shorts:     _IECShorts,
}

// BytesJEDEC is a Unit instance representing bytes in JEDEC standard
var BytesJEDEC = units.Unit{
	Name:       "BytesJEDEC",
	Size:       Byte,
	Multiplier: units.IECMultiplier,
	Names:      _JEDECNames,
	Shorts:     _JEDECShorts,
}
