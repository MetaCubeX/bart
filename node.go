// Copyright (c) 2024 Karl Gaissmaier
// SPDX-License-Identifier: MIT

package bart

import (
	"net/netip"
	"sort"

	"github.com/metacubex/bart/internal/art"
	"github.com/metacubex/bart/internal/lpm"
	"github.com/metacubex/bart/internal/sparse"
)

const (
	strideLen    = 8   // byte, a multibit trie with stride len 8
	maxTreeDepth = 16  // max 16 bytes for IPv6
	maxItems     = 256 // max 256 prefixes or children in node
)

// stridePath, max 16 octets deep
type stridePath [maxTreeDepth]uint8

// node is a level node in the multibit-trie.
// A node has prefixes and children, forming the multibit trie.
//
// The prefixes, mapped by the baseIndex() function from the ART algorithm,
// form a complete binary tree.
// See the artlookup.pdf paper in the doc folder to understand the mapping function
// and the binary tree of prefixes.
//
// In contrast to the ART algorithm, sparse arrays (popcount-compressed slices)
// are used instead of fixed-size arrays.
//
// The array slots are also not pre-allocated (alloted) as described
// in the ART algorithm, fast bitset operations are used to find the
// longest-prefix-match.
//
// The child array recursively spans the trie with a branching factor of 256
// and also records path-compressed leaves in the free node slots.
type node[V any] struct {
	// prefixes contains the routes, indexed as a complete binary tree with payload V
	// with the help of the baseIndex mapping function from the ART algorithm.
	prefixes sparse.Array256[V]

	// children, recursively spans the trie with a branching factor of 256.
	children sparse.Array256[any] // [any] is a *node or a *leaf
}

// isEmpty returns true if node has neither prefixes nor children
func (n *node[V]) isEmpty() bool {
	return n.prefixes.Len() == 0 && n.children.Len() == 0
}

// leaf is a prefix with value, used as a path compressed child, with a fringe flag.
type leaf[V any] struct {
	prefix netip.Prefix
	value  V
	fringe bool
}

// isFringe, prefixes with /8, /16, ... /128 bits
//
// A leaf inserted at the last possible level (depth == lastIdx-1)
// before becoming a prefix in the next level down (depth == lastIdx).
//
// A leaf being a fringe is the default-route for all nodes below this slot.
//
//	e.g. prefix is addr/8, or addr/16, or ... addr/128
//	depth <  lastIdx-1 : a leaf, path-compressed
//	depth == lastIdx-1 : a fringe leaf, path-compressed
//	depth == lastIdx   : a prefix with octet/0 => idx == 1, a default route
func isFringe(depth, bits int) bool {
	lastIdx, lastBits := lastOctetIdxAndBits(bits)
	return depth == lastIdx-1 && lastBits == 0
}

// cloneOrCopy, helper function,
// deep copy if v implements the Cloner interface.
func cloneOrCopy[V any](val V) V {
	if cloner, ok := any(val).(Cloner[V]); ok {
		return cloner.Clone()
	}
	// just a shallow copy
	return val
}

// cloneLeaf returns a clone of the leaf
// if the value implements the Cloner interface.
func (l *leaf[V]) cloneLeaf() *leaf[V] {
	return &leaf[V]{prefix: l.prefix, value: cloneOrCopy(l.value), fringe: l.fringe}
}

