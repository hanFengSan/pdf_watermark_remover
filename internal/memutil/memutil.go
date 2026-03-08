package memutil

import (
	"runtime"
	"runtime/debug"
)

func ForceRelease() {
	runtime.GC()
	debug.FreeOSMemory()
}
