package collector

import (
	"sync"
	"time"
)

// Global variables
var (
	statsLock    sync.RWMutex
	lastCPUAlert time.Time
	lastMemAlert time.Time
	alertLock    sync.Mutex
)

func CollectMetrics() {
	// Start background collectors
	// go collectSystemMetrics()
	// go collectProcessMetrics()
}
