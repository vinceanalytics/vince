package sroar

func Xor(a, b *Bitmap) *Bitmap {
	x := Or(a, b)
	x.AndNot(And(a, b))
	return x
}

func (b *Bitmap) Xor(a *Bitmap) {
	and := And(a, b)
	b.Or(a)
	b.AndNot(and)
}
