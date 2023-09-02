package collector

import (
	"log"
	"sync"
	"sync/atomic"
)

// update counter of file lines
func UpdateCounter(counters *sync.Map, key string) {
	val, _ := counters.LoadOrStore(key, new(int64))
	ptr := val.(*int64)
	atomic.AddInt64(ptr, 1)
}

// print sync.Map
func PrintSyncMap(m sync.Map) {
	// print map,
	i := 0
	m.Range(func(key, value any) bool {
		log.Printf("\t[%d] key: %v, value: %d\n", i, key, *value.(*int64))
		i++
		return true
	})
}
