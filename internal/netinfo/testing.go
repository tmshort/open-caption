package netinfo

import "time"

// SetTableForTest makes this computer look as if it had the given addresses (address -> interface name) for the next
// couple of seconds. It is for tests of the packages that use netinfo, which cannot depend on the real network.
func SetTableForTest(by map[string]string) {
	tableCache.mu.Lock()
	tableCache.by = by
	tableCache.at = time.Now()
	tableCache.mu.Unlock()
}
