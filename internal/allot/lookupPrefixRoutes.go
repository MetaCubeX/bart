// Copyright (c) 2025 Karl Gaissmaier
// SPDX-License-Identifier: MIT

package allot

import "github.com/gaissmai/bart/internal/bitset"

// IdxToPrefixRoutes as precalculated bitsets,
//
// Map the baseIndex to a bitset as a precomputed complete binary tree.
//
//	  // 1 <= idx <= 511
//		func allotRec(aTbl *bitset.BitSet, idx uint) {
//			aTbl = aTbl.Set(idx)
//			if idx > 255 {
//				return
//			}
//			allotRec(aTbl, idx<<1)
//			allotRec(aTbl, idx<<1+1)
//		}
//
// Only used for fast bitset intersections instead of
// range loops in table overlaps methods.
func IdxToPrefixRoutes(idx uint) *bitset.BitSet256 {
	return &pfxRoutesLookupTbl[uint8(idx)] // uint8() is BCE
}

// pfxRoutesLookupTbl, only the first 256 Bits, see also the hostRoutesLookupTbl for the second 256 Bits
// we split the 512 Bits to 2x256 for the BitSet256 optimizations.
var pfxRoutesLookupTbl = [256]bitset.BitSet256{
	/* idx:   0 */ {0x0, 0x0, 0x0, 0x0}, // invalid
	/* idx:   1 */ {0xfffffffffffffffe, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff}, // [1 2 3 4 5 6 7 8 9 ...
	/* idx:   2 */ {0xffff00ff0f34, 0xffffffff, 0xffffffffffffffff, 0x0}, // [2 4 5 8 9 10 11 16 17 18 19 20 21 22 23 32 33 ...
	/* idx:   3 */ {0xffff0000ff00f0c8, 0xffffffff00000000, 0x0, 0xffffffffffffffff}, // [3 6 7 12 13 14 15 24 25 26 27 28 ...
	/* idx:   4 */ {0xff000f0310, 0xffff, 0xffffffff, 0x0}, // [4 8 9 16 17 18 19 32 33 34 35 36 37 38 39 64 65 66 67 68 ...
	/* idx:   5 */ {0xff0000f00c20, 0xffff0000, 0xffffffff00000000, 0x0}, // [5 10 11 20 21 22 23 40 41 42 43 44 45 46 47 80 ...
	/* idx:   6 */ {0xff00000f003040, 0xffff00000000, 0x0, 0xffffffff}, // [6 12 13 24 25 26 27 48 49 50 51 52 53 54 55 96 ...
	/* idx:   7 */ {0xff000000f000c080, 0xffff000000000000, 0x0, 0xffffffff00000000}, // [7 14 15 28 29 30 31 56 57 58 59 ...
	/* idx:   8 */ {0xf00030100, 0xff, 0xffff, 0x0}, // [8 16 17 32 33 34 35 64 65 66 67 68 69 70 71 128 129 130 131 132 ...
	/* idx:   9 */ {0xf0000c0200, 0xff00, 0xffff0000, 0x0}, // [9 18 19 36 37 38 39 72 73 74 75 76 77 78 79 144 145 146 ...
	/* idx:  10 */ {0xf0000300400, 0xff0000, 0xffff00000000, 0x0}, // [10 20 21 40 41 42 43 80 81 82 83 84 85 86 87 160 ...
	/* idx:  11 */ {0xf00000c00800, 0xff000000, 0xffff000000000000, 0x0}, // [11 22 23 44 45 46 47 88 89 90 91 92 93 94 ...
	/* idx:  12 */ {0xf000003001000, 0xff00000000, 0x0, 0xffff}, // [12 24 25 48 49 50 51 96 97 98 99 100 101 102 103 192 ...
	/* idx:  13 */ {0xf000000c002000, 0xff0000000000, 0x0, 0xffff0000}, // [13 26 27 52 53 54 55 104 105 106 107 108 109 ...
	/* idx:  14 */ {0xf00000030004000, 0xff000000000000, 0x0, 0xffff00000000}, // [14 28 29 56 57 58 59 112 113 114 115 ...
	/* idx:  15 */ {0xf0000000c0008000, 0xff00000000000000, 0x0, 0xffff000000000000}, // [15 30 31 60 61 62 63 120 121 ...
	/* idx:  16 */ {0x300010000, 0xf, 0xff, 0x0}, // [16 32 33 64 65 66 67 128 129 130 131 132 133 134 135]
	/* idx:  17 */ {0xc00020000, 0xf0, 0xff00, 0x0}, // [17 34 35 68 69 70 71 136 137 138 139 140 141 142 143]
	/* idx:  18 */ {0x3000040000, 0xf00, 0xff0000, 0x0}, // [18 36 37 72 73 74 75 144 145 146 147 148 149 150 151]
	/* idx:  19 */ {0xc000080000, 0xf000, 0xff000000, 0x0}, // [19 38 39 76 77 78 79 152 153 154 155 156 157 158 159]
	/* idx:  20 */ {0x30000100000, 0xf0000, 0xff00000000, 0x0}, // [20 40 41 80 81 82 83 160 161 162 163 164 165 166 167]
	/* idx:  21 */ {0xc0000200000, 0xf00000, 0xff0000000000, 0x0}, // [21 42 43 84 85 86 87 168 169 170 171 172 173 174 175]
	/* idx:  22 */ {0x300000400000, 0xf000000, 0xff000000000000, 0x0}, // [22 44 45 88 89 90 91 176 177 178 179 180 181 182 183]
	/* idx:  23 */ {0xc00000800000, 0xf0000000, 0xff00000000000000, 0x0}, // [23 46 47 92 93 94 95 184 185 186 187 188 189 190 191]
	/* idx:  24 */ {0x3000001000000, 0xf00000000, 0x0, 0xff}, // [24 48 49 96 97 98 99 192 193 194 195 196 197 198 199]
	/* idx:  25 */ {0xc000002000000, 0xf000000000, 0x0, 0xff00}, // [25 50 51 100 101 102 103 200 201 202 203 204 205 206 207]
	/* idx:  26 */ {0x30000004000000, 0xf0000000000, 0x0, 0xff0000}, // [26 52 53 104 105 106 107 208 209 210 211 212 213 214 215]
	/* idx:  27 */ {0xc0000008000000, 0xf00000000000, 0x0, 0xff000000}, // [27 54 55 108 109 110 111 216 217 218 219 220 221 222 223]
	/* idx:  28 */ {0x300000010000000, 0xf000000000000, 0x0, 0xff00000000}, // [28 56 57 112 113 114 115 224 225 226 227 228 229 230 231]
	/* idx:  29 */ {0xc00000020000000, 0xf0000000000000, 0x0, 0xff0000000000}, // [29 58 59 116 117 118 119 232 233 234 235 236 237 238 239]
	/* idx:  30 */ {0x3000000040000000, 0xf00000000000000, 0x0, 0xff000000000000}, // [30 60 61 120 121 122 123 240 241 242 243 244 245 246 247]
	/* idx:  31 */ {0xc000000080000000, 0xf000000000000000, 0x0, 0xff00000000000000}, // [31 62 63 124 125 126 127 248 249 250 251 252 253 254 255]
	/* idx:  32 */ {0x100000000, 0x3, 0xf, 0x0}, // [32 64 65 128 129 130 131]
	/* idx:  33 */ {0x200000000, 0xc, 0xf0, 0x0}, // [33 66 67 132 133 134 135]
	/* idx:  34 */ {0x400000000, 0x30, 0xf00, 0x0}, // [34 68 69 136 137 138 139]
	/* idx:  35 */ {0x800000000, 0xc0, 0xf000, 0x0}, // [35 70 71 140 141 142 143]
	/* idx:  36 */ {0x1000000000, 0x300, 0xf0000, 0x0}, // [36 72 73 144 145 146 147]
	/* idx:  37 */ {0x2000000000, 0xc00, 0xf00000, 0x0}, // [37 74 75 148 149 150 151]
	/* idx:  38 */ {0x4000000000, 0x3000, 0xf000000, 0x0}, // [38 76 77 152 153 154 155]
	/* idx:  39 */ {0x8000000000, 0xc000, 0xf0000000, 0x0}, // [39 78 79 156 157 158 159]
	/* idx:  40 */ {0x10000000000, 0x30000, 0xf00000000, 0x0}, // [40 80 81 160 161 162 163]
	/* idx:  41 */ {0x20000000000, 0xc0000, 0xf000000000, 0x0}, // [41 82 83 164 165 166 167]
	/* idx:  42 */ {0x40000000000, 0x300000, 0xf0000000000, 0x0}, // [42 84 85 168 169 170 171]
	/* idx:  43 */ {0x80000000000, 0xc00000, 0xf00000000000, 0x0}, // [43 86 87 172 173 174 175]
	/* idx:  44 */ {0x100000000000, 0x3000000, 0xf000000000000, 0x0}, // [44 88 89 176 177 178 179]
	/* idx:  45 */ {0x200000000000, 0xc000000, 0xf0000000000000, 0x0}, // [45 90 91 180 181 182 183]
	/* idx:  46 */ {0x400000000000, 0x30000000, 0xf00000000000000, 0x0}, // [46 92 93 184 185 186 187]
	/* idx:  47 */ {0x800000000000, 0xc0000000, 0xf000000000000000, 0x0}, // [47 94 95 188 189 190 191]
	/* idx:  48 */ {0x1000000000000, 0x300000000, 0x0, 0xf}, // [48 96 97 192 193 194 195]
	/* idx:  49 */ {0x2000000000000, 0xc00000000, 0x0, 0xf0}, // [49 98 99 196 197 198 199]
	/* idx:  50 */ {0x4000000000000, 0x3000000000, 0x0, 0xf00}, // [50 100 101 200 201 202 203]
	/* idx:  51 */ {0x8000000000000, 0xc000000000, 0x0, 0xf000}, // [51 102 103 204 205 206 207]
	/* idx:  52 */ {0x10000000000000, 0x30000000000, 0x0, 0xf0000}, // [52 104 105 208 209 210 211]
	/* idx:  53 */ {0x20000000000000, 0xc0000000000, 0x0, 0xf00000}, // [53 106 107 212 213 214 215]
	/* idx:  54 */ {0x40000000000000, 0x300000000000, 0x0, 0xf000000}, // [54 108 109 216 217 218 219]
	/* idx:  55 */ {0x80000000000000, 0xc00000000000, 0x0, 0xf0000000}, // [55 110 111 220 221 222 223]
	/* idx:  56 */ {0x100000000000000, 0x3000000000000, 0x0, 0xf00000000}, // [56 112 113 224 225 226 227]
	/* idx:  57 */ {0x200000000000000, 0xc000000000000, 0x0, 0xf000000000}, // [57 114 115 228 229 230 231]
	/* idx:  58 */ {0x400000000000000, 0x30000000000000, 0x0, 0xf0000000000}, // [58 116 117 232 233 234 235]
	/* idx:  59 */ {0x800000000000000, 0xc0000000000000, 0x0, 0xf00000000000}, // [59 118 119 236 237 238 239]
	/* idx:  60 */ {0x1000000000000000, 0x300000000000000, 0x0, 0xf000000000000}, // [60 120 121 240 241 242 243]
	/* idx:  61 */ {0x2000000000000000, 0xc00000000000000, 0x0, 0xf0000000000000}, // [61 122 123 244 245 246 247]
	/* idx:  62 */ {0x4000000000000000, 0x3000000000000000, 0x0, 0xf00000000000000}, // [62 124 125 248 249 250 251]
	/* idx:  63 */ {0x8000000000000000, 0xc000000000000000, 0x0, 0xf000000000000000}, // [63 126 127 252 253 254 255]
	/* idx:  64 */ {0x0, 0x1, 0x3, 0x0}, // [64 128 129]
	/* idx:  65 */ {0x0, 0x2, 0xc, 0x0}, // [65 130 131]
	/* idx:  66 */ {0x0, 0x4, 0x30, 0x0}, // [66 132 133]
	/* idx:  67 */ {0x0, 0x8, 0xc0, 0x0}, // [67 134 135]
	/* idx:  68 */ {0x0, 0x10, 0x300, 0x0}, // [68 136 137]
	/* idx:  69 */ {0x0, 0x20, 0xc00, 0x0}, // [69 138 139]
	/* idx:  70 */ {0x0, 0x40, 0x3000, 0x0}, // [70 140 141]
	/* idx:  71 */ {0x0, 0x80, 0xc000, 0x0}, // [71 142 143]
	/* idx:  72 */ {0x0, 0x100, 0x30000, 0x0}, // [72 144 145]
	/* idx:  73 */ {0x0, 0x200, 0xc0000, 0x0}, // [73 146 147]
	/* idx:  74 */ {0x0, 0x400, 0x300000, 0x0}, // [74 148 149]
	/* idx:  75 */ {0x0, 0x800, 0xc00000, 0x0}, // [75 150 151]
	/* idx:  76 */ {0x0, 0x1000, 0x3000000, 0x0}, // [76 152 153]
	/* idx:  77 */ {0x0, 0x2000, 0xc000000, 0x0}, // [77 154 155]
	/* idx:  78 */ {0x0, 0x4000, 0x30000000, 0x0}, // [78 156 157]
	/* idx:  79 */ {0x0, 0x8000, 0xc0000000, 0x0}, // [79 158 159]
	/* idx:  80 */ {0x0, 0x10000, 0x300000000, 0x0}, // [80 160 161]
	/* idx:  81 */ {0x0, 0x20000, 0xc00000000, 0x0}, // [81 162 163]
	/* idx:  82 */ {0x0, 0x40000, 0x3000000000, 0x0}, // [82 164 165]
	/* idx:  83 */ {0x0, 0x80000, 0xc000000000, 0x0}, // [83 166 167]
	/* idx:  84 */ {0x0, 0x100000, 0x30000000000, 0x0}, // [84 168 169]
	/* idx:  85 */ {0x0, 0x200000, 0xc0000000000, 0x0}, // [85 170 171]
	/* idx:  86 */ {0x0, 0x400000, 0x300000000000, 0x0}, // [86 172 173]
	/* idx:  87 */ {0x0, 0x800000, 0xc00000000000, 0x0}, // [87 174 175]
	/* idx:  88 */ {0x0, 0x1000000, 0x3000000000000, 0x0}, // [88 176 177]
	/* idx:  89 */ {0x0, 0x2000000, 0xc000000000000, 0x0}, // [89 178 179]
	/* idx:  90 */ {0x0, 0x4000000, 0x30000000000000, 0x0}, // [90 180 181]
	/* idx:  91 */ {0x0, 0x8000000, 0xc0000000000000, 0x0}, // [91 182 183]
	/* idx:  92 */ {0x0, 0x10000000, 0x300000000000000, 0x0}, // [92 184 185]
	/* idx:  93 */ {0x0, 0x20000000, 0xc00000000000000, 0x0}, // [93 186 187]
	/* idx:  94 */ {0x0, 0x40000000, 0x3000000000000000, 0x0}, // [94 188 189]
	/* idx:  95 */ {0x0, 0x80000000, 0xc000000000000000, 0x0}, // [95 190 191]
	/* idx:  96 */ {0x0, 0x100000000, 0x0, 0x3}, // [96 192 193]
	/* idx:  97 */ {0x0, 0x200000000, 0x0, 0xc}, // [97 194 195]
	/* idx:  98 */ {0x0, 0x400000000, 0x0, 0x30}, // [98 196 197]
	/* idx:  99 */ {0x0, 0x800000000, 0x0, 0xc0}, // [99 198 199]
	/* idx: 100 */ {0x0, 0x1000000000, 0x0, 0x300}, // [100 200 201]
	/* idx: 101 */ {0x0, 0x2000000000, 0x0, 0xc00}, // [101 202 203]
	/* idx: 102 */ {0x0, 0x4000000000, 0x0, 0x3000}, // [102 204 205]
	/* idx: 103 */ {0x0, 0x8000000000, 0x0, 0xc000}, // [103 206 207]
	/* idx: 104 */ {0x0, 0x10000000000, 0x0, 0x30000}, // [104 208 209]
	/* idx: 105 */ {0x0, 0x20000000000, 0x0, 0xc0000}, // [105 210 211]
	/* idx: 106 */ {0x0, 0x40000000000, 0x0, 0x300000}, // [106 212 213]
	/* idx: 107 */ {0x0, 0x80000000000, 0x0, 0xc00000}, // [107 214 215]
	/* idx: 108 */ {0x0, 0x100000000000, 0x0, 0x3000000}, // [108 216 217]
	/* idx: 109 */ {0x0, 0x200000000000, 0x0, 0xc000000}, // [109 218 219]
	/* idx: 110 */ {0x0, 0x400000000000, 0x0, 0x30000000}, // [110 220 221]
	/* idx: 111 */ {0x0, 0x800000000000, 0x0, 0xc0000000}, // [111 222 223]
	/* idx: 112 */ {0x0, 0x1000000000000, 0x0, 0x300000000}, // [112 224 225]
	/* idx: 113 */ {0x0, 0x2000000000000, 0x0, 0xc00000000}, // [113 226 227]
	/* idx: 114 */ {0x0, 0x4000000000000, 0x0, 0x3000000000}, // [114 228 229]
	/* idx: 115 */ {0x0, 0x8000000000000, 0x0, 0xc000000000}, // [115 230 231]
	/* idx: 116 */ {0x0, 0x10000000000000, 0x0, 0x30000000000}, // [116 232 233]
	/* idx: 117 */ {0x0, 0x20000000000000, 0x0, 0xc0000000000}, // [117 234 235]
	/* idx: 118 */ {0x0, 0x40000000000000, 0x0, 0x300000000000}, // [118 236 237]
	/* idx: 119 */ {0x0, 0x80000000000000, 0x0, 0xc00000000000}, // [119 238 239]
	/* idx: 120 */ {0x0, 0x100000000000000, 0x0, 0x3000000000000}, // [120 240 241]
	/* idx: 121 */ {0x0, 0x200000000000000, 0x0, 0xc000000000000}, // [121 242 243]
	/* idx: 122 */ {0x0, 0x400000000000000, 0x0, 0x30000000000000}, // [122 244 245]
	/* idx: 123 */ {0x0, 0x800000000000000, 0x0, 0xc0000000000000}, // [123 246 247]
	/* idx: 124 */ {0x0, 0x1000000000000000, 0x0, 0x300000000000000}, // [124 248 249]
	/* idx: 125 */ {0x0, 0x2000000000000000, 0x0, 0xc00000000000000}, // [125 250 251]
	/* idx: 126 */ {0x0, 0x4000000000000000, 0x0, 0x3000000000000000}, // [126 252 253]
	/* idx: 127 */ {0x0, 0x8000000000000000, 0x0, 0xc000000000000000}, // [127 254 255]
	/* idx: 128 */ {0x0, 0x0, 0x1, 0x0}, // [128]
	/* idx: 129 */ {0x0, 0x0, 0x2, 0x0}, // [129]
	/* idx: 130 */ {0x0, 0x0, 0x4, 0x0}, // [130]
	/* idx: 131 */ {0x0, 0x0, 0x8, 0x0}, // [131]
	/* idx: 132 */ {0x0, 0x0, 0x10, 0x0}, // [132]
	/* idx: 133 */ {0x0, 0x0, 0x20, 0x0}, // [133]
	/* idx: 134 */ {0x0, 0x0, 0x40, 0x0}, // [134]
	/* idx: 135 */ {0x0, 0x0, 0x80, 0x0}, // [135]
	/* idx: 136 */ {0x0, 0x0, 0x100, 0x0}, // [136]
	/* idx: 137 */ {0x0, 0x0, 0x200, 0x0}, // [137]
	/* idx: 138 */ {0x0, 0x0, 0x400, 0x0}, // [138]
	/* idx: 139 */ {0x0, 0x0, 0x800, 0x0}, // [139]
	/* idx: 140 */ {0x0, 0x0, 0x1000, 0x0}, // [140]
	/* idx: 141 */ {0x0, 0x0, 0x2000, 0x0}, // [141]
	/* idx: 142 */ {0x0, 0x0, 0x4000, 0x0}, // [142]
	/* idx: 143 */ {0x0, 0x0, 0x8000, 0x0}, // [143]
	/* idx: 144 */ {0x0, 0x0, 0x10000, 0x0}, // [144]
	/* idx: 145 */ {0x0, 0x0, 0x20000, 0x0}, // [145]
	/* idx: 146 */ {0x0, 0x0, 0x40000, 0x0}, // [146]
	/* idx: 147 */ {0x0, 0x0, 0x80000, 0x0}, // [147]
	/* idx: 148 */ {0x0, 0x0, 0x100000, 0x0}, // [148]
	/* idx: 149 */ {0x0, 0x0, 0x200000, 0x0}, // [149]
	/* idx: 150 */ {0x0, 0x0, 0x400000, 0x0}, // [150]
	/* idx: 151 */ {0x0, 0x0, 0x800000, 0x0}, // [151]
	/* idx: 152 */ {0x0, 0x0, 0x1000000, 0x0}, // [152]
	/* idx: 153 */ {0x0, 0x0, 0x2000000, 0x0}, // [153]
	/* idx: 154 */ {0x0, 0x0, 0x4000000, 0x0}, // [154]
	/* idx: 155 */ {0x0, 0x0, 0x8000000, 0x0}, // [155]
	/* idx: 156 */ {0x0, 0x0, 0x10000000, 0x0}, // [156]
	/* idx: 157 */ {0x0, 0x0, 0x20000000, 0x0}, // [157]
	/* idx: 158 */ {0x0, 0x0, 0x40000000, 0x0}, // [158]
	/* idx: 159 */ {0x0, 0x0, 0x80000000, 0x0}, // [159]
	/* idx: 160 */ {0x0, 0x0, 0x100000000, 0x0}, // [160]
	/* idx: 161 */ {0x0, 0x0, 0x200000000, 0x0}, // [161]
	/* idx: 162 */ {0x0, 0x0, 0x400000000, 0x0}, // [162]
	/* idx: 163 */ {0x0, 0x0, 0x800000000, 0x0}, // [163]
	/* idx: 164 */ {0x0, 0x0, 0x1000000000, 0x0}, // [164]
	/* idx: 165 */ {0x0, 0x0, 0x2000000000, 0x0}, // [165]
	/* idx: 166 */ {0x0, 0x0, 0x4000000000, 0x0}, // [166]
	/* idx: 167 */ {0x0, 0x0, 0x8000000000, 0x0}, // [167]
	/* idx: 168 */ {0x0, 0x0, 0x10000000000, 0x0}, // [168]
	/* idx: 169 */ {0x0, 0x0, 0x20000000000, 0x0}, // [169]
	/* idx: 170 */ {0x0, 0x0, 0x40000000000, 0x0}, // [170]
	/* idx: 171 */ {0x0, 0x0, 0x80000000000, 0x0}, // [171]
	/* idx: 172 */ {0x0, 0x0, 0x100000000000, 0x0}, // [172]
	/* idx: 173 */ {0x0, 0x0, 0x200000000000, 0x0}, // [173]
	/* idx: 174 */ {0x0, 0x0, 0x400000000000, 0x0}, // [174]
	/* idx: 175 */ {0x0, 0x0, 0x800000000000, 0x0}, // [175]
	/* idx: 176 */ {0x0, 0x0, 0x1000000000000, 0x0}, // [176]
	/* idx: 177 */ {0x0, 0x0, 0x2000000000000, 0x0}, // [177]
	/* idx: 178 */ {0x0, 0x0, 0x4000000000000, 0x0}, // [178]
	/* idx: 179 */ {0x0, 0x0, 0x8000000000000, 0x0}, // [179]
	/* idx: 180 */ {0x0, 0x0, 0x10000000000000, 0x0}, // [180]
	/* idx: 181 */ {0x0, 0x0, 0x20000000000000, 0x0}, // [181]
	/* idx: 182 */ {0x0, 0x0, 0x40000000000000, 0x0}, // [182]
	/* idx: 183 */ {0x0, 0x0, 0x80000000000000, 0x0}, // [183]
	/* idx: 184 */ {0x0, 0x0, 0x100000000000000, 0x0}, // [184]
	/* idx: 185 */ {0x0, 0x0, 0x200000000000000, 0x0}, // [185]
	/* idx: 186 */ {0x0, 0x0, 0x400000000000000, 0x0}, // [186]
	/* idx: 187 */ {0x0, 0x0, 0x800000000000000, 0x0}, // [187]
	/* idx: 188 */ {0x0, 0x0, 0x1000000000000000, 0x0}, // [188]
	/* idx: 189 */ {0x0, 0x0, 0x2000000000000000, 0x0}, // [189]
	/* idx: 190 */ {0x0, 0x0, 0x4000000000000000, 0x0}, // [190]
	/* idx: 191 */ {0x0, 0x0, 0x8000000000000000, 0x0}, // [191]
	/* idx: 192 */ {0x0, 0x0, 0x0, 0x1}, // [192]
	/* idx: 193 */ {0x0, 0x0, 0x0, 0x2}, // [193]
	/* idx: 194 */ {0x0, 0x0, 0x0, 0x4}, // [194]
	/* idx: 195 */ {0x0, 0x0, 0x0, 0x8}, // [195]
	/* idx: 196 */ {0x0, 0x0, 0x0, 0x10}, // [196]
	/* idx: 197 */ {0x0, 0x0, 0x0, 0x20}, // [197]
	/* idx: 198 */ {0x0, 0x0, 0x0, 0x40}, // [198]
	/* idx: 199 */ {0x0, 0x0, 0x0, 0x80}, // [199]
	/* idx: 200 */ {0x0, 0x0, 0x0, 0x100}, // [200]
	/* idx: 201 */ {0x0, 0x0, 0x0, 0x200}, // [201]
	/* idx: 202 */ {0x0, 0x0, 0x0, 0x400}, // [202]
	/* idx: 203 */ {0x0, 0x0, 0x0, 0x800}, // [203]
	/* idx: 204 */ {0x0, 0x0, 0x0, 0x1000}, // [204]
	/* idx: 205 */ {0x0, 0x0, 0x0, 0x2000}, // [205]
	/* idx: 206 */ {0x0, 0x0, 0x0, 0x4000}, // [206]
	/* idx: 207 */ {0x0, 0x0, 0x0, 0x8000}, // [207]
	/* idx: 208 */ {0x0, 0x0, 0x0, 0x10000}, // [208]
	/* idx: 209 */ {0x0, 0x0, 0x0, 0x20000}, // [209]
	/* idx: 210 */ {0x0, 0x0, 0x0, 0x40000}, // [210]
	/* idx: 211 */ {0x0, 0x0, 0x0, 0x80000}, // [211]
	/* idx: 212 */ {0x0, 0x0, 0x0, 0x100000}, // [212]
	/* idx: 213 */ {0x0, 0x0, 0x0, 0x200000}, // [213]
	/* idx: 214 */ {0x0, 0x0, 0x0, 0x400000}, // [214]
	/* idx: 215 */ {0x0, 0x0, 0x0, 0x800000}, // [215]
	/* idx: 216 */ {0x0, 0x0, 0x0, 0x1000000}, // [216]
	/* idx: 217 */ {0x0, 0x0, 0x0, 0x2000000}, // [217]
	/* idx: 218 */ {0x0, 0x0, 0x0, 0x4000000}, // [218]
	/* idx: 219 */ {0x0, 0x0, 0x0, 0x8000000}, // [219]
	/* idx: 220 */ {0x0, 0x0, 0x0, 0x10000000}, // [220]
	/* idx: 221 */ {0x0, 0x0, 0x0, 0x20000000}, // [221]
	/* idx: 222 */ {0x0, 0x0, 0x0, 0x40000000}, // [222]
	/* idx: 223 */ {0x0, 0x0, 0x0, 0x80000000}, // [223]
	/* idx: 224 */ {0x0, 0x0, 0x0, 0x100000000}, // [224]
	/* idx: 225 */ {0x0, 0x0, 0x0, 0x200000000}, // [225]
	/* idx: 226 */ {0x0, 0x0, 0x0, 0x400000000}, // [226]
	/* idx: 227 */ {0x0, 0x0, 0x0, 0x800000000}, // [227]
	/* idx: 228 */ {0x0, 0x0, 0x0, 0x1000000000}, // [228]
	/* idx: 229 */ {0x0, 0x0, 0x0, 0x2000000000}, // [229]
	/* idx: 230 */ {0x0, 0x0, 0x0, 0x4000000000}, // [230]
	/* idx: 231 */ {0x0, 0x0, 0x0, 0x8000000000}, // [231]
	/* idx: 232 */ {0x0, 0x0, 0x0, 0x10000000000}, // [232]
	/* idx: 233 */ {0x0, 0x0, 0x0, 0x20000000000}, // [233]
	/* idx: 234 */ {0x0, 0x0, 0x0, 0x40000000000}, // [234]
	/* idx: 235 */ {0x0, 0x0, 0x0, 0x80000000000}, // [235]
	/* idx: 236 */ {0x0, 0x0, 0x0, 0x100000000000}, // [236]
	/* idx: 237 */ {0x0, 0x0, 0x0, 0x200000000000}, // [237]
	/* idx: 238 */ {0x0, 0x0, 0x0, 0x400000000000}, // [238]
	/* idx: 239 */ {0x0, 0x0, 0x0, 0x800000000000}, // [239]
	/* idx: 240 */ {0x0, 0x0, 0x0, 0x1000000000000}, // [240]
	/* idx: 241 */ {0x0, 0x0, 0x0, 0x2000000000000}, // [241]
	/* idx: 242 */ {0x0, 0x0, 0x0, 0x4000000000000}, // [242]
	/* idx: 243 */ {0x0, 0x0, 0x0, 0x8000000000000}, // [243]
	/* idx: 244 */ {0x0, 0x0, 0x0, 0x10000000000000}, // [244]
	/* idx: 245 */ {0x0, 0x0, 0x0, 0x20000000000000}, // [245]
	/* idx: 246 */ {0x0, 0x0, 0x0, 0x40000000000000}, // [246]
	/* idx: 247 */ {0x0, 0x0, 0x0, 0x80000000000000}, // [247]
	/* idx: 248 */ {0x0, 0x0, 0x0, 0x100000000000000}, // [248]
	/* idx: 249 */ {0x0, 0x0, 0x0, 0x200000000000000}, // [249]
	/* idx: 250 */ {0x0, 0x0, 0x0, 0x400000000000000}, // [250]
	/* idx: 251 */ {0x0, 0x0, 0x0, 0x800000000000000}, // [251]
	/* idx: 252 */ {0x0, 0x0, 0x0, 0x1000000000000000}, // [252]
	/* idx: 253 */ {0x0, 0x0, 0x0, 0x2000000000000000}, // [253]
	/* idx: 254 */ {0x0, 0x0, 0x0, 0x4000000000000000}, // [254]
	/* idx: 255 */ {0x0, 0x0, 0x0, 0x8000000000000000}, // [255]
}
