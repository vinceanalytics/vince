package identicon

import (
	"image/color"
	"testing"
)

func TestGeneratorHashFunction(t *testing.T) {
	digest := _sha256([]byte("nullrocks"))
	expected := []byte{223, 188, 30, 226, 43, 144, 46, 172, 218, 208, 158, 51, 4, 71, 132, 85, 19, 106, 214, 201, 130, 122, 193, 17, 220, 101, 213, 93, 164, 86, 106, 83}
	if len(digest) != 32 {
		t.Errorf("Incorrect len(digest): expected %v, actual %v", 32, len(digest))
	}

	for i, b := range digest {
		if b != expected[i] {
			t.Errorf("Incorrect byte at index %v: expected %v, actual %v", i, expected[i], digest[i])
		}
	}
}

func TestGeneratorFillColorFunction(t *testing.T) {
	type testS struct {
		digest []byte
		hsl    color.Color
	}
	tt := []testS{
		{
			digest: []byte{223, 188, 30}, // normalized saturation and lightness
			hsl:    HSL{51, 80, 28},
		},
		{
			digest: []byte{123, 123, 123},
			hsl:    HSL{246, 63, 43},
		},
		{
			digest: []byte{3, 3, 3},
			hsl:    HSL{6, 63, 43},
		},
		{
			digest: []byte{0, 0, 0},
			hsl:    HSL{0, 60, 40},
		},
		{
			digest: []byte{255, 255, 255},
			hsl:    HSL{150, 75, 45},
		},
	}

	for i, ut := range tt {
		fc := _fillColor(ut.digest)
		if fc != ut.hsl {
			t.Errorf("Unexpected color at index %v: expected %v, actual %v", i, ut.hsl, fc)
		}
	}
}

func TestGeneratorBackgroundColorFunction(t *testing.T) {
	// This is completely be unnecessary at the moment. I look forward to
	// the moment that the background color is dynamically generated.
	expected := color.NRGBA{R: 240, G: 240, B: 240, A: 255}

	type testS struct {
		digest []byte
		fill   color.Color
	}
	tt := []testS{
		{
			digest: []byte{223, 188, 30}, // normalized saturation and lightness
			fill:   HSL{51, 80, 28},
		},
		{
			digest: []byte{123, 123, 123},
			fill:   HSL{246, 63, 43},
		},
		{
			digest: []byte{3, 3, 3},
			fill:   HSL{6, 63, 43},
		},
	}

	for i, ut := range tt {
		fc := _backgroundColor(ut.digest, ut.fill)
		if fc != expected {
			t.Errorf("Unexpected color at index %v: expected %v, actual %v", i, expected, fc)
		}
	}
}
