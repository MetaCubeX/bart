// Copyright (c) 2024 Karl Gaissmaier
// SPDX-License-Identifier: MIT

package bart

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"math/rand"
	"net/netip"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var prng = rand.New(&lockedSource{source: rand.NewSource(42)})

type lockedSource struct {
	source rand.Source
	mu     sync.Mutex
}

func (s *lockedSource) Seed(seed int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.source.Seed(seed)
}

func (s *lockedSource) Int63() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.source.Int63()
}
func (s *lockedSource) Uint64() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.source.(rand.Source64).Uint64()
}

// full internet prefix list, gzipped
const prefixFile = "testdata/prefixes.txt.gz"

var (
	routes  []route
	routes4 []route
	routes6 []route

	randRoute4 route
	randRoute6 route
)

type route struct {
	CIDR  netip.Prefix
	Value any
}

func init() {
	fillRouteTables()

	randRoute4 = routes4[prng.Intn(len(routes4))]
	randRoute6 = routes6[prng.Intn(len(routes6))]
}

var (
	intSink  int
	okSink   bool
	boolSink bool
)

func BenchmarkFullMatch4(b *testing.B) {
	var rt Table[int]
	var lt Lite

	for i, route := range routes {
		rt.Insert(route.CIDR, i)
		lt.Insert(route.CIDR)
	}

	var ip netip.Addr
	var ipAsPfx netip.Prefix

	// find a random match
	for {
		ip = randomIP4()
		_, ok := rt.Lookup(ip)
		if ok {
			ipAsPfx, _ = ip.Prefix(ip.BitLen())
			break
		}
	}

	b.Run("Lite.Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = lt.Contains(ip)
		}
	})

	b.Run("Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = rt.Contains(ip)
		}
	})

	b.Run("Lookup", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.Lookup(ip)
		}
	})

	b.Run("LookupPrefix", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.LookupPrefix(ipAsPfx)
		}
	})

	b.Run("LookupPfxLPM", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			_, intSink, okSink = rt.LookupPrefixLPM(ipAsPfx)
		}
	})
}

func BenchmarkFullMatch6(b *testing.B) {
	var rt Table[int]
	var lt Lite

	for i, route := range routes {
		rt.Insert(route.CIDR, i)
		lt.Insert(route.CIDR)
	}

	var ip netip.Addr
	var ipAsPfx netip.Prefix

	// find a random match
	for {
		ip = randomIP6()
		_, ok := rt.Lookup(ip)
		if ok {
			ipAsPfx, _ = ip.Prefix(ip.BitLen())
			break
		}
	}

	b.Run("Lite.Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = lt.Contains(ip)
		}
	})

	b.Run("Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = rt.Contains(ip)
		}
	})

	b.Run("Lookup", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.Lookup(ip)
		}
	})

	b.Run("LookupPrefix", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.LookupPrefix(ipAsPfx)
		}
	})

	b.Run("LookupPfxLPM", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			_, intSink, okSink = rt.LookupPrefixLPM(ipAsPfx)
		}
	})
}

func BenchmarkFullMiss4(b *testing.B) {
	var rt Table[int]
	var lt Lite

	for i, route := range routes {
		rt.Insert(route.CIDR, i)
		lt.Insert(route.CIDR)
	}

	var ip netip.Addr
	var ipAsPfx netip.Prefix

	// find a random miss
	for {
		ip = randomIP4()
		_, ok := rt.Lookup(ip)
		if !ok {
			ipAsPfx, _ = ip.Prefix(ip.BitLen())
			break
		}
	}

	b.Run("Lite.Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = lt.Contains(ip)
		}
	})

	b.Run("Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = rt.Contains(ip)
		}
	})

	b.Run("Lookup", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.Lookup(ip)
		}
	})

	b.Run("LookupPrefix", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.LookupPrefix(ipAsPfx)
		}
	})

	b.Run("LookupPfxLPM", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			_, intSink, okSink = rt.LookupPrefixLPM(ipAsPfx)
		}
	})
}

func BenchmarkFullMiss6(b *testing.B) {
	var rt Table[int]
	var lt Lite

	for i, route := range routes {
		rt.Insert(route.CIDR, i)
		lt.Insert(route.CIDR)
	}

	var ip netip.Addr
	var ipAsPfx netip.Prefix

	// find a random miss
	for {
		ip = randomIP6()
		_, ok := rt.Lookup(ip)
		if !ok {
			ipAsPfx, _ = ip.Prefix(ip.BitLen())
			break
		}
	}

	b.Run("Lite.Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = lt.Contains(ip)
		}
	})

	b.Run("Contains", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			okSink = rt.Contains(ip)
		}
	})

	b.Run("Lookup", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.Lookup(ip)
		}
	})

	b.Run("LookupPrefix", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			intSink, okSink = rt.LookupPrefix(ipAsPfx)
		}
	})

	b.Run("LookupPfxLPM", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			_, intSink, okSink = rt.LookupPrefixLPM(ipAsPfx)
		}
	})
}

