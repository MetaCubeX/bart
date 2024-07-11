// Copyright (c) 2024 Karl Gaissmaier
// SPDX-License-Identifier: MIT

package bart_test

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/gaissmai/bart"
)

var (
	a = netip.MustParseAddr
	p = netip.MustParsePrefix
)

var input = []struct {
	cidr    netip.Prefix
	nextHop netip.Addr
}{
	{p("fe80::/10"), a("::1%eth0")},
	{p("172.16.0.0/12"), a("8.8.8.8")},
	{p("10.0.0.0/24"), a("8.8.8.8")},
	{p("::1/128"), a("::1%lo")},
	{p("192.168.0.0/16"), a("9.9.9.9")},
	{p("10.0.0.0/8"), a("9.9.9.9")},
	{p("::/0"), a("2001:db8::1")},
	{p("10.0.1.0/24"), a("10.0.0.0")},
	{p("169.254.0.0/16"), a("10.0.0.0")},
	{p("2000::/3"), a("2001:db8::1")},
	{p("2001:db8::/32"), a("2001:db8::1")},
	{p("127.0.0.0/8"), a("127.0.0.1")},
	{p("127.0.0.1/32"), a("127.0.0.1")},
	{p("192.168.1.0/24"), a("127.0.0.1")},
}

func ExampleTable_Lookup() {
	rtbl := new(bart.Table[netip.Addr])
	for _, item := range input {
		rtbl.Insert(item.cidr, item.nextHop)
	}
	rtbl.Fprint(os.Stdout)

	fmt.Println()

	ip := a("42.0.0.0")
	value, ok := rtbl.Lookup(ip)
	fmt.Printf("Lookup: %-20v value: %11v, ok: %v\n", ip, value, ok)

	ip = a("10.0.1.17")
	value, ok = rtbl.Lookup(ip)
	fmt.Printf("Lookup: %-20v value: %11v, ok: %v\n", ip, value, ok)

	ip = a("2001:7c0:3100:1::111")
	value, ok = rtbl.Lookup(ip)
	fmt.Printf("Lookup: %-20v value: %11v, ok: %v\n", ip, value, ok)

	// Output:
	// ▼
	// ├─ 10.0.0.0/8 (9.9.9.9)
	// │  ├─ 10.0.0.0/24 (8.8.8.8)
	// │  └─ 10.0.1.0/24 (10.0.0.0)
	// ├─ 127.0.0.0/8 (127.0.0.1)
	// │  └─ 127.0.0.1/32 (127.0.0.1)
	// ├─ 169.254.0.0/16 (10.0.0.0)
	// ├─ 172.16.0.0/12 (8.8.8.8)
	// └─ 192.168.0.0/16 (9.9.9.9)
	//    └─ 192.168.1.0/24 (127.0.0.1)
	// ▼
	// └─ ::/0 (2001:db8::1)
	//    ├─ ::1/128 (::1%lo)
	//    ├─ 2000::/3 (2001:db8::1)
	//    │  └─ 2001:db8::/32 (2001:db8::1)
	//    └─ fe80::/10 (::1%eth0)
	//
	// Lookup: 42.0.0.0             value:  invalid IP, ok: false
	// Lookup: 10.0.1.17            value:    10.0.0.0, ok: true
	// Lookup: 2001:7c0:3100:1::111 value: 2001:db8::1, ok: true
}

func ExampleTable_All4() {
	rtbl := new(bart.Table[netip.Addr])
	for _, item := range input {
		rtbl.Insert(item.cidr, item.nextHop)
	}
	rtbl.All4Sorted(func(pfx netip.Prefix, val netip.Addr) bool {
		fmt.Printf("%v\t%v\n", pfx, val)
		return true
	})

	// Output:
	// 10.0.0.0/8	9.9.9.9
	// 10.0.0.0/24	8.8.8.8
	// 10.0.1.0/24	10.0.0.0
	// 127.0.0.0/8	127.0.0.1
	// 127.0.0.1/32	127.0.0.1
	// 169.254.0.0/16	10.0.0.0
	// 172.16.0.0/12	8.8.8.8
	// 192.168.0.0/16	9.9.9.9
	// 192.168.1.0/24	127.0.0.1
}

func ExampleTable_EachSubnet() {
	rtbl := new(bart.Table[netip.Addr])
	for _, item := range input {
		rtbl.Insert(item.cidr, item.nextHop)
	}

	cidr := netip.MustParsePrefix("0.0.0.0/1")

	rtbl.EachSubnet(cidr, func(pfx netip.Prefix, _ netip.Addr) bool {
		fmt.Printf("%v\n", pfx)
		return true
	})

	// Output:
	// 10.0.0.0/8
	// 10.0.0.0/24
	// 10.0.1.0/24
	// 127.0.0.0/8
	// 127.0.0.1/32

}

func ExampleTable_EachLookupPrefix() {
	rtbl := new(bart.Table[netip.Addr])
	for _, item := range input {
		rtbl.Insert(item.cidr, item.nextHop)
	}

	cidr := netip.MustParsePrefix("2001:db8::/32")

	counter := 0
	rtbl.EachLookupPrefix(cidr, func(pfx netip.Prefix, _ netip.Addr) bool {
		fmt.Printf("%v\n", pfx)
		counter++

		// premature end of iteration?
		return counter < 2
	})

	// Output:
	// 2001:db8::/32
	// 2000::/3
}
