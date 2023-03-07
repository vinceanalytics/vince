package identicon

import (
	"image/color"
	"testing"
)

func TestIdentIconHSL(t *testing.T) {

	type rgba struct {
		r uint32
		g uint32
		b uint32
		a uint32
	}
	type testS struct {
		hsl      color.Color
		expected rgba
	}
	tt := []testS{
		{
			hsl:      HSL{H: 51, S: 80, L: 28},
			expected: rgba{r: 32896, g: 28527, b: 3598, a: 65535},
		},
		{
			hsl:      HSL{H: 246, S: 63, L: 43},
			expected: rgba{r: 13878, g: 10280, b: 45746, a: 65535},
		},
		{
			hsl:      HSL{H: 6, S: 63, L: 43},
			expected: rgba{r: 45746, g: 13878, b: 10280, a: 65535},
		},
		{
			hsl:      HSL{H: 0, S: 60, L: 40},
			expected: rgba{r: 41891, g: 10280, b: 10280, a: 65535},
		},
		{
			hsl:      HSL{H: 150, S: 75, L: 45},
			expected: rgba{r: 7196, g: 51400, b: 29298, a: 65535},
		},
		{
			hsl:      HSL{H: 150, S: 75, L: 45},
			expected: rgba{r: 7196, g: 51400, b: 29298, a: 65535},
		},
		{
			hsl:      HSL{H: 160, S: 40, L: 0},
			expected: rgba{r: 0, g: 0, b: 0, a: 65535},
		},
		{
			hsl:      HSL{H: 200, S: 43, L: 100},
			expected: rgba{r: 65535, g: 65535, b: 65535, a: 65535},
		},
		{
			hsl:      HSL{H: 0, S: 0, L: 0},
			expected: rgba{r: 0, g: 0, b: 0, a: 65535},
		},
		{
			hsl:      HSL{H: 360, S: 100, L: 100},
			expected: rgba{r: 65535, g: 65535, b: 65535, a: 65535},
		},
	}

	for i, ut := range tt {
		r, g, b, a := ut.hsl.RGBA()
		if r != ut.expected.r || g != ut.expected.g || b != ut.expected.b || a != ut.expected.a {
			t.Errorf("Unexpected color at index %v: expected %v, actual %v", i, ut.expected, rgba{r, g, b, a})
		}
	}

}
