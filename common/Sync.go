package common

import (
	"sync"
)

var globeLocker *sync.RWMutex

func GetGlobeLocker() *sync.RWMutex {
	if globeLocker == nil {
		globeLocker = new(sync.RWMutex)
	}
	return globeLocker
}
