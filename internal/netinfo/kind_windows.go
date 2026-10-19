//go:build windows

package netinfo

import (
	"net"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	ifTypeEthernet = 6  // IF_TYPE_ETHERNET_CSMACD
	ifTypeWiFi     = 71 // IF_TYPE_IEEE80211
)

// platformKind reads the adapter's type from GetAdaptersAddresses, which is exact (the names are whatever the
// person called the connection).
func platformKind(ifc net.Interface) string {
	size := uint32(15000)
	var buf []byte
	for i := 0; i < 3; i++ {
		buf = make([]byte, size)
		err := windows.GetAdaptersAddresses(windows.AF_INET, windows.GAA_FLAG_SKIP_ANYCAST|windows.GAA_FLAG_SKIP_MULTICAST|windows.GAA_FLAG_SKIP_DNS_SERVER, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0])), &size)
		if err == nil {
			break
		}
		if err != windows.ERROR_BUFFER_OVERFLOW {
			return ""
		}
	}
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0])); aa != nil; aa = aa.Next {
		if int(aa.IfIndex) != ifc.Index {
			continue
		}
		switch aa.IfType {
		case ifTypeWiFi:
			return KindWiFi
		case ifTypeEthernet:
			return KindEthernet
		}
		return KindOther
	}
	return ""
}
