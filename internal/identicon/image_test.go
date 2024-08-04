package identicon

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
)

func ExampleIdentIcon_Jpeg() {
	ig, err := New("", 7, 4)

	if err != nil {
		panic(err)
	}

	username := "yourUsername"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	// File writer
	img, _ := os.Create("./examples/" + username + ".jpg")
	defer img.Close()

	quality := 90

	// Takes the size in pixels, quality and an io.Writer
	ii.Jpeg(300, quality, img) // 300px * 300px
}

func ExampleIdentIcon_Png() {
	ig, err := New("", 7, 4)

	if err != nil {
		panic(err)
	}

	username := "yourUsername"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	// File writer
	img, _ := os.Create("./examples/" + username + ".png")
	defer img.Close()
	// Takes the size in pixels and an io.Writer
	ii.Png(300, img) // 300px * 300px
}

func ExampleIdentIcon_Png_base64Encoded() {
	ig, err := New("", 7, 4)

	if err != nil {
		panic(err)
	}

	username := "yourUsername"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	// File writer
	out := new(bytes.Buffer)
	// Takes the size in pixels and an io.Writer
	ii.Png(300, out) // 300px * 300px

	str := base64.StdEncoding.EncodeToString(out.Bytes())
	fmt.Println(str)
	// Output:
	// iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAIAAAD2HxkiAAAD1UlEQVR4nOzVsY2dUBBA0cX6rVAE/VAW/VAETUwJlrMNLMJ/kTgnJZjhSVfzmZkfoPOnXgDeToQQEyHERAgxEUJMhBATIcRECDERQuxz//na9m9t8gjreSRzq3d+2/9W7t/ZJYSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQogtM1Pv8CDXttcrvMJ6HvUKD+ISQkyEEBMhxEQIMRFCTIQQEyHERAgxEUJMhBATIcRECDERQkyEEBMhxEQIMRFCTIQQEyHERAgxEUJMhBATIcRECDERQkyEEBMhxEQIMRFCTIQQEyHERAgxEUJMhBATIcRECDERQkyEEBMhxJaZqXfg59r2ZO56HslcfnMJISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYp96gf+7tr1e4avW80jmeucncAkhJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiy8zcfL62/YvL9NbzSOZW7/y2/63cv7NLCDERQkyEEBMhxEQIMRFCTIQQEyHERAgxEUJMhBATIcRECDERQkyEEBMhxEQIMRFCTIQQEyHERAgxEUJMhBATIcRECDERQkyEEBMhxEQIMRFCTIQQEyHERAgxEUJMhBATIcRECDERQkyEEFtmpt4BXs0lhJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiH3qBfjn2vZk7noeyVx+cwkhJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiIoSYCCEmQoiJEGIihJgIISZCiIkQYiKEmAghJkKIiRBiy8zUO8CruYQQEyHERAgxEUJMhBATIcRECDERQkyEEPsbAAD//6PdNDxqoCsvAAAAAElFTkSuQmCC
}

func TestIdentIconImage(t *testing.T) {
	ig, _ := New(
		"",
		5,
		1,
	)

	pixels := 10

	ii, _ := ig.Draw("i")
	img := ii.Image(pixels)

	b := img.Bounds()
	s := b.Size()

	if b.Empty() {
		t.Errorf("Image shouldn't be empty: expected %v, actual %v", false, b.Empty())
	}

	if s.X != pixels || s.Y != pixels {
		t.Errorf("Image should be of size %v: expected %v, actual %v", pixels, s.X, pixels)
	}

}

func TestIdentIconImagePng(t *testing.T) {
	ig, _ := New(
		"",
		5,
		1,
	)

	pixels := 10

	ii, _ := ig.Draw("i")
	out := new(bytes.Buffer)
	ii.Png(pixels, out)
	expected := []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 10, 0, 0, 0, 10, 8, 2, 0, 0, 0, 2, 80, 88, 234, 0, 0, 0, 64, 73, 68, 65, 84, 120, 156, 98, 249, 240, 225, 3, 3, 110, 192, 132, 71, 142, 129, 129, 129, 241, 168, 130, 51, 156, 163, 125, 97, 45, 3, 3, 195, 85, 131, 96, 162, 117, 67, 236, 70, 214, 129, 108, 18, 113, 186, 33, 0, 98, 6, 68, 31, 41, 118, 227, 2, 4, 116, 3, 2, 0, 0, 255, 255, 130, 28, 21, 123, 255, 212, 164, 59, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}
	bs := out.Bytes()

	if len(bs) != len(expected) {
		t.Errorf("Incorrect len(bs): expected %v, actual %v", len(expected), len(bs))
	}

	for i, b := range bs {
		if b != expected[i] {
			t.Errorf("Incorrect byte at index %v: expected %v, actual %v", i, expected[i], bs[i])
		}
	}
}

