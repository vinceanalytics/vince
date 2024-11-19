package ro2

import "math/bits"

const (
	exponent                 = 20
	shardWidth               = 1 << exponent
	rowExponent              = (exponent - 16)  // for instance, 20-16 = 4
	rowWidth                 = 1 << rowExponent // containers per row, for instance 1<<4 = 16
	keyMask                  = (rowWidth - 1)   // a mask for offset within the row
	shardVsContainerExponent = 4
)

func WriteMutex(ra *Bitmap, id uint64, value uint64) {
	ra.DirectAdd(value*shardWidth + (id % shardWidth))
}

const (
	// Row ids used for boolean fields.
	falseRowID = uint64(0)
	trueRowID  = uint64(1)

	falseRowOffset = 0 * shardWidth // fragment row 0
	trueRowOffset  = 1 * shardWidth // fragment row 1
)

func WriteBool(ra *Bitmap, id uint64, value bool) {
	fragmentColumn := id % shardWidth
	if value {
		ra.DirectAdd(trueRowOffset + fragmentColumn)
	} else {
		ra.DirectAdd(falseRowOffset + fragmentColumn)
	}
}

func WriteBSI(ra *Bitmap, id uint64, svalue int64) {
	fragmentColumn := id % shardWidth
	ra.DirectAdd(fragmentColumn)
	negative := svalue < 0
	var value uint64
	if negative {
		ra.DirectAdd(shardWidth + fragmentColumn) // set sign bit
		value = uint64(svalue * -1)
	} else {
		value = uint64(svalue)
	}
	lz := bits.LeadingZeros64(value)
	row := uint64(2)
	for mask := uint64(0x1); mask <= 1<<(64-lz) && mask != 0; mask = mask << 1 {
		if value&mask > 0 {
			ra.DirectAdd(row*shardWidth + fragmentColumn)
		}
		row++
	}
}