// insertAtDepth insert a prefix/val into a node tree at depth.
// n must not be nil, prefix must be valid and already in canonical form.
//
// If a path compression has to be resolved because a new value is added
// that collides with a leaf, the compressed leaf is then reinserted
// one depth down in the node trie.
func (n *node[V]) insertAtDepth(pfx netip.Prefix, val V, depth int) (exists bool) {
	ip := pfx.Addr()
	bits := pfx.Bits()
	octets := ip.AsSlice()
	lastIdx, lastBits := lastOctetIdxAndBits(bits)

	// find the proper trie node to insert prefix
	// start with prefix octet at depth
	for ; depth < len(octets); depth++ {
		octet := octets[depth]
		addr := uint(octet)

		// last significant octet: insert/override prefix/val into node
		if depth == lastIdx {
			return n.prefixes.InsertAt(art.PfxToIdx(octet, lastBits), val)
		}

		// reached end of trie path ...
		if !n.children.Test(addr) {
			// insert prefix path compressed as leaf or fringe
			return n.children.InsertAt(addr, &leaf[V]{prefix: pfx, value: val, fringe: isFringe(depth, bits)})
		}

		// ... or decend down the trie
		kid := n.children.MustGet(addr)

		// kid is node or leaf at addr
		switch kid := kid.(type) {
		case *node[V]:
			n = kid
			continue // descend down to next trie level

		case *leaf[V]:
			// reached a path compressed prefix
			// override value in slot if prefixes are equal
			if kid.prefix == pfx {
				kid.value = val
				// exists
				return true
			}

			// create new node
			// push the leaf down
			// insert new child at current leaf position (addr)
			// descend down, replace n with new child
			newNode := new(node[V])
			newNode.insertAtDepth(kid.prefix, kid.value, depth+1)

			n.children.InsertAt(addr, newNode)
			n = newNode

		default:
			panic("logic error, wrong node type")
		}
	}

	panic("unreachable")
}

// insertAtDepthPersist is the immutable version of insertAtDepth.
// All visited nodes are cloned during insertion.
func (n *node[V]) insertAtDepthPersist(pfx netip.Prefix, val V, depth int) (exists bool) {
	ip := pfx.Addr()
	bits := pfx.Bits()
	octets := ip.AsSlice()
	lastIdx, lastBits := lastOctetIdxAndBits(bits)

	// find the proper trie node to insert prefix
	// start with prefix octet at depth
	for ; depth < len(octets); depth++ {

		octet := octets[depth]
		addr := uint(octet)

		// last significant octet: insert/override prefix/val into node
		if depth == lastIdx {
			return n.prefixes.InsertAt(art.PfxToIdx(octet, lastBits), val)
		}

		if !n.children.Test(addr) {
			// insert new prefix path compressed
			return n.children.InsertAt(addr, &leaf[V]{prefix: pfx, value: val, fringe: isFringe(depth, bits)})
		}
		kid := n.children.MustGet(addr)

		// kid is node or leaf at addr
		switch kid := kid.(type) {
		case *node[V]:
			// proceed to next level
			kid = kid.cloneFlat()
			n.children.InsertAt(addr, kid)
			n = kid
			continue // descend down to next trie level

		case *leaf[V]:
			kid = kid.cloneLeaf()

			// override value in slot if prefixes are equal
			if kid.prefix == pfx {
				// exists
				kid.value = val
				return true
			}

			// create new node
			// push the leaf down
			// insert new child at current leaf position (addr)
			// descend down, replace n with new child
			newNode := new(node[V])
			newNode.insertAtDepth(kid.prefix, kid.value, depth+1)

			n.children.InsertAt(addr, newNode)
			n = newNode

		default:
			panic("logic error, wrong node type")
		}
	}

	panic("unreachable")
}

// purgeAndCompress, purge empty nodes or compress nodes with single prefix or leaf.
func (n *node[V]) purgeAndCompress(parentStack []*node[V], childPath []uint8, is4 bool) {
	// unwind the stack
	for depth := len(parentStack) - 1; depth >= 0; depth-- {
		parent := parentStack[depth]
		addr := uint(childPath[depth])

		pfxCount := n.prefixes.Len()
		childCount := n.children.Len()

		switch {
		case n.isEmpty():
			// just delete this empty node
			parent.children.DeleteAt(addr)

		case pfxCount == 0 && childCount == 1:
			// if child is a leaf (not a node), shift it up one level
			if kid, ok := n.children.Items[0].(*leaf[V]); ok {
				// delete this node
				parent.children.DeleteAt(addr)

				// ... insert prefix/value at parents depth
				parent.insertAtDepth(kid.prefix, kid.value, depth)
			}

		case pfxCount == 1 && childCount == 0:
			// get prefix back from idx
			idx, _ := n.prefixes.FirstSet() // single idx must be first bit set
			val := n.prefixes.Items[0]      // single value must be at Items[0]

			// ... and octet path
			path := stridePath{}
			copy(path[:], childPath)
			pfx := cidrFromPath(path, depth+1, is4, idx)

			// delete this node
			parent.children.DeleteAt(addr)

			// ... insert prefix/value at parents depth
			parent.insertAtDepth(pfx, val, depth)
		}

		// climb up the stack
		n = parent
	}
}

