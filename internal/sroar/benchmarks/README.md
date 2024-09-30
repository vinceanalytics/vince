The benchmarks were run on Manish's desktop with AMD Threadripper 2950x.

sroar.txt: Serialized Roaring Bitmaps
roar64.txt: RoaringBitmap/roaring using roaring64.
roar64-serial.txt: Same as above, but with extra step of Marshal after FastOr.
roar32.txt: RoaringBitmap/roaring using 32-bit roaring.