func BenchmarkFullTableOverlaps4(b *testing.B) {
	var rt Table[int]

	for i, route := range routes4 {
		rt.Insert(route.CIDR, i)
	}

	for i := 1; i <= 1<<20; i *= 2 {
		rt2 := new(Table[int])
		for j, pfx := range randomRealWorldPrefixes4(i) {
			rt2.Insert(pfx, j)
		}

		b.Run(fmt.Sprintf("With_%4d", i), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				boolSink = rt.Overlaps(rt2)
			}
		})
	}
}

func BenchmarkFullTableOverlaps6(b *testing.B) {
	var rt Table[int]

	for i, route := range routes6 {
		rt.Insert(route.CIDR, i)
	}

	for i := 1; i <= 1<<20; i *= 2 {
		rt2 := new(Table[int])
		for j, pfx := range randomRealWorldPrefixes6(i) {
			rt2.Insert(pfx, j)
		}

		b.Run(fmt.Sprintf("With_%4d", i), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				boolSink = rt.Overlaps(rt2)
			}
		})
	}
}

func BenchmarkFullTableOverlapsPrefix(b *testing.B) {
	var rt Table[int]

	for i, route := range routes {
		rt.Insert(route.CIDR, i)
	}

	pfx := randomRealWorldPrefixes(1)[0]

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		boolSink = rt.OverlapsPrefix(pfx)
	}
}

func BenchmarkFullTableClone(b *testing.B) {
	var rt4 Table[int]

	for i, route := range routes4 {
		rt4.Insert(route.CIDR, i)
	}

	b.ResetTimer()
	b.Run("CloneIP4", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			_ = rt4.Clone()
		}
	})

	var rt6 Table[int]

	for i, route := range routes6 {
		rt6.Insert(route.CIDR, i)
	}

	b.ResetTimer()
	b.Run("CloneIP6", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			_ = rt6.Clone()
		}
	})

	var rt Table[int]

	for i, route := range routes {
		rt.Insert(route.CIDR, i)
	}

	b.ResetTimer()
	b.Run("Clone", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			_ = rt.Clone()
		}
	})
}

func BenchmarkFullTableMemory4(b *testing.B) {
	var startMem, endMem runtime.MemStats

	rt := new(Table[struct{}])
	runtime.GC()
	runtime.ReadMemStats(&startMem)

	b.Run(fmt.Sprintf("Table[]: %d", len(routes4)), func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for _, route := range routes4 {
				rt.Insert(route.CIDR, struct{}{})
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&endMem)

		stats := rt.root4.nodeStatsRec()
		b.ReportMetric(float64(endMem.HeapAlloc-startMem.HeapAlloc)/1024, "KByte")
		b.ReportMetric(float64(stats.pfxs), "pfx")
		b.ReportMetric(float64(stats.nodes), "node")
		b.ReportMetric(float64(stats.leaves), "leave")
		b.ReportMetric(float64(stats.fringes), "fringe")
		b.ReportMetric(0, "ns/op")
	})

	lt := new(Lite)
	runtime.GC()
	runtime.ReadMemStats(&startMem)

	b.Run(fmt.Sprintf("Lite: %d", len(routes4)), func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for _, route := range routes4 {
				lt.Insert(route.CIDR)
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&endMem)

		stats := rt.root4.nodeStatsRec()
		b.ReportMetric(float64(endMem.HeapAlloc-startMem.HeapAlloc)/1024, "KByte")
		b.ReportMetric(float64(stats.pfxs), "pfx")
		b.ReportMetric(float64(stats.nodes), "node")
		b.ReportMetric(float64(stats.leaves), "leave")
		b.ReportMetric(float64(stats.fringes), "fringe")
		b.ReportMetric(0, "ns/op")
	})
}

func BenchmarkFullTableMemory6(b *testing.B) {
	var startMem, endMem runtime.MemStats

	rt := new(Table[struct{}])
	runtime.GC()
	runtime.ReadMemStats(&startMem)

	b.Run(fmt.Sprintf("Table[]: %d", len(routes6)), func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for _, route := range routes6 {
				rt.Insert(route.CIDR, struct{}{})
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&endMem)

		stats := rt.root6.nodeStatsRec()
		b.ReportMetric(float64(endMem.HeapAlloc-startMem.HeapAlloc)/1024, "KByte")
		b.ReportMetric(float64(stats.pfxs), "pfx")
		b.ReportMetric(float64(stats.nodes), "node")
		b.ReportMetric(float64(stats.leaves), "leave")
		b.ReportMetric(float64(stats.fringes), "fringe")
		b.ReportMetric(0, "ns/op")
	})

	lt := new(Lite)
	runtime.GC()
	runtime.ReadMemStats(&startMem)

	b.Run(fmt.Sprintf("Lite: %d", len(routes6)), func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for _, route := range routes6 {
				lt.Insert(route.CIDR)
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&endMem)

		stats := rt.root6.nodeStatsRec()
		b.ReportMetric(float64(endMem.HeapAlloc-startMem.HeapAlloc)/1024, "KByte")
		b.ReportMetric(float64(stats.pfxs), "pfx")
		b.ReportMetric(float64(stats.nodes), "node")
		b.ReportMetric(float64(stats.leaves), "leave")
		b.ReportMetric(float64(stats.fringes), "fringe")
		b.ReportMetric(0, "ns/op")
	})
}

