package roaring

func (b *Bitmap) AndCardinality(a *Bitmap) (answer uint64) {
	return AndCardinality(b, a)
}

func AndCardinality(a, b *Bitmap) (answer uint64) {
	ai, an := 0, a.keys.numKeys()
	bi, bn := 0, b.keys.numKeys()

	for ai < an && bi < bn {
		ak := a.keys.key(ai)
		bk := a.keys.key(bi)
		if ak == bk {
			// Do the intersection.
			off := a.keys.val(ai)
			ac := a.getContainer(off)

			off = b.keys.val(bi)
			bc := b.getContainer(off)

			answer += uint64(containerAndCardinality(ac, bc))
			ai++
			bi++
		} else if ak < bk {
			ai++
		} else {
			bi++
		}
	}
	return
}