// lpmGet does a route lookup for idx in the 8-bit (stride) routing table
// at this depth and returns (baseIdx, value, true) if a matching
// longest prefix exists, or ok=false otherwise.
//
// The prefixes in the stride form a complete binary tree (CBT) using the baseIndex function.
// In contrast to the ART algorithm, I do not use an allotment approach but map
// the backtracking in the CBT by a bitset operation with a precalculated backtracking path
// for the respective idx.
func (n *node[V]) lpmGet(idx uint) (baseIdx uint, val V, ok bool) {
	// top is the idx of the longest-prefix-match
	if top, ok := n.prefixes.IntersectionTop(lpm.BackTrackingBitset(idx)); ok {
		return top, n.prefixes.MustGet(top), true
	}

	// not found (on this level)
	return 0, val, false
}

// lpmTest, true if idx has a (any) longest-prefix-match in node.
// this is a contains test, faster as lookup and without value returns.
func (n *node[V]) lpmTest(idx uint) bool {
	return n.prefixes.IntersectsAny(lpm.BackTrackingBitset(idx))
}

// cloneRec, clones the node recursive.
func (n *node[V]) cloneRec() *node[V] {
	if n == nil {
		return nil
	}

	c := new(node[V])
	if n.isEmpty() {
		return c
	}

	// shallow
	c.prefixes = *(n.prefixes.Copy())

	_, isCloner := any(*new(V)).(Cloner[V])

	// deep copy if V implements Cloner[V]
	if isCloner {
		for i, val := range c.prefixes.Items {
			c.prefixes.Items[i] = cloneOrCopy(val)
		}
	}

	// shallow
	c.children = *(n.children.Copy())

	// deep copy of nodes and leaves
	for i, kidAny := range c.children.Items {
		switch kid := kidAny.(type) {
		case *node[V]:
			// clone the child node rec-descent
			c.children.Items[i] = kid.cloneRec()
		case *leaf[V]:
			// deep copy if V implements Cloner[V]
			c.children.Items[i] = kid.cloneLeaf()

		default:
			panic("logic error, wrong node type")
		}
	}

	return c
}

// cloneFlat, copies the node and clone the values in prefixes and path compressed leaves
// if V implements Cloner. Used in the various ...Persist functions.
func (n *node[V]) cloneFlat() *node[V] {
	if n == nil {
		return nil
	}

	c := new(node[V])
	if n.isEmpty() {
		return c
	}

	// shallow copy
	c.prefixes = *(n.prefixes.Copy())
	c.children = *(n.children.Copy())

	if _, ok := any(*new(V)).(Cloner[V]); !ok {
		// if V doesn't implement Cloner[V], return early
		return c
	}

	// deep copy of values in prefixes
	for i, val := range c.prefixes.Items {
		c.prefixes.Items[i] = cloneOrCopy(val)
	}

	// deep copy of values in path compressed leaves
	for i, kidAny := range c.children.Items {
		if kidLeaf, ok := kidAny.(*leaf[V]); ok {
			c.children.Items[i] = kidLeaf.cloneLeaf()
		}
	}

	return c
}

