package limit

import (
	"reflect"
	"testing"

	"golang.org/x/time/rate"
)

func TestYay(t *testing.T) {
	t.Error(reflect.ValueOf(rate.Limiter{}).Type().Size())
}