func TestIdentIconImageJpeg(t *testing.T) {
	ig, _ := New(
		"",
		5,
		1,
	)

	pixels := 10

	ii, _ := ig.Draw("i")
	out := new(bytes.Buffer)
	ii.Jpeg(pixels, 100, out)
	expected := []byte{255, 216, 255, 219, 0, 132, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 255, 192, 0, 17, 8, 0, 10, 0, 10, 3, 1, 34, 0, 2, 17, 1, 3, 17, 1, 255, 196, 1, 162, 0, 0, 1, 5, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 16, 0, 2, 1, 3, 3, 2, 4, 3, 5, 5, 4, 4, 0, 0, 1, 125, 1, 2, 3, 0, 4, 17, 5, 18, 33, 49, 65, 6, 19, 81, 97, 7, 34, 113, 20, 50, 129, 145, 161, 8, 35, 66, 177, 193, 21, 82, 209, 240, 36, 51, 98, 114, 130, 9, 10, 22, 23, 24, 25, 26, 37, 38, 39, 40, 41, 42, 52, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 1, 0, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 17, 0, 2, 1, 2, 4, 4, 3, 4, 7, 5, 4, 4, 0, 1, 2, 119, 0, 1, 2, 3, 17, 4, 5, 33, 49, 6, 18, 65, 81, 7, 97, 113, 19, 34, 50, 129, 8, 20, 66, 145, 161, 177, 193, 9, 35, 51, 82, 240, 21, 98, 114, 209, 10, 22, 36, 52, 225, 37, 241, 23, 24, 25, 26, 38, 39, 40, 41, 42, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 130, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 226, 227, 228, 229, 230, 231, 232, 233, 234, 242, 243, 244, 245, 246, 247, 248, 249, 250, 255, 218, 0, 12, 3, 1, 0, 2, 17, 3, 17, 0, 63, 0, 253, 98, 248, 199, 241, 143, 254, 10, 205, 240, 3, 226, 71, 136, 255, 0, 101, 125, 63, 196, 127, 240, 208, 223, 16, 191, 104, 111, 236, 143, 248, 100, 63, 143, 127, 217, 31, 179, 95, 194, 95, 236, 255, 0, 248, 84, 186, 14, 151, 241, 27, 227, 247, 252, 90, 201, 244, 189, 103, 68, 187, 251, 94, 137, 172, 191, 131, 63, 226, 231, 120, 135, 68, 242, 63, 177, 63, 225, 34, 240, 95, 246, 149, 222, 164, 44, 98, 251, 163, 254, 25, 107, 254, 10, 61, 255, 0, 73, 81, 255, 0, 205, 30, 248, 17, 255, 0, 205, 13, 126, 155, 81, 92, 48, 192, 165, 41, 185, 226, 113, 117, 34, 223, 238, 160, 241, 85, 225, 236, 161, 119, 39, 30, 106, 117, 33, 58, 175, 158, 82, 181, 74, 178, 156, 213, 53, 78, 157, 253, 199, 41, 126, 221, 155, 120, 219, 137, 197, 224, 178, 106, 89, 39, 135, 94, 18, 240, 230, 99, 67, 10, 223, 19, 230, 180, 60, 50, 224, 92, 234, 92, 83, 156, 70, 150, 23, 47, 165, 153, 195, 45, 226, 62, 29, 205, 178, 174, 21, 194, 199, 42, 203, 114, 247, 87, 35, 225, 76, 30, 87, 149, 87, 226, 12, 71, 17, 113, 4, 232, 169, 231, 116, 240, 57, 103, 255, 217}
	bs := out.Bytes()

	if len(bs) != len(expected) {
		t.Fatalf("Incorrect len(bs): expected %v, actual %v", len(expected), len(bs))
	}

	for i, b := range bs {
		if b != expected[i] {
			t.Errorf("Incorrect byte at index %v: expected %v, actual %v", i, expected[i], bs[i])
		}
	}
}

func TestIdentIconImageSvg(t *testing.T) {
	ig, _ := New("", 5, 2)

	username := "nullrocks"
	ii, _ := ig.Draw(username)

	out := new(bytes.Buffer)
	err := ii.Svg(300, out)

	if err != nil {
		t.Errorf("Image SVG shouldn't error: expected %v, actual %v", nil, err)
	}
}
