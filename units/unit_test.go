package units

import (
	"math"
	"testing"
)

func Test_getUnit(t *testing.T) {

	var distanceMetric = Unit{
		Name:       "Distance",
		Size:       1,
		Multiplier: MetricMultiplier,
		Names:      []string{"metre", "kilometre"},
		Shorts:     []string{"m", "km"},
	}

	var bytesIEC = Unit{
		Name:       "Bytes",
		Size:       1,
		Multiplier: IECMultiplier,
		Names: []string{
			"byte",
			"kibibyte",
			"mebibyte",
			"gibibyte",
			"tebibyte",
			"pebibyte",
		},
		Shorts: []string{
			"B",
			"KiB",
			"MiB",
			"GiB",
			"TiB",
			"PiB",
		},
	}

	type args struct {
		ss   Unit
		size int64
	}
	tests := []struct {
		name        string
		args        args
		wantDivider int64
		wantName    string
		wantShort   string
	}{
		{
			name:        "IEC 0",
			args:        args{ss: bytesIEC, size: 0},
			wantDivider: 1, wantName: "byte", wantShort: "B",
		},
		{
			name:        "IEC 1234",
			args:        args{ss: bytesIEC, size: 1234},
			wantDivider: 1024, wantName: "kibibyte", wantShort: "KiB",
		},
		{
			name:        "IEC 1234567",
			args:        args{ss: bytesIEC, size: 1234567},
			wantDivider: int64(math.Pow(1024, 2)), wantName: "mebibyte", wantShort: "MiB",
		},
		{ // Uses top element
			name:        "IEC top edge case",
			args:        args{ss: bytesIEC, size: int64(math.Pow(1024, 5))},
			wantDivider: int64(math.Pow(1024, 5)), wantName: "pebibyte", wantShort: "PiB",
		},
		{ // Exceeds top element
			name:        "IEC exceeds top edge case",
			args:        args{ss: bytesIEC, size: IECMultiplier * int64(math.Pow(1024, 5))},
			wantDivider: int64(math.Pow(1024, 5)), wantName: "pebibyte", wantShort: "PiB",
		},
		{
			name:        "Distance 1",
			args:        args{ss: distanceMetric, size: 1},
			wantDivider: 1, wantName: "metre", wantShort: "m",
		},
		{
			name:        "Distance 1000",
			args:        args{ss: distanceMetric, size: 1000},
			wantDivider: 1000, wantName: "kilometre", wantShort: "km",
		},
		{ // There is no mega-metre defined, so should return highest mapped divisor - 1000 for km
			name:        "Distance 1000000",
			args:        args{ss: distanceMetric, size: 1000000},
			wantDivider: 1000, wantName: "kilometre", wantShort: "km",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDivider, gotName, gotShort := tt.args.ss.getUnit(tt.args.size)
			if gotDivider != tt.wantDivider {
				t.Errorf("getUnit() gotDivider = %v, want %v", gotDivider, tt.wantDivider)
			}
			if gotName != tt.wantName {
				t.Errorf("getUnit() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotShort != tt.wantShort {
				t.Errorf("getUnit() gotShort = %v, want %v", gotShort, tt.wantShort)
			}
		})
	}
}
