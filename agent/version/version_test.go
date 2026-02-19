package version

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		// Equal versions
		{"1.0.0", "1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.0", "v1.0.0", 0},

		// v1 > v2
		{"1.1.0", "1.0.0", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.1", "1.0.0", 1},
		{"1.10.0", "1.9.0", 1},

		// v1 < v2
		{"1.0.0", "1.1.0", -1},
		{"1.9.9", "2.0.0", -1},
		{"1.0.0", "1.0.1", -1},
		{"1.9.0", "1.10.0", -1},

		// Different lengths
		{"1.0", "1.0.0", 0},
		{"1.0.1", "1.0", 1},
		{"1.0", "1.0.1", -1},
	}

	for _, tt := range tests {
		result := CompareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("CompareVersions(%q, %q) = %d; want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("GetVersion() returned empty string")
	}
}
