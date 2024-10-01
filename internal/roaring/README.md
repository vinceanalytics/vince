# sroar: Serialized Roaring Bitmaps

sroar is a re-written version of Roaring Bitmaps in Go, with the aim to have
equality between in-memory representation and on-disk representation. An
sroar.Bitmap does not need to be marshalled or unmarshalled, as the underlying
represetation is a byte slice. Therefore, it can be written to disk, brought to
memory, or shipped over the network immediately. This is needed in [Dgraph][], where
we need to deal with lots of bitmaps.

sroar only implements array and bitmap containers. It does NOT implement run
containers, which is an optimization that RoaringBitmaps has. Despite that, it
outperforms RoaringBitmaps as shown in the Benchmarks section.

[Dgraph]: https://github.com/dgraph-io/dgraph
[Roaring]: https://github.com/RoaringBitmap/roaring

The code borrows concepts and code from [RoaringBitmaps][Roaring].

## Benchmarks

The benchmarks were run:
- Using real data set as described in [RoaringBitmaps][Roaring].
- Only on the 64-bit version of roaring bitmaps (roaring64).
- Only on `FastOr`, which is the more expensive operation than `And` or
    equivalent.
- On AMD Ryzen Threadripper 2950X 16-Core Processor.
- Using Go benchmarks serially.

Based on the benchmarks, sroar is:
- 6.5x faster (-85% p50) for benchmarks >1ms, uses
- 15x (-93.5% p50) less memory for allocations >1MB.
- 25x fewer allocations.

The benchmark command and the results are:

```
$ go test -bench BenchmarkRealDataFastOr --run=XXX --count=5 --benchmem

name CPU                                    old time/op    new time/op    delta
RealDataFastOr/census1881-32                 302ms ± 2%       2ms ± 3%   -99.29%  (p=0.008 n=5+5)
RealDataFastOr/wikileaks-noquotes-32        76.5ms ± 1%     0.9ms ± 1%   -98.83%  (p=0.008 n=5+5)
RealDataFastOr/wikileaks-noquotes_srt-32    34.8ms ± 5%     1.0ms ± 2%   -97.07%  (p=0.008 n=5+5)
RealDataFastOr/dimension_033-32             55.0ms ± 3%     2.7ms ± 0%   -95.16%  (p=0.016 n=5+4)
RealDataFastOr/census1881_srt-32            36.8ms ± 3%     2.9ms ± 1%   -92.13%  (p=0.008 n=5+5)
RealDataFastOr/dimension_003-32             50.4ms ± 1%    11.6ms ± 4%   -77.06%  (p=0.008 n=5+5)
RealDataFastOr/dimension_008-32             10.0ms ± 2%     3.7ms ± 2%   -62.69%  (p=0.008 n=5+5)
RealDataFastOr/weather_sept_85_srt-32       6.13ms ± 3%    2.72ms ± 2%   -55.66%  (p=0.008 n=5+5)
RealDataFastOr/census-income-32             1.70ms ± 3%    1.05ms ± 1%   -38.53%  (p=0.008 n=5+5)
RealDataFastOr/weather_sept_85-32           2.28ms ± 2%    4.07ms ± 2%   +78.52%  (p=0.008 n=5+5)

RealDataFastOr/uscensus2000-32               556µs ± 2%     791µs ± 1%   +42.17%  (p=0.008 n=5+5)
RealDataFastOr/census-income_srt-32          260µs ± 4%     986µs ± 2%  +279.09%  (p=0.008 n=5+5)

name MEM_BYTES                             old alloc/op   new alloc/op   delta
RealDataFastOr/census1881-32                 585MB ± 0%       1MB ± 0%   -99.75%  (p=0.008 n=5+5)
RealDataFastOr/wikileaks-noquotes-32        76.3MB ± 0%     0.6MB ± 0%   -99.24%  (p=0.008 n=5+5)
RealDataFastOr/wikileaks-noquotes_srt-32    22.8MB ± 0%     0.6MB ± 0%   -97.46%  (p=0.008 n=5+5)
RealDataFastOr/census1881_srt-32            15.3MB ± 0%     1.4MB ± 0%   -90.58%  (p=0.008 n=5+5)
RealDataFastOr/dimension_003-32             7.78MB ± 0%    1.44MB ± 0%   -81.49%  (p=0.008 n=5+5)
RealDataFastOr/dimension_033-32             1.10MB ± 0%    1.44MB ± 0%   +30.92%  (p=0.008 n=5+5)

RealDataFastOr/dimension_008-32              537kB ± 0%      97kB ± 0%   -81.94%  (p=0.008 n=5+5)
RealDataFastOr/census-income-32              187kB ± 0%      70kB ± 0%   -62.86%  (p=0.008 n=5+5)
RealDataFastOr/census-income_srt-32         99.1kB ± 0%    69.6kB ± 0%   -29.81%  (p=0.008 n=5+5)
RealDataFastOr/weather_sept_85_srt-32        375kB ± 0%     292kB ± 0%   -21.95%  (p=0.008 n=5+5)
RealDataFastOr/uscensus2000-32               169kB ± 0%     231kB ± 0%   +36.97%  (p=0.008 n=5+5)
RealDataFastOr/weather_sept_85-32            169kB ± 0%     292kB ± 0%   +72.93%  (p=0.008 n=5+5)

name MEM_ALLOCS                           old allocs/op  new allocs/op  delta
RealDataFastOr/census1881_srt-32             29.7k ± 0%      0.0k ± 0%   -99.91%  (p=0.008 n=5+5)
RealDataFastOr/wikileaks-noquotes_srt-32     6.06k ± 0%     0.02k ± 0%   -99.74%  (p=0.008 n=5+5)
RealDataFastOr/dimension_003-32              4.57k ± 0%     0.03k ± 2%   -99.42%  (p=0.008 n=5+5)
RealDataFastOr/dimension_033-32              4.33k ± 0%     0.03k ± 0%   -99.38%  (p=0.000 n=5+4)
RealDataFastOr/uscensus2000-32               1.75k ± 0%     0.06k ± 0%   -96.85%  (p=0.008 n=5+5)
RealDataFastOr/dimension_008-32                704 ± 0%        23 ± 3%   -96.79%  (p=0.008 n=5+5)
RealDataFastOr/census-income-32                271 ± 0%         9 ± 0%   -96.68%  (p=0.008 n=5+5)
RealDataFastOr/weather_sept_85_srt-32          248 ± 0%        14 ± 0%   -94.35%  (p=0.008 n=5+5)
RealDataFastOr/weather_sept_85-32             81.0 ± 0%      14.0 ± 0%   -82.72%  (p=0.008 n=5+5)
RealDataFastOr/census-income_srt-32           40.0 ± 0%       9.0 ± 0%   -77.50%  (p=0.008 n=5+5)
RealDataFastOr/census1881-32                 54.5k ± 0%      0.0k ± 0%      ~     (p=0.079 n=4+5)
RealDataFastOr/wikileaks-noquotes-32         39.2k ± 0%      0.0k ± 0%      ~     (p=0.079 n=4+5)
```