// allRec runs recursive the trie, starting at this node and
// the yield function is called for each route entry with prefix and value.
// If the yield function returns false the recursion ends prematurely and the
// false value is propagated.
//
// The iteration order is not defined, just the simplest and fastest recursive implementation.
func (n *node[V]) allRec(path stridePath, depth int, is4 bool, yield func(netip.Prefix, V) bool) bool {
	for _, idx := range n.prefixes.AsSlice(make([]uint, 0, maxItems)) {
		cidr := cidrFromPath(path, depth, is4, idx)

		// callback for this prefix and val
		if !yield(cidr, n.prefixes.MustGet(idx)) {
			// early exit
			return false
		}
	}

	// for all children (nodes and leaves) in this node do ...
	for i, addr := range n.children.AsSlice(make([]uint, 0, maxItems)) {
		switch kid := n.children.Items[i].(type) {
		case *node[V]:
			// rec-descent with this node
			path[depth] = byte(addr)
			if !kid.allRec(path, depth+1, is4, yield) {
				// early exit
				return false
			}
		case *leaf[V]:
			// callback for this leaf
			if !yield(kid.prefix, kid.value) {
				// early exit
				return false
			}

		default:
			panic("logic error, wrong node type")
		}
	}

	return true
}

// allRecSorted runs recursive the trie, starting at node and
// the yield function is called for each route entry with prefix and value.
// The iteration is in prefix sort order.
//
// If the yield function returns false the recursion ends prematurely and the
// false value is propagated.
func (n *node[V]) allRecSorted(path stridePath, depth int, is4 bool, yield func(netip.Prefix, V) bool) bool {
	// get slice of all child octets, sorted by addr
	allChildAddrs := n.children.AsSlice(make([]uint, 0, maxItems))

	// get slice of all indexes, sorted by idx
	allIndices := n.prefixes.AsSlice(make([]uint, 0, maxItems))

	// sort indices in CIDR sort order
	sort.Slice(allIndices, func(i, j int) bool {
		return lessIndexRank(allIndices[i], allIndices[j])
	})

	childCursor := 0

	// yield indices and childs in CIDR sort order
	for _, pfxIdx := range allIndices {
		pfxOctet, _ := art.IdxToPfx(pfxIdx)

		// yield all childs before idx
		for j := childCursor; j < len(allChildAddrs); j++ {
			childAddr := allChildAddrs[j]

			if childAddr >= uint(pfxOctet) {
				break
			}

			// yield the node (rec-descent) or leaf
			switch kid := n.children.Items[j].(type) {
			case *node[V]:
				path[depth] = byte(childAddr)
				if !kid.allRecSorted(path, depth+1, is4, yield) {
					return false
				}
			case *leaf[V]:
				if !yield(kid.prefix, kid.value) {
					return false
				}

			default:
				panic("logic error, wrong node type")
			}

			childCursor++
		}

		// yield the prefix for this idx
		cidr := cidrFromPath(path, depth, is4, pfxIdx)
		// n.prefixes.Items[i] not possible after sorting allIndices
		if !yield(cidr, n.prefixes.MustGet(pfxIdx)) {
			return false
		}
	}

	// yield the rest of leaves and nodes (rec-descent)
	for j := childCursor; j < len(allChildAddrs); j++ {
		addr := allChildAddrs[j]
		switch kid := n.children.Items[j].(type) {
		case *node[V]:
			path[depth] = byte(addr)
			if !kid.allRecSorted(path, depth+1, is4, yield) {
				return false
			}
		case *leaf[V]:
			if !yield(kid.prefix, kid.value) {
				return false
			}

		default:
			panic("logic error, wrong node type")
		}
	}

	return true
}

