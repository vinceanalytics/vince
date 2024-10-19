package models

import (
	"fmt"
	"testing"
)

func TestTranslate(t *testing.T) {
	for n := range int(TranslatedFieldsSize) {
		fmt.Println(n, Mutex(n).Mutex(), Mutex(n))
	}
	for n := range int(BSIFieldsSize) {
		fmt.Println(n, BSI(n).BSI(), BSI(n))
	}
	t.Error()
}
