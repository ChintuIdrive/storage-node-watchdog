package collector

import (
	"sync"
)

// Global variables
var (
	statsLock sync.RWMutex
)

func CollectMetrics() {
	// Start background collectors
	// go collectSystemMetrics()
	// go collectProcessMetrics()
}