func BenchmarkFullTableMemory(b *testing.B) {
	var startMem, endMem runtime.MemStats

	rt := new(Table[struct{}])
	runtime.GC()
	runtime.ReadMemStats(&startMem)

	b.Run(fmt.Sprintf("Table[]: %d", len(routes)), func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for _, route := range routes {
				rt.Insert(route.CIDR, struct{}{})
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&endMem)

		s4 := rt.root4.nodeStatsRec()
		s6 := rt.root6.nodeStatsRec()
		stats := stats{
			s4.pfxs + s6.pfxs,
			s4.childs + s6.childs,
			s4.nodes + s6.nodes,
			s4.leaves + s6.leaves,
			s4.fringes + s6.fringes,
		}

		b.ReportMetric(float64(endMem.HeapAlloc-startMem.HeapAlloc)/1024, "KByte")
		b.ReportMetric(float64(stats.pfxs), "pfx")
		b.ReportMetric(float64(stats.nodes), "node")
		b.ReportMetric(float64(stats.leaves), "leave")
		b.ReportMetric(float64(stats.fringes), "fringe")
		b.ReportMetric(0, "ns/op")
	})

	lt := new(Lite)
	runtime.GC()
	runtime.ReadMemStats(&startMem)

	b.Run(fmt.Sprintf("Lite: %d", len(routes)), func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for _, route := range routes {
				lt.Insert(route.CIDR)
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&endMem)

		s4 := rt.root4.nodeStatsRec()
		s6 := rt.root6.nodeStatsRec()
		stats := stats{
			s4.pfxs + s6.pfxs,
			s4.childs + s6.childs,
			s4.nodes + s6.nodes,
			s4.leaves + s6.leaves,
			s4.fringes + s6.fringes,
		}

		b.ReportMetric(float64(endMem.HeapAlloc-startMem.HeapAlloc)/1024, "KByte")
		b.ReportMetric(float64(stats.pfxs), "pfx")
		b.ReportMetric(float64(stats.nodes), "node")
		b.ReportMetric(float64(stats.leaves), "leave")
		b.ReportMetric(float64(stats.fringes), "fringe")
		b.ReportMetric(0, "ns/op")
	})
}

func fillRouteTables() {
	file, err := os.Open(prefixFile)
	if err != nil {
		log.Fatal(err)
	}

	rgz, err := gzip.NewReader(file)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(rgz)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		cidr := netip.MustParsePrefix(line)
		cidr = cidr.Masked()

		routes = append(routes, route{cidr, cidr})

		if cidr.Addr().Is4() {
			routes4 = append(routes4, route{cidr, cidr})
		} else {
			routes6 = append(routes6, route{cidr, cidr})
		}
	}

	if err = scanner.Err(); err != nil {
		log.Printf("reading from %v, %v", rgz, err)
	}
}

// #########################################################

func randomRealWorldPrefixes4(n int) []netip.Prefix {
	set := map[netip.Prefix]netip.Prefix{}
	pfxs := make([]netip.Prefix, 0, n)

	for {
		pfx := randomPrefix4()

		// skip too small or too big masks
		if pfx.Bits() < 8 || pfx.Bits() > 28 {
			continue
		}

		// skip multicast ...
		if pfx.Overlaps(mpp("240.0.0.0/8")) {
			continue
		}

		if _, ok := set[pfx]; !ok {
			set[pfx] = pfx
			pfxs = append(pfxs, pfx)
		}

		if len(set) >= n {
			break
		}
	}
	return pfxs
}

func randomRealWorldPrefixes6(n int) []netip.Prefix {
	set := map[netip.Prefix]netip.Prefix{}
	pfxs := make([]netip.Prefix, 0, n)

	for {
		pfx := randomPrefix6()

		// skip too small or too big masks
		if pfx.Bits() < 16 || pfx.Bits() > 56 {
			continue
		}

		// skip non global routes seen in the real world
		if !pfx.Overlaps(mpp("2000::/3")) {
			continue
		}
		if pfx.Addr().Compare(mpp("2c0f::/16").Addr()) == 1 {
			continue
		}

		if _, ok := set[pfx]; !ok {
			set[pfx] = pfx
			pfxs = append(pfxs, pfx)
		}
		if len(set) >= n {
			break
		}
	}
	return pfxs
}

func randomRealWorldPrefixes(n int) []netip.Prefix {
	pfxs := make([]netip.Prefix, 0, n)
	pfxs = append(pfxs, randomRealWorldPrefixes4(n/2)...)
	pfxs = append(pfxs, randomRealWorldPrefixes6(n-len(pfxs))...)

	prng.Shuffle(n, func(i, j int) {
		pfxs[i], pfxs[j] = pfxs[j], pfxs[i]
	})

	return pfxs
}
