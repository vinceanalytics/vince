package len64

import (
	"runtime"
	"sync"
)

var numCpu int
var parOnce sync.Once

func parallel() int {
	parOnce.Do(func() {
		numCpu = runtime.NumCPU()
	})
	return numCpu
}
