package identicon

import (
	"fmt"
	"image/color"
	"testing"
)

func Example() {
	ig, err := New(
		"github.com", // namespace
		5,            // number of blocks (size)
		2,            // density of points
	)

	if err != nil {
		panic(err)
	}

	username := "nullrocks"
	ii, err := ig.Draw(username) // Generate an IdentIcon

	if err != nil {
		panic(err)
	}

	// File writer
	// img, _ := os.Create("./examples/" + username + ".png")
	// defer img.Close()
	// Takes the size in pixels and an io.Writer
	// ii.Png(300, img) // 300px * 300px

	fmt.Println(ii.String(" ", "."))
	// Output:
	// . . . . .
	// 1 1 1 1 1
	// 1 . 1 . 1
	// . . 2 . .
	// . . . . .
}

func ExampleGenerator_Draw_size7x7withNamespace() {
	ig, err := New(
		"null.rocks",
		7,
		4,
	)

	if err != nil {
		panic(err)
	}

	username := "admin"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	fmt.Println(ii.GeneratorText())
	fmt.Println(ii.String(" ", "."))
	// Output:
	// admin:null.rocks
	// . 1 2 . 2 1 .
	// . 2 . . . 2 .
	// . . . 1 . . .
	// . . . 1 . . .
	// 2 . . 4 . . 2
	// . . 4 2 4 . .
	// . 1 2 1 2 1 .
}

func ExampleNew_customOptions() {

	alwaysRed := func(cb []byte) color.Color {
		return color.RGBA{255, 0, 0, 255}
	}

	transparentBg := func(cb []byte, fc color.Color) color.Color {
		return color.Transparent
	}

	ig, _ := New(
		"",
		7,
		4,
		SetRandom(true),                 // Resultant image will be random
		SetFillColorFunction(alwaysRed), // Points will be red
		SetBackgroundColorFunction(transparentBg), // Background will be transparent
	)

	// All generated IdentIcons will match configuration (fill=red, bg=transparent, isRandom=true)
	ig.Draw("nullrocks")
	ig.Draw("username")
	ig.Draw("admin")

}

func TestGeneratorNew_success(t *testing.T) {

	ig, err := New(
		"github.com",
		5,
		2,
	)

	if err != nil {
		t.Fatalf("err should be nil: expected %v, actual %v", nil, err)
	}

	if ig == nil {
		t.Fatalf("ig should be valid: expected %v, actual %v", "not nil", ig)
	}
}

func TestGeneratorNew_error(t *testing.T) {

	testData := []struct {
		namespace string
		size      int
		density   int
	}{
		{"", 0, 4},
		{"", 10, 0},
		{"", 0, 0},
	}

	for _, td := range testData {
		ig, err := New(
			td.namespace,
			td.size,
			td.density,
		)

		if err == nil {
			t.Fatalf("err should be a value: expected %v, actual %v", "Error", err)
		}

		if ig != nil {
			t.Fatalf("ii should be nil: expected %v, actual %v", nil, ig)
		}
	}
}

func TestGeneratorDraw(t *testing.T) {
	ig, err := New(
		"github.com",
		5,
		2,
	)

	if err != nil {
		t.Errorf("Generator should be valid: expected %v, actual %v", nil, err)
	}

	ii, err := ig.Draw("nullrocks")

	if err != nil {
		t.Errorf("IdentIcon should be valid: expected %v, actual %v", nil, err)
	}

	str := ii.String("", ".")
	expected := ".....\n11111\n1.1.1\n..2..\n....."
	if str != expected {
		t.Errorf("Unexpected IdentIcon result: expected %v, actual %v", expected, str)
	}
}

func TestGeneratorDraw_size(t *testing.T) {
	size := 7
	ig, _ := New(
		"github.com",
		size,
		2,
	)

	ii, _ := ig.Draw("nullrocks")

	str := ii.String("", ".")
	expected := "11.2.11\n2.....2\n.......\n.......\n.......\n.......\n1.212.1"
	if str != expected {
		t.Errorf("Unexpected IdentIcon result: expected %v, actual %v", expected, str)
	}
}

func TestGeneratorDraw_emptyTextError(t *testing.T) {
	ig, _ := New(
		"github.com",
		5,
		2,
	)

	ii, err := ig.Draw("")

	if err == nil {
		t.Errorf("Error should be returned: expected %v, actual %v", "Error", err)
	}

	if ii != nil {
		t.Errorf("IdentIcon should be nil: expected %v, actual %v", nil, ii)
	}
}
