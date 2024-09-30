// Package bsi implements roaring bit sliced index on top of sroar package. The
// ain goal is to remove decoding step for reads because we erite/read a lot bsi
// takingadvantage of badger for reads yieds significant perfomance boost.
package bsi