// unionRec combines two nodes, changing the receiver node.
// If there are duplicate entries, the value is taken from the other node.
// Count duplicate entries to adjust the t.size struct members.
func (n *node[V]) unionRec(o *node[V], depth int) (duplicates int) {
	// for all prefixes in other node do ...
	for i, oIdx := range o.prefixes.AsSlice(make([]uint, 0, maxItems)) {
		// insert/overwrite prefix/value from oNode to nNode
		exists := n.prefixes.InsertAt(oIdx, o.prefixes.Items[i])

		// this prefix is duplicate in n and o
		if exists {
			duplicates++
		}
	}

LOOP:
	// for all child addrs in other node do ...
	for i, addr := range o.children.AsSlice(make([]uint, 0, maxItems)) {
		//  6 possible combinations for this child and other child child
		//
		//  THIS, OTHER:
		//  ----------
		//  NULL, node  <-- easy,    insert at cloned node
		//  NULL, leaf  <-- easy,    insert at cloned leaf
		//  node, node  <-- easy,    union rec-descent
		//  node, leaf  <-- easy,    insert other cloned leaf at depth+1
		//  leaf, node  <-- complex, push this leaf down, union rec-descent
		//  leaf, leaf  <-- complex, push this leaf down, insert other cloned leaf at depth+1
		//
		// try to get child at same addr from n
		thisChild, thisExists := n.children.Get(addr)
		if !thisExists {
			switch otherChild := o.children.Items[i].(type) {
			case *node[V]: // NULL, node
				if !thisExists {
					n.children.InsertAt(addr, otherChild.cloneRec())
					continue LOOP
				}

			case *leaf[V]: // NULL, leaf
				if !thisExists {
					n.children.InsertAt(addr, otherChild.cloneLeaf())
					continue LOOP
				}

			default:
				panic("logic error, wrong node type")
			}
		}

		switch otherChild := o.children.Items[i].(type) {
		case *node[V]:
			switch this := thisChild.(type) {
			case *node[V]: // node, node
				// both childs have node in octet, call union rec-descent on child nodes
				duplicates += this.unionRec(otherChild, depth+1)
				continue LOOP

			case *leaf[V]: // leaf, node
				// create new node
				nc := new(node[V])

				// push this leaf down
				nc.insertAtDepth(this.prefix, this.value, depth+1)

				// insert new node at current addr
				n.children.InsertAt(addr, nc)

				// union rec-descent new node with other node
				duplicates += nc.unionRec(otherChild, depth+1)
				continue LOOP
			}

		case *leaf[V]:
			switch this := thisChild.(type) {
			case *node[V]: // node, leaf
				clonedLeaf := otherChild.cloneLeaf()
				if this.insertAtDepth(clonedLeaf.prefix, clonedLeaf.value, depth+1) {
					duplicates++
				}
				continue LOOP

			case *leaf[V]: // leaf, leaf
				// create new node
				nc := new(node[V])

				// push this leaf down
				nc.insertAtDepth(this.prefix, this.value, depth+1)

				// insert at depth cloned leaf
				clonedLeaf := otherChild.cloneLeaf()
				if nc.insertAtDepth(clonedLeaf.prefix, clonedLeaf.value, depth+1) {
					duplicates++
				}

				// insert the new node at current addr
				n.children.InsertAt(addr, nc)
				continue LOOP
			}

		default:
			panic("logic error, wrong node type")
		}
	}

	return duplicates
}

// eachLookupPrefix does an all prefix match in the 8-bit (stride) routing table
// at this depth and calls yield() for any matching CIDR.
func (n *node[V]) eachLookupPrefix(octets []byte, depth int, is4 bool, pfxLen int, yield func(netip.Prefix, V) bool) (ok bool) {
	if n.prefixes.Len() == 0 {
		return true
	}

	// octets as array, needed below more than once
	var path stridePath
	copy(path[:], octets)

	// backtracking the CBT
	for idx := art.PfxToIdx(octets[depth], pfxLen); idx > 0; idx >>= 1 {
		if n.prefixes.Test(idx) {
			val := n.prefixes.MustGet(idx)
			cidr := cidrFromPath(path, depth, is4, idx)

			if !yield(cidr, val) {
				return false
			}
		}
	}

	return true
}

