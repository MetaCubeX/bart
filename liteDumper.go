// Copyright (c) 2025 Karl Gaissmaier
// SPDX-License-Identifier: MIT

package bart

import (
	"fmt"
	"io"
	"strings"

	"github.com/metacubex/bart/internal/art"
)

// ##################################################
//  useful during development, debugging and testing
// ##################################################

// dumpString is just a wrapper for dump.
func (l *Lite) dumpString() string {
	w := new(strings.Builder)
	l.dump(w)

	return w.String()
}

// dump the table structure and all the nodes to w.
func (l *Lite) dump(w io.Writer) {
	if l == nil {
		return
	}

	if !l.root4.isEmpty() {
		stats := l.root4.nodeStatsRec()
		fmt.Fprintln(w)
		fmt.Fprintf(w, "### IPv4: nodes(%d), pfxs(%d), leaves(%d), fringes(%d),",
			stats.nodes, stats.pfxs, stats.leaves, stats.fringes)
		l.root4.dumpRec(w, stridePath{}, 0, true)
	}

	if !l.root6.isEmpty() {
		stats := l.root6.nodeStatsRec()
		fmt.Fprintln(w)
		fmt.Fprintf(w, "### IPv6: nodes(%d), pfxs(%d), leaves(%d), fringes(%d),",
			stats.nodes, stats.pfxs, stats.leaves, stats.fringes)
		l.root6.dumpRec(w, stridePath{}, 0, false)
	}
}

// dumpRec, rec-descent the trie.
func (n *liteNode) dumpRec(w io.Writer, path stridePath, depth int, is4 bool) {
	// dump this node
	n.dump(w, path, depth, is4)

	// the node may have childs, rec-descent down
	for i, addr := range n.children.All() {
		octet := byte(addr)
		path[depth&15] = octet

		if child, ok := n.children.Items[i].(*liteNode); ok {
			child.dumpRec(w, path, depth+1, is4)
		}
	}
}

// dump the node to w.
func (n *liteNode) dump(w io.Writer, path stridePath, depth int, is4 bool) {
	bits := depth * strideLen
	indent := strings.Repeat(".", depth)

	// node type with depth and octet path and bits.
	fmt.Fprintf(w, "\n%s[%s] depth:  %d path: [%s] / %d\n",
		indent, n.hasType(), depth, ipStridePath(path, depth, is4), bits)

	if nPfxCount := n.prefixes.Size(); nPfxCount != 0 {
		// no heap allocs
		allIndices := n.prefixes.All()

		// print the baseIndices for this node.
		fmt.Fprintf(w, "%sindexs(#%d): %v\n", indent, nPfxCount, allIndices)

		// print the prefixes for this node
		fmt.Fprintf(w, "%sprefxs(#%d):", indent, nPfxCount)

		for _, idx := range allIndices {
			octet, pfxLen := art.IdxToPfx(idx)
			fmt.Fprintf(w, " %s/%d", octetFmt(octet, is4), pfxLen)
		}

		fmt.Fprintln(w)

		/* Lite has no values
		// print the values for this node
		fmt.Fprintf(w, "%svalues(#%d):", indent, nPfxCount)

		for _, val := range n.prefixes.Items {
			fmt.Fprintf(w, " %v", val)
		}

		fmt.Fprintln(w)
		*/
	}

	if n.children.Len() != 0 {

		nodeAddrs := make([]uint, 0, maxItems)
		leafAddrs := make([]uint, 0, maxItems)
		fringeAddrs := make([]uint, 0, maxItems)

		// the node has recursive child nodes or path-compressed leaves
		for i, addr := range n.children.All() {
			switch kid := n.children.Items[i].(type) {
			case *liteNode:
				nodeAddrs = append(nodeAddrs, addr)
				continue

			case *liteLeaf:
				if kid.fringe {
					fringeAddrs = append(fringeAddrs, addr)
				} else {
					leafAddrs = append(leafAddrs, addr)
				}

			default:
				panic("logic error, wrong node type")
			}
		}

		if nodeCount := len(nodeAddrs); nodeCount > 0 {
			// print the childs for this node
			fmt.Fprintf(w, "%schilds(#%d):", indent, nodeCount)

			for _, addr := range nodeAddrs {
				octet := byte(addr)
				fmt.Fprintf(w, " %s", octetFmt(octet, is4))
			}

			fmt.Fprintln(w)
		}

		if leafCount := len(leafAddrs); leafCount > 0 {
			// print the pathcomp prefixes for this node
			fmt.Fprintf(w, "%sleaves(#%d):", indent, leafCount)

			for _, addr := range leafAddrs {
				octet := byte(addr)
				k := n.children.MustGet(addr)
				pc := k.(*liteLeaf)

				fmt.Fprintf(w, " %s:{%s}", octetFmt(octet, is4), pc.prefix)
			}
			fmt.Fprintln(w)
		}

		if fringeCount := len(fringeAddrs); fringeCount > 0 {
			// print the pathcomp prefixes for this node
			fmt.Fprintf(w, "%sfringe(#%d):", indent, fringeCount)

			for _, addr := range fringeAddrs {
				octet := byte(addr)
				k := n.children.MustGet(addr)
				pc := k.(*liteLeaf)

				fmt.Fprintf(w, " %s:{%s}", octetFmt(octet, is4), pc.prefix)
			}
			fmt.Fprintln(w)
		}
	}
}

// hasType returns the nodeType.
func (n *liteNode) hasType() nodeType {
	s := n.nodeStats()

	switch {
	case s.pfxs == 0 && s.childs == 0:
		return nullNode
	case s.nodes == 0:
		return leafNode
	case (s.pfxs > 0 || s.leaves > 0 || s.fringes > 0) && s.nodes > 0:
		return fullNode
	case (s.pfxs == 0 && s.leaves == 0 && s.fringes == 0) && s.nodes > 0:
		return intermediateNode
	default:
		panic(fmt.Sprintf("UNREACHABLE: pfx: %d, chld: %d, node: %d, leaf: %d, fringe: %d",
			s.pfxs, s.childs, s.nodes, s.leaves, s.fringes))
	}
}

// node statistics for this single node
func (n *liteNode) nodeStats() stats {
	var s stats

	s.pfxs = n.prefixes.Size()
	s.childs = n.children.Len()

	for i := range n.children.All() {
		switch kid := n.children.Items[i].(type) {
		case *liteNode:
			s.nodes++

		case *liteLeaf:
			if kid.fringe {
				s.fringes++
			} else {
				s.leaves++
			}

		default:
			panic("logic error, wrong node type")
		}
	}

	return s
}

// nodeStatsRec, calculate the number of pfxs, nodes and leaves under n, rec-descent.
func (n *liteNode) nodeStatsRec() stats {
	var s stats
	if n == nil || n.isEmpty() {
		return s
	}

	s.pfxs = n.prefixes.Size()
	s.childs = n.children.Len()
	s.nodes = 1 // this node
	s.leaves = 0
	s.fringes = 0

	for _, kidAny := range n.children.Items {
		switch kid := kidAny.(type) {
		case *liteNode:
			// rec-descent
			rs := kid.nodeStatsRec()

			s.pfxs += rs.pfxs
			s.childs += rs.childs
			s.nodes += rs.nodes
			s.leaves += rs.leaves
			s.fringes += rs.fringes

		case *liteLeaf:
			if kid.fringe {
				s.fringes++
			} else {
				s.leaves++
			}

		default:
			panic("logic error, wrong node type")
		}
	}

	return s
}
