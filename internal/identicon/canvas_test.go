package identicon

import (
	"fmt"
	"testing"
)

func ExampleIdentIcon_Array() {
	ig, err := New("github.com", 7, 4)

	if err != nil {
		panic(err)
	}

	username := "nullrocks"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	// Array representation of the IdentIcon
	arr := ii.Array()
	fmt.Print(arr)
	// Output:
	// [[3 1 0 2 0 1 3] [5 1 0 1 0 1 5] [2 2 0 0 0 2 2] [0 0 0 0 0 0 0] [0 0 0 0 0 0 0] [0 0 0 0 0 0 0] [1 0 2 1 2 0 1]]
}

func ExampleIdentIcon_Points() {
	ig, err := New("github.com", 7, 4)

	if err != nil {
		panic(err)
	}

	username := "userName123"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	// Array of image.Points representation of the IdentIcon
	ps := ii.Points()
	fmt.Print(ps)
}

func ExampleIdentIcon_String() {
	ig, err := New("github.com", 7, 4)

	if err != nil {
		panic(err)
	}

	username := "userName12345"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	// String representation of the IdentIcon
	// separator = " "
	// fill empty blocks = "."
	str := ii.String(" ", ".")
	fmt.Print(str)
	// Output:
	// . 2 4 . 4 2 .
	// 1 1 . . . 1 1
	// . . 2 2 2 . .
	// 1 . . 1 . . 1
	// . . . . . . .
	// . . . . . . .
	// . 2 1 1 1 2 .
}

func TestIdentIconCanvas(t *testing.T) {
	size := 7
	ig, _ := New(
		"github.com",
		size,
		2,
	)

	ii, _ := ig.Draw("nullrocks")

	pmlen := len(ii.Canvas.PointsMap)
	if pmlen != 3 {
		t.Errorf("Canvas shouldn't be empty: expected %v, actual %v", 3, pmlen)
	}

	if ii.Canvas.Size != size {
		t.Errorf("Canvas Size should be %v: expected %v, actual %v", size, size, ii.Size)
	}

}

func TestIdentIconCanvasString_generator(t *testing.T) {
	ii := generatedCanvas()
	str := ii.Canvas.String("|", ".")

	expected := ".|.|.|.|.\n1|1|1|1|1\n1|.|1|.|1\n.|.|2|.|.\n.|.|.|.|."

	if str != expected {
		t.Errorf("Unexpected Canvas String result: expected %v, actual %v", expected, str)
	}
}

func TestIdentIconCanvasString_custom(t *testing.T) {
	c := customCanvas()
	str := c.String("|", ".")
	expected := "1|1|.|1|1\n.|.|.|.|.\n.|.|.|.|.\n.|1|1|1|.\n.|.|.|.|."

	if str != expected {
		t.Errorf("Unexpected Canvas String result: expected %v, actual %v", expected, str)
	}
}

func TestIdentIconCanvasArray_generator(t *testing.T) {
	ii := generatedCanvas()
	arr := ii.Canvas.Array()
	expected := [][]int{{0, 0, 0, 0, 0}, {1, 1, 1, 1, 1}, {1, 0, 1, 0, 1}, {0, 0, 2, 0, 0}, {0, 0, 0, 0, 0}}

	if len(arr) != len(expected) {
		t.Fatalf("Unexpected Canvas Array length: expected %v, actual %v", len(expected), len(arr))
	}

	for i, y := range arr {
		for j, x := range y {
			if x != expected[i][j] {
				t.Errorf("Incorrect Array Value at index %v: expected %v, actual %v", []int{i, j}, expected[i][j], x)
			}
		}
	}
}

func TestIdentIconCanvasArray_custom(t *testing.T) {
	c := customCanvas()
	arr := c.Array()
	expected := [][]int{{1, 1, 0, 1, 1}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 1, 1, 1, 0}, {0, 0, 0, 0, 0}}

	if len(arr) != len(expected) {
		t.Fatalf("Unexpected Canvas Array length: expected %v, actual %v", len(expected), len(arr))
	}

	for i, y := range arr {
		for j, x := range y {
			if x != expected[i][j] {
				t.Errorf("Incorrect Array Value at index %v: expected %v, actual %v", []int{i, j}, expected[i][j], x)
			}
		}
	}
}

func TestIdentIconCanvasPoints_generator(t *testing.T) {
	ii := generatedCanvas()
	ps := ii.Canvas.Points()

	if len(ps) != ii.Canvas.FilledPoints {
		t.Fatalf("Unexpected Canvas Points length: expected %v, actual %v", ii.Canvas.FilledPoints, len(ps))
	}

	for _, p := range ps {
		if _, exists := ii.Canvas.PointsMap[p.Y][p.X]; !exists {
			t.Errorf("Point should exist in Canvas at %v: expected %v, actual %v", p, true, exists)
		}
	}
}

func TestIdentIconCanvasPoints_custom(t *testing.T) {
	c := customCanvas()
	ps := c.Points()

	if len(ps) != c.FilledPoints {
		t.Fatalf("Unexpected Canvas Points length: expected %v, actual %v", c.FilledPoints, len(ps))
	}

	for _, p := range ps {
		if _, exists := c.PointsMap[p.Y][p.X]; !exists {
			t.Errorf("Point should exist in Canvas at %v: expected %v, actual %v", p, true, exists)
		}
	}
}

func TestIdentIconCanvasIntCoordinates_generator(t *testing.T) {
	ii := generatedCanvas()
	ics := ii.Canvas.IntCoordinates()

	if len(ics) != ii.Canvas.FilledPoints {
		t.Fatalf("Unexpected Canvas IntCoordinates length: expected %v, actual %v", ii.Canvas.FilledPoints, len(ics))
	}

	for _, ic := range ics {
		if _, exists := ii.Canvas.PointsMap[ic[1]][ic[0]]; !exists {
			t.Errorf("Point should exist in Canvas at %v: expected %v, actual %v", ic, true, exists)
		}
	}
}

func TestIdentIconCanvasIntCoordinates_custom(t *testing.T) {
	c := customCanvas()
	ics := c.IntCoordinates()

	if len(ics) != c.FilledPoints {
		t.Fatalf("Unexpected Canvas IntCoordinates length: expected %v, actual %v", c.FilledPoints, len(ics))
	}

	for _, ic := range ics {
		if _, exists := c.PointsMap[ic[1]][ic[0]]; !exists {
			t.Errorf("Point should exist in Canvas at %v: expected %v, actual %v", ic, true, exists)
		}
	}
}

func generatedCanvas() *IdentIcon {
	size := 5
	ig, _ := New(
		"github.com",
		size,
		2,
	)

	ii, _ := ig.Draw("nullrocks")

	return ii
}

func customCanvas() Canvas {
	pm := make(map[int]map[int]int)
	pm[0] = make(map[int]int)
	pm[0][0] = 1
	pm[0][1] = 1
	pm[0][3] = 1
	pm[0][4] = 1
	pm[3] = make(map[int]int)
	pm[3][1] = 1
	pm[3][2] = 1
	pm[3][3] = 1

	visited := make(map[int]bool)
	visited[0] = true
	visited[3] = true

	c := Canvas{
		Size:           5,
		PointsMap:      pm,
		MinY:           0,
		MaxY:           3,
		VisitedYPoints: visited,
		FilledPoints:   7,
	}

	return c
}