// eachSubnet calls yield() for any covered CIDR by parent prefix in natural CIDR sort order.
func (n *node[V]) eachSubnet(octets []byte, depth int, is4 bool, pfxLen int, yield func(netip.Prefix, V) bool) bool {
	// octets as array, needed below more than once
	var path stridePath
	copy(path[:], octets)

	pfxFirstAddr, pfxLastAddr := art.IdxToRange(art.PfxToIdx(octets[depth], pfxLen))

	allCoveredIndices := make([]uint, 0, maxItems)
	for _, idx := range n.prefixes.AsSlice(make([]uint, 0, maxItems)) {
		thisFirstAddr, thisLastAddr := art.IdxToRange(idx)

		if thisFirstAddr >= pfxFirstAddr && thisLastAddr <= pfxLastAddr {
			allCoveredIndices = append(allCoveredIndices, idx)
		}
	}

	// sort indices in CIDR sort order
	sort.Slice(allCoveredIndices, func(i, j int) bool {
		return lessIndexRank(allCoveredIndices[i], allCoveredIndices[j])
	})

	// 2. collect all covered child addrs by prefix

	allCoveredChildAddrs := make([]uint, 0, maxItems)
	for _, addr := range n.children.AsSlice(make([]uint, 0, maxItems)) {
		octet := uint8(addr)
		if octet >= pfxFirstAddr && octet <= pfxLastAddr {
			allCoveredChildAddrs = append(allCoveredChildAddrs, addr)
		}
	}

	// 3. yield covered indices, pathcomp prefixes and childs in CIDR sort order

	addrCursor := 0

	// yield indices and childs in CIDR sort order
	for _, pfxIdx := range allCoveredIndices {
		pfxOctet, _ := art.IdxToPfx(pfxIdx)

		// yield all childs before idx
		for j := addrCursor; j < len(allCoveredChildAddrs); j++ {
			addr := allCoveredChildAddrs[j]
			if addr >= uint(pfxOctet) {
				break
			}

			// yield the node or leaf?
			switch kid := n.children.MustGet(addr).(type) {
			case *node[V]:
				path[depth] = byte(addr)
				if !kid.allRecSorted(path, depth+1, is4, yield) {
					return false
				}

			case *leaf[V]:
				if !yield(kid.prefix, kid.value) {
					return false
				}

			default:
				panic("logic error, wrong node type")
			}

			addrCursor++
		}

		// yield the prefix for this idx
		cidr := cidrFromPath(path, depth, is4, pfxIdx)
		// n.prefixes.Items[i] not possible after sorting allIndices
		if !yield(cidr, n.prefixes.MustGet(pfxIdx)) {
			return false
		}
	}

	// yield the rest of leaves and nodes (rec-descent)
	for _, addr := range allCoveredChildAddrs[addrCursor:] {
		// yield the node or leaf?
		switch kid := n.children.MustGet(addr).(type) {
		case *node[V]:
			path[depth] = byte(addr)
			if !kid.allRecSorted(path, depth+1, is4, yield) {
				return false
			}

		case *leaf[V]:
			if !yield(kid.prefix, kid.value) {
				return false
			}

		default:
			panic("logic error, wrong node type")
		}
	}

	return true
}

// lessIndexRank, sort indexes in prefix sort order.
func lessIndexRank(aIdx, bIdx uint) bool {
	// convert idx to prefix
	aOctet, aBits := art.IdxToPfx(aIdx)
	bOctet, bBits := art.IdxToPfx(bIdx)

	// cmp the prefixes, first by address and then by bits
	if aOctet == bOctet {
		if aBits < bBits {
			return true
		}

		return false
	}

	if aOctet < bOctet {
		return true
	}

	return false
}

// cidrFromPath, helper function,
// get prefix back from stride path, depth and idx.
func cidrFromPath(path stridePath, depth int, is4 bool, idx uint) netip.Prefix {
	octet, pfxLen := art.IdxToPfx(idx)

	// set masked byte in path at depth
	path[depth] = octet

	// zero/mask the bytes after prefix bits
	for i := range path[depth+1:] {
		path[depth+1+i] = 0
	}

	// make ip addr from octets
	var ip netip.Addr
	if is4 {
		ip = netip.AddrFrom4([4]byte(path[:4]))
	} else {
		ip = netip.AddrFrom16(path)
	}

	// calc bits with pathLen and pfxLen
	bits := depth<<3 + pfxLen

	// return a normalized prefix from ip/bits
	return netip.PrefixFrom(ip, bits)
}
