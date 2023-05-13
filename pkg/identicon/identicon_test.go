package identicon

import (
	"fmt"
	"strings"
	"testing"
)

func ExampleIdentIcon_GeneratorText() {
	ig, err := New("namespace", 7, 4)

	if err != nil {
		panic(err)
	}

	username := "text"
	ii, err := ig.Draw(username)

	if err != nil {
		panic(err)
	}

	fmt.Print(ii.GeneratorText())
	// Output:
	// text:namespace
}

func TestIdentIconGeneratorText(t *testing.T) {
	ig, _ := New(
		"",
		4,
		1,
	)

	username := "nullrocks"
	ii, _ := ig.Draw(username)

	gt := ii.GeneratorText()
	if gt != username {
		t.Errorf("GeneratorText should be the text: expected %v, actual %v", username, gt)
	}

}

func TestIdentIconGeneratorText_withNamespace(t *testing.T) {
	ig, _ := New(
		"github.com",
		4,
		1,
	)

	username := "nullrocks"
	ii, _ := ig.Draw(username)

	gt := ii.GeneratorText()
	expected := "nullrocks:github.com"
	if gt != expected {
		t.Errorf("GeneratorText should be the namespace and the text: expected %v, actual %v", expected, gt)
	}

}

func TestIdentIconGeneratorText_withRandom(t *testing.T) {
	ig, _ := New(
		"github.com",
		4,
		1,
		SetRandom(true),
	)
	username := "nullrocks"
	ii, _ := ig.Draw(username)
	gt := ii.GeneratorText()
	expected := "nullrocks:github.com:"
	if !strings.HasPrefix(gt, expected) || len(gt) <= len(expected) {
		t.Errorf("GeneratorText should be the namespace, the text and the random string: expected %v, actual %v", expected, gt)
	}

}
