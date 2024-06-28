// Copyright (c) 2024 Karl Gaissmaier
// SPDX-License-Identifier: MIT

package bart

// Please read the ART paper ./doc/artlookup.pdf
// to understand the baseIndex algorithm.

// hostMasks as lookup table
var hostMasks = []uint8{
	0b1111_1111, // bits == 0
	0b0111_1111, // bits == 1
	0b0011_1111, // bits == 2
	0b0001_1111, // bits == 3
	0b0000_1111, // bits == 4
	0b0000_0111, // bits == 5
	0b0000_0011, // bits == 6
	0b0000_0001, // bits == 7
	0b0000_0000, // bits == 8
}

func netMask(mask int) uint8 {
	return ^hostMasks[uint8(mask)]
}

const (

	// baseIndex of the first host route: prefixToBaseIndex(0,8)
	firstHostIndex = 0b1_0000_0000 // 256

	// baseIndex of the last host route: prefixToBaseIndex(255,8)
	lastHostIndex = 0b1_1111_1111 // 511
)

// prefixToBaseIndex, maps a prefix table as a 'complete binary tree'.
// This is the so-called baseIndex a.k.a heapFunc:
func prefixToBaseIndex(octet byte, prefixLen int) uint {
	return uint(octet>>(strideLen-prefixLen)) + (1 << prefixLen)
}

// octetToBaseIndex, just prefixToBaseIndex(octet, 8), a.k.a host routes
// but faster, use it for host routes in Lookup.
func octetToBaseIndex(octet byte) uint {
	return uint(octet) + firstHostIndex // just: octet + 256
}

// baseIndexToPrefixMask, calc the bits from baseIndex and octect depth
func baseIndexToPrefixMask(baseIdx uint, depth int) int {
	_, pfxLen := baseIndexToPrefix(baseIdx)
	return depth*strideLen + pfxLen
}

// hostRoutesByIndex, get range of host routes for this idx.
//
//	idx:    72
//	prefix: 32/6
//	lower:  256 + 32 = 288
//	upper:  256 + (32 | 0b0000_0011) = 291
//
// Use the pre computed lookup table.
//
//	 func hostRoutesByIndex(idx uint) (uint, uint) {
//		 octet, bits := baseIndexToPrefix(idx)
//		 return octetToBaseIndex(octet), octetToBaseIndex(octet | hostMasks[bits])
//	 }
func hostRoutesByIndex(idx uint) (uint, uint) {
	item := baseIdxLookupTbl[idx]
	return uint(item.lower), uint(item.upper)
}

// baseIndexToPrefix returns the octet and prefix len of baseIdx.
// It's the inverse to prefixToBaseIndex.
//
// Use the pre computed lookup table, bits.LeadingZeros is too slow.
//
//	func baseIndexToPrefix(baseIdx uint) (octet byte, pfxLen int) {
//		nlz := bits.LeadingZeros(baseIdx)
//		pfxLen = strconv.IntSize - nlz - 1
//		octet = (baseIdx & (0xFF >> (8 - pfxLen))) << (8 - pfxLen)
//		return octet, pfxLen
//	}
func baseIndexToPrefix(baseIdx uint) (octet byte, pfxLen int) {
	item := baseIdxLookupTbl[baseIdx]
	return item.octet, int(item.bits)
}

// prefixSortRankByIndex, get the prefix sort rank for baseIndex.
// Use the pre computed lookup table.
func prefixSortRankByIndex(baseIdx uint) int {
	return int(baseIdxLookupTbl[baseIdx].rank)
}

// baseIdxLookupTbl
//
//	octet, bits,
//	host route boundaries,
//	prefix sort rank
//
// as lookup table.
var baseIdxLookupTbl = [512]struct {
	octet byte
	bits  int8
	lower uint16 // host route lower bound
	upper uint16 // host route upper bound
	rank  uint16 // prefix sort rank
}{
	{0, -1, 0, 0, 0},        // idx == 0 invalid!
	{0, 0, 256, 511, 1},     // idx == 1
	{0, 1, 256, 383, 2},     // idx == 2
	{128, 1, 384, 511, 257}, // idx == 3
	{0, 2, 256, 319, 3},     // idx == 4
	{64, 2, 320, 383, 130},  // idx == 5
	{128, 2, 384, 447, 258}, // idx == 6
	{192, 2, 448, 511, 385}, // idx == 7
	{0, 3, 256, 287, 4},     // idx == 8
	{32, 3, 288, 319, 67},   // idx == 9
	{64, 3, 320, 351, 131},  // idx == 10
	{96, 3, 352, 383, 194},  // idx == 11
	{128, 3, 384, 415, 259}, // idx == 12
	{160, 3, 416, 447, 322}, // idx == 13
	{192, 3, 448, 479, 386}, // idx == 14
	{224, 3, 480, 511, 449}, // idx == 15
	{0, 4, 256, 271, 5},     // idx == 16
	{16, 4, 272, 287, 36},   // idx == 17
	{32, 4, 288, 303, 68},   // idx == 18
	{48, 4, 304, 319, 99},   // idx == 19
	{64, 4, 320, 335, 132},  // idx == 20
	{80, 4, 336, 351, 163},  // idx == 21
	{96, 4, 352, 367, 195},  // idx == 22
	{112, 4, 368, 383, 226}, // idx == 23
	{128, 4, 384, 399, 260}, // idx == 24
	{144, 4, 400, 415, 291}, // idx == 25
	{160, 4, 416, 431, 323}, // idx == 26
	{176, 4, 432, 447, 354}, // idx == 27
	{192, 4, 448, 463, 387}, // idx == 28
	{208, 4, 464, 479, 418}, // idx == 29
	{224, 4, 480, 495, 450}, // idx == 30
	{240, 4, 496, 511, 481}, // idx == 31
	{0, 5, 256, 263, 6},     // idx == 32
	{8, 5, 264, 271, 21},    // idx == 33
	{16, 5, 272, 279, 37},   // idx == 34
	{24, 5, 280, 287, 52},   // idx == 35
	{32, 5, 288, 295, 69},   // idx == 36
	{40, 5, 296, 303, 84},   // idx == 37
	{48, 5, 304, 311, 100},  // idx == 38
	{56, 5, 312, 319, 115},  // idx == 39
	{64, 5, 320, 327, 133},  // idx == 40
	{72, 5, 328, 335, 148},  // idx == 41
	{80, 5, 336, 343, 164},  // idx == 42
	{88, 5, 344, 351, 179},  // idx == 43
	{96, 5, 352, 359, 196},  // idx == 44
	{104, 5, 360, 367, 211}, // idx == 45
	{112, 5, 368, 375, 227}, // idx == 46
	{120, 5, 376, 383, 242}, // idx == 47
	{128, 5, 384, 391, 261}, // idx == 48
	{136, 5, 392, 399, 276}, // idx == 49
	{144, 5, 400, 407, 292}, // idx == 50
	{152, 5, 408, 415, 307}, // idx == 51
	{160, 5, 416, 423, 324}, // idx == 52
	{168, 5, 424, 431, 339}, // idx == 53
	{176, 5, 432, 439, 355}, // idx == 54
	{184, 5, 440, 447, 370}, // idx == 55
	{192, 5, 448, 455, 388}, // idx == 56
	{200, 5, 456, 463, 403}, // idx == 57
	{208, 5, 464, 471, 419}, // idx == 58
	{216, 5, 472, 479, 434}, // idx == 59
	{224, 5, 480, 487, 451}, // idx == 60
	{232, 5, 488, 495, 466}, // idx == 61
	{240, 5, 496, 503, 482}, // idx == 62
	{248, 5, 504, 511, 497}, // idx == 63
	{0, 6, 256, 259, 7},     // idx == 64
	{4, 6, 260, 263, 14},    // idx == 65
	{8, 6, 264, 267, 22},    // idx == 66
	{12, 6, 268, 271, 29},   // idx == 67
	{16, 6, 272, 275, 38},   // idx == 68
	{20, 6, 276, 279, 45},   // idx == 69
	{24, 6, 280, 283, 53},   // idx == 70
	{28, 6, 284, 287, 60},   // idx == 71
	{32, 6, 288, 291, 70},   // idx == 72
	{36, 6, 292, 295, 77},   // idx == 73
	{40, 6, 296, 299, 85},   // idx == 74
	{44, 6, 300, 303, 92},   // idx == 75
	{48, 6, 304, 307, 101},  // idx == 76
	{52, 6, 308, 311, 108},  // idx == 77
	{56, 6, 312, 315, 116},  // idx == 78
	{60, 6, 316, 319, 123},  // idx == 79
	{64, 6, 320, 323, 134},  // idx == 80
	{68, 6, 324, 327, 141},  // idx == 81
	{72, 6, 328, 331, 149},  // idx == 82
	{76, 6, 332, 335, 156},  // idx == 83
	{80, 6, 336, 339, 165},  // idx == 84
	{84, 6, 340, 343, 172},  // idx == 85
	{88, 6, 344, 347, 180},  // idx == 86
	{92, 6, 348, 351, 187},  // idx == 87
	{96, 6, 352, 355, 197},  // idx == 88
	{100, 6, 356, 359, 204}, // idx == 89
	{104, 6, 360, 363, 212}, // idx == 90
	{108, 6, 364, 367, 219}, // idx == 91
	{112, 6, 368, 371, 228}, // idx == 92
	{116, 6, 372, 375, 235}, // idx == 93
	{120, 6, 376, 379, 243}, // idx == 94
	{124, 6, 380, 383, 250}, // idx == 95
	{128, 6, 384, 387, 262}, // idx == 96
	{132, 6, 388, 391, 269}, // idx == 97
	{136, 6, 392, 395, 277}, // idx == 98
	{140, 6, 396, 399, 284}, // idx == 99
	{144, 6, 400, 403, 293}, // idx == 100
	{148, 6, 404, 407, 300}, // idx == 101
	{152, 6, 408, 411, 308}, // idx == 102
	{156, 6, 412, 415, 315}, // idx == 103
	{160, 6, 416, 419, 325}, // idx == 104
	{164, 6, 420, 423, 332}, // idx == 105
	{168, 6, 424, 427, 340}, // idx == 106
	{172, 6, 428, 431, 347}, // idx == 107
	{176, 6, 432, 435, 356}, // idx == 108
	{180, 6, 436, 439, 363}, // idx == 109
	{184, 6, 440, 443, 371}, // idx == 110
	{188, 6, 444, 447, 378}, // idx == 111
	{192, 6, 448, 451, 389}, // idx == 112
	{196, 6, 452, 455, 396}, // idx == 113
	{200, 6, 456, 459, 404}, // idx == 114
	{204, 6, 460, 463, 411}, // idx == 115
	{208, 6, 464, 467, 420}, // idx == 116
	{212, 6, 468, 471, 427}, // idx == 117
	{216, 6, 472, 475, 435}, // idx == 118
	{220, 6, 476, 479, 442}, // idx == 119
	{224, 6, 480, 483, 452}, // idx == 120
	{228, 6, 484, 487, 459}, // idx == 121
	{232, 6, 488, 491, 467}, // idx == 122
	{236, 6, 492, 495, 474}, // idx == 123
	{240, 6, 496, 499, 483}, // idx == 124
	{244, 6, 500, 503, 490}, // idx == 125
	{248, 6, 504, 507, 498}, // idx == 126
	{252, 6, 508, 511, 505}, // idx == 127
	{0, 7, 256, 257, 8},     // idx == 128
	{2, 7, 258, 259, 11},    // idx == 129
	{4, 7, 260, 261, 15},    // idx == 130
	{6, 7, 262, 263, 18},    // idx == 131
	{8, 7, 264, 265, 23},    // idx == 132
	{10, 7, 266, 267, 26},   // idx == 133
	{12, 7, 268, 269, 30},   // idx == 134
	{14, 7, 270, 271, 33},   // idx == 135
	{16, 7, 272, 273, 39},   // idx == 136
	{18, 7, 274, 275, 42},   // idx == 137
	{20, 7, 276, 277, 46},   // idx == 138
	{22, 7, 278, 279, 49},   // idx == 139
	{24, 7, 280, 281, 54},   // idx == 140
	{26, 7, 282, 283, 57},   // idx == 141
	{28, 7, 284, 285, 61},   // idx == 142
	{30, 7, 286, 287, 64},   // idx == 143
	{32, 7, 288, 289, 71},   // idx == 144
	{34, 7, 290, 291, 74},   // idx == 145
	{36, 7, 292, 293, 78},   // idx == 146
	{38, 7, 294, 295, 81},   // idx == 147
	{40, 7, 296, 297, 86},   // idx == 148
	{42, 7, 298, 299, 89},   // idx == 149
	{44, 7, 300, 301, 93},   // idx == 150
	{46, 7, 302, 303, 96},   // idx == 151
	{48, 7, 304, 305, 102},  // idx == 152
	{50, 7, 306, 307, 105},  // idx == 153
	{52, 7, 308, 309, 109},  // idx == 154
	{54, 7, 310, 311, 112},  // idx == 155
	{56, 7, 312, 313, 117},  // idx == 156
	{58, 7, 314, 315, 120},  // idx == 157
	{60, 7, 316, 317, 124},  // idx == 158
	{62, 7, 318, 319, 127},  // idx == 159
	{64, 7, 320, 321, 135},  // idx == 160
	{66, 7, 322, 323, 138},  // idx == 161
	{68, 7, 324, 325, 142},  // idx == 162
	{70, 7, 326, 327, 145},  // idx == 163
	{72, 7, 328, 329, 150},  // idx == 164
	{74, 7, 330, 331, 153},  // idx == 165
	{76, 7, 332, 333, 157},  // idx == 166
	{78, 7, 334, 335, 160},  // idx == 167
	{80, 7, 336, 337, 166},  // idx == 168
	{82, 7, 338, 339, 169},  // idx == 169
	{84, 7, 340, 341, 173},  // idx == 170
	{86, 7, 342, 343, 176},  // idx == 171
	{88, 7, 344, 345, 181},  // idx == 172
	{90, 7, 346, 347, 184},  // idx == 173
	{92, 7, 348, 349, 188},  // idx == 174
	{94, 7, 350, 351, 191},  // idx == 175
	{96, 7, 352, 353, 198},  // idx == 176
	{98, 7, 354, 355, 201},  // idx == 177
	{100, 7, 356, 357, 205}, // idx == 178
	{102, 7, 358, 359, 208}, // idx == 179
	{104, 7, 360, 361, 213}, // idx == 180
	{106, 7, 362, 363, 216}, // idx == 181
	{108, 7, 364, 365, 220}, // idx == 182
	{110, 7, 366, 367, 223}, // idx == 183
	{112, 7, 368, 369, 229}, // idx == 184
	{114, 7, 370, 371, 232}, // idx == 185
	{116, 7, 372, 373, 236}, // idx == 186
	{118, 7, 374, 375, 239}, // idx == 187
	{120, 7, 376, 377, 244}, // idx == 188
	{122, 7, 378, 379, 247}, // idx == 189
	{124, 7, 380, 381, 251}, // idx == 190
	{126, 7, 382, 383, 254}, // idx == 191
	{128, 7, 384, 385, 263}, // idx == 192
	{130, 7, 386, 387, 266}, // idx == 193
	{132, 7, 388, 389, 270}, // idx == 194
	{134, 7, 390, 391, 273}, // idx == 195
	{136, 7, 392, 393, 278}, // idx == 196
	{138, 7, 394, 395, 281}, // idx == 197
	{140, 7, 396, 397, 285}, // idx == 198
	{142, 7, 398, 399, 288}, // idx == 199
	{144, 7, 400, 401, 294}, // idx == 200
	{146, 7, 402, 403, 297}, // idx == 201
	{148, 7, 404, 405, 301}, // idx == 202
	{150, 7, 406, 407, 304}, // idx == 203
	{152, 7, 408, 409, 309}, // idx == 204
	{154, 7, 410, 411, 312}, // idx == 205
	{156, 7, 412, 413, 316}, // idx == 206
	{158, 7, 414, 415, 319}, // idx == 207
	{160, 7, 416, 417, 326}, // idx == 208
	{162, 7, 418, 419, 329}, // idx == 209
	{164, 7, 420, 421, 333}, // idx == 210
	{166, 7, 422, 423, 336}, // idx == 211
	{168, 7, 424, 425, 341}, // idx == 212
	{170, 7, 426, 427, 344}, // idx == 213
	{172, 7, 428, 429, 348}, // idx == 214
	{174, 7, 430, 431, 351}, // idx == 215
	{176, 7, 432, 433, 357}, // idx == 216
	{178, 7, 434, 435, 360}, // idx == 217
	{180, 7, 436, 437, 364}, // idx == 218
	{182, 7, 438, 439, 367}, // idx == 219
	{184, 7, 440, 441, 372}, // idx == 220
	{186, 7, 442, 443, 375}, // idx == 221
	{188, 7, 444, 445, 379}, // idx == 222
	{190, 7, 446, 447, 382}, // idx == 223
	{192, 7, 448, 449, 390}, // idx == 224
	{194, 7, 450, 451, 393}, // idx == 225
	{196, 7, 452, 453, 397}, // idx == 226
	{198, 7, 454, 455, 400}, // idx == 227
	{200, 7, 456, 457, 405}, // idx == 228
	{202, 7, 458, 459, 408}, // idx == 229
	{204, 7, 460, 461, 412}, // idx == 230
	{206, 7, 462, 463, 415}, // idx == 231
	{208, 7, 464, 465, 421}, // idx == 232
	{210, 7, 466, 467, 424}, // idx == 233
	{212, 7, 468, 469, 428}, // idx == 234
	{214, 7, 470, 471, 431}, // idx == 235
	{216, 7, 472, 473, 436}, // idx == 236
	{218, 7, 474, 475, 439}, // idx == 237
	{220, 7, 476, 477, 443}, // idx == 238
	{222, 7, 478, 479, 446}, // idx == 239
	{224, 7, 480, 481, 453}, // idx == 240
	{226, 7, 482, 483, 456}, // idx == 241
	{228, 7, 484, 485, 460}, // idx == 242
	{230, 7, 486, 487, 463}, // idx == 243
	{232, 7, 488, 489, 468}, // idx == 244
	{234, 7, 490, 491, 471}, // idx == 245
	{236, 7, 492, 493, 475}, // idx == 246
	{238, 7, 494, 495, 478}, // idx == 247
	{240, 7, 496, 497, 484}, // idx == 248
	{242, 7, 498, 499, 487}, // idx == 249
	{244, 7, 500, 501, 491}, // idx == 250
	{246, 7, 502, 503, 494}, // idx == 251
	{248, 7, 504, 505, 499}, // idx == 252
	{250, 7, 506, 507, 502}, // idx == 253
	{252, 7, 508, 509, 506}, // idx == 254
	{254, 7, 510, 511, 509}, // idx == 255
	{0, 8, 256, 256, 9},     // idx == 256 -- first host route
	{1, 8, 257, 257, 10},    // idx == 257
	{2, 8, 258, 258, 12},    // idx == 258
	{3, 8, 259, 259, 13},    // idx == 259
	{4, 8, 260, 260, 16},    // idx == 260
	{5, 8, 261, 261, 17},    // idx == 261
	{6, 8, 262, 262, 19},    // idx == 262
	{7, 8, 263, 263, 20},    // idx == 263
	{8, 8, 264, 264, 24},    // idx == 264
	{9, 8, 265, 265, 25},    // idx == 265
	{10, 8, 266, 266, 27},   // idx == 266
	{11, 8, 267, 267, 28},   // idx == 267
	{12, 8, 268, 268, 31},   // idx == 268
	{13, 8, 269, 269, 32},   // idx == 269
	{14, 8, 270, 270, 34},   // idx == 270
	{15, 8, 271, 271, 35},   // idx == 271
	{16, 8, 272, 272, 40},   // idx == 272
	{17, 8, 273, 273, 41},   // idx == 273
	{18, 8, 274, 274, 43},   // idx == 274
	{19, 8, 275, 275, 44},   // idx == 275
	{20, 8, 276, 276, 47},   // idx == 276
	{21, 8, 277, 277, 48},   // idx == 277
	{22, 8, 278, 278, 50},   // idx == 278
	{23, 8, 279, 279, 51},   // idx == 279
	{24, 8, 280, 280, 55},   // idx == 280
	{25, 8, 281, 281, 56},   // idx == 281
	{26, 8, 282, 282, 58},   // idx == 282
	{27, 8, 283, 283, 59},   // idx == 283
	{28, 8, 284, 284, 62},   // idx == 284
	{29, 8, 285, 285, 63},   // idx == 285
	{30, 8, 286, 286, 65},   // idx == 286
	{31, 8, 287, 287, 66},   // idx == 287
	{32, 8, 288, 288, 72},   // idx == 288
	{33, 8, 289, 289, 73},   // idx == 289
	{34, 8, 290, 290, 75},   // idx == 290
	{35, 8, 291, 291, 76},   // idx == 291
	{36, 8, 292, 292, 79},   // idx == 292
	{37, 8, 293, 293, 80},   // idx == 293
	{38, 8, 294, 294, 82},   // idx == 294
	{39, 8, 295, 295, 83},   // idx == 295
	{40, 8, 296, 296, 87},   // idx == 296
	{41, 8, 297, 297, 88},   // idx == 297
	{42, 8, 298, 298, 90},   // idx == 298
	{43, 8, 299, 299, 91},   // idx == 299
	{44, 8, 300, 300, 94},   // idx == 300
	{45, 8, 301, 301, 95},   // idx == 301
	{46, 8, 302, 302, 97},   // idx == 302
	{47, 8, 303, 303, 98},   // idx == 303
	{48, 8, 304, 304, 103},  // idx == 304
	{49, 8, 305, 305, 104},  // idx == 305
	{50, 8, 306, 306, 106},  // idx == 306
	{51, 8, 307, 307, 107},  // idx == 307
	{52, 8, 308, 308, 110},  // idx == 308
	{53, 8, 309, 309, 111},  // idx == 309
	{54, 8, 310, 310, 113},  // idx == 310
	{55, 8, 311, 311, 114},  // idx == 311
	{56, 8, 312, 312, 118},  // idx == 312
	{57, 8, 313, 313, 119},  // idx == 313
	{58, 8, 314, 314, 121},  // idx == 314
	{59, 8, 315, 315, 122},  // idx == 315
	{60, 8, 316, 316, 125},  // idx == 316
	{61, 8, 317, 317, 126},  // idx == 317
	{62, 8, 318, 318, 128},  // idx == 318
	{63, 8, 319, 319, 129},  // idx == 319
	{64, 8, 320, 320, 136},  // idx == 320
	{65, 8, 321, 321, 137},  // idx == 321
	{66, 8, 322, 322, 139},  // idx == 322
	{67, 8, 323, 323, 140},  // idx == 323
	{68, 8, 324, 324, 143},  // idx == 324
	{69, 8, 325, 325, 144},  // idx == 325
	{70, 8, 326, 326, 146},  // idx == 326
	{71, 8, 327, 327, 147},  // idx == 327
	{72, 8, 328, 328, 151},  // idx == 328
	{73, 8, 329, 329, 152},  // idx == 329
	{74, 8, 330, 330, 154},  // idx == 330
	{75, 8, 331, 331, 155},  // idx == 331
	{76, 8, 332, 332, 158},  // idx == 332
	{77, 8, 333, 333, 159},  // idx == 333
	{78, 8, 334, 334, 161},  // idx == 334
	{79, 8, 335, 335, 162},  // idx == 335
	{80, 8, 336, 336, 167},  // idx == 336
	{81, 8, 337, 337, 168},  // idx == 337
	{82, 8, 338, 338, 170},  // idx == 338
	{83, 8, 339, 339, 171},  // idx == 339
	{84, 8, 340, 340, 174},  // idx == 340
	{85, 8, 341, 341, 175},  // idx == 341
	{86, 8, 342, 342, 177},  // idx == 342
	{87, 8, 343, 343, 178},  // idx == 343
	{88, 8, 344, 344, 182},  // idx == 344
	{89, 8, 345, 345, 183},  // idx == 345
	{90, 8, 346, 346, 185},  // idx == 346
	{91, 8, 347, 347, 186},  // idx == 347
	{92, 8, 348, 348, 189},  // idx == 348
	{93, 8, 349, 349, 190},  // idx == 349
	{94, 8, 350, 350, 192},  // idx == 350
	{95, 8, 351, 351, 193},  // idx == 351
	{96, 8, 352, 352, 199},  // idx == 352
	{97, 8, 353, 353, 200},  // idx == 353
	{98, 8, 354, 354, 202},  // idx == 354
	{99, 8, 355, 355, 203},  // idx == 355
	{100, 8, 356, 356, 206}, // idx == 356
	{101, 8, 357, 357, 207}, // idx == 357
	{102, 8, 358, 358, 209}, // idx == 358
	{103, 8, 359, 359, 210}, // idx == 359
	{104, 8, 360, 360, 214}, // idx == 360
	{105, 8, 361, 361, 215}, // idx == 361
	{106, 8, 362, 362, 217}, // idx == 362
	{107, 8, 363, 363, 218}, // idx == 363
	{108, 8, 364, 364, 221}, // idx == 364
	{109, 8, 365, 365, 222}, // idx == 365
	{110, 8, 366, 366, 224}, // idx == 366
	{111, 8, 367, 367, 225}, // idx == 367
	{112, 8, 368, 368, 230}, // idx == 368
	{113, 8, 369, 369, 231}, // idx == 369
	{114, 8, 370, 370, 233}, // idx == 370
	{115, 8, 371, 371, 234}, // idx == 371
	{116, 8, 372, 372, 237}, // idx == 372
	{117, 8, 373, 373, 238}, // idx == 373
	{118, 8, 374, 374, 240}, // idx == 374
	{119, 8, 375, 375, 241}, // idx == 375
	{120, 8, 376, 376, 245}, // idx == 376
	{121, 8, 377, 377, 246}, // idx == 377
	{122, 8, 378, 378, 248}, // idx == 378
	{123, 8, 379, 379, 249}, // idx == 379
	{124, 8, 380, 380, 252}, // idx == 380
	{125, 8, 381, 381, 253}, // idx == 381
	{126, 8, 382, 382, 255}, // idx == 382
	{127, 8, 383, 383, 256}, // idx == 383
	{128, 8, 384, 384, 264}, // idx == 384
	{129, 8, 385, 385, 265}, // idx == 385
	{130, 8, 386, 386, 267}, // idx == 386
	{131, 8, 387, 387, 268}, // idx == 387
	{132, 8, 388, 388, 271}, // idx == 388
	{133, 8, 389, 389, 272}, // idx == 389
	{134, 8, 390, 390, 274}, // idx == 390
	{135, 8, 391, 391, 275}, // idx == 391
	{136, 8, 392, 392, 279}, // idx == 392
	{137, 8, 393, 393, 280}, // idx == 393
	{138, 8, 394, 394, 282}, // idx == 394
	{139, 8, 395, 395, 283}, // idx == 395
	{140, 8, 396, 396, 286}, // idx == 396
	{141, 8, 397, 397, 287}, // idx == 397
	{142, 8, 398, 398, 289}, // idx == 398
	{143, 8, 399, 399, 290}, // idx == 399
	{144, 8, 400, 400, 295}, // idx == 400
	{145, 8, 401, 401, 296}, // idx == 401
	{146, 8, 402, 402, 298}, // idx == 402
	{147, 8, 403, 403, 299}, // idx == 403
	{148, 8, 404, 404, 302}, // idx == 404
	{149, 8, 405, 405, 303}, // idx == 405
	{150, 8, 406, 406, 305}, // idx == 406
	{151, 8, 407, 407, 306}, // idx == 407
	{152, 8, 408, 408, 310}, // idx == 408
	{153, 8, 409, 409, 311}, // idx == 409
	{154, 8, 410, 410, 313}, // idx == 410
	{155, 8, 411, 411, 314}, // idx == 411
	{156, 8, 412, 412, 317}, // idx == 412
	{157, 8, 413, 413, 318}, // idx == 413
	{158, 8, 414, 414, 320}, // idx == 414
	{159, 8, 415, 415, 321}, // idx == 415
	{160, 8, 416, 416, 327}, // idx == 416
	{161, 8, 417, 417, 328}, // idx == 417
	{162, 8, 418, 418, 330}, // idx == 418
	{163, 8, 419, 419, 331}, // idx == 419
	{164, 8, 420, 420, 334}, // idx == 420
	{165, 8, 421, 421, 335}, // idx == 421
	{166, 8, 422, 422, 337}, // idx == 422
	{167, 8, 423, 423, 338}, // idx == 423
	{168, 8, 424, 424, 342}, // idx == 424
	{169, 8, 425, 425, 343}, // idx == 425
	{170, 8, 426, 426, 345}, // idx == 426
	{171, 8, 427, 427, 346}, // idx == 427
	{172, 8, 428, 428, 349}, // idx == 428
	{173, 8, 429, 429, 350}, // idx == 429
	{174, 8, 430, 430, 352}, // idx == 430
	{175, 8, 431, 431, 353}, // idx == 431
	{176, 8, 432, 432, 358}, // idx == 432
	{177, 8, 433, 433, 359}, // idx == 433
	{178, 8, 434, 434, 361}, // idx == 434
	{179, 8, 435, 435, 362}, // idx == 435
	{180, 8, 436, 436, 365}, // idx == 436
	{181, 8, 437, 437, 366}, // idx == 437
	{182, 8, 438, 438, 368}, // idx == 438
	{183, 8, 439, 439, 369}, // idx == 439
	{184, 8, 440, 440, 373}, // idx == 440
	{185, 8, 441, 441, 374}, // idx == 441
	{186, 8, 442, 442, 376}, // idx == 442
	{187, 8, 443, 443, 377}, // idx == 443
	{188, 8, 444, 444, 380}, // idx == 444
	{189, 8, 445, 445, 381}, // idx == 445
	{190, 8, 446, 446, 383}, // idx == 446
	{191, 8, 447, 447, 384}, // idx == 447
	{192, 8, 448, 448, 391}, // idx == 448
	{193, 8, 449, 449, 392}, // idx == 449
	{194, 8, 450, 450, 394}, // idx == 450
	{195, 8, 451, 451, 395}, // idx == 451
	{196, 8, 452, 452, 398}, // idx == 452
	{197, 8, 453, 453, 399}, // idx == 453
	{198, 8, 454, 454, 401}, // idx == 454
	{199, 8, 455, 455, 402}, // idx == 455
	{200, 8, 456, 456, 406}, // idx == 456
	{201, 8, 457, 457, 407}, // idx == 457
	{202, 8, 458, 458, 409}, // idx == 458
	{203, 8, 459, 459, 410}, // idx == 459
	{204, 8, 460, 460, 413}, // idx == 460
	{205, 8, 461, 461, 414}, // idx == 461
	{206, 8, 462, 462, 416}, // idx == 462
	{207, 8, 463, 463, 417}, // idx == 463
	{208, 8, 464, 464, 422}, // idx == 464
	{209, 8, 465, 465, 423}, // idx == 465
	{210, 8, 466, 466, 425}, // idx == 466
	{211, 8, 467, 467, 426}, // idx == 467
	{212, 8, 468, 468, 429}, // idx == 468
	{213, 8, 469, 469, 430}, // idx == 469
	{214, 8, 470, 470, 432}, // idx == 470
	{215, 8, 471, 471, 433}, // idx == 471
	{216, 8, 472, 472, 437}, // idx == 472
	{217, 8, 473, 473, 438}, // idx == 473
	{218, 8, 474, 474, 440}, // idx == 474
	{219, 8, 475, 475, 441}, // idx == 475
	{220, 8, 476, 476, 444}, // idx == 476
	{221, 8, 477, 477, 445}, // idx == 477
	{222, 8, 478, 478, 447}, // idx == 478
	{223, 8, 479, 479, 448}, // idx == 479
	{224, 8, 480, 480, 454}, // idx == 480
	{225, 8, 481, 481, 455}, // idx == 481
	{226, 8, 482, 482, 457}, // idx == 482
	{227, 8, 483, 483, 458}, // idx == 483
	{228, 8, 484, 484, 461}, // idx == 484
	{229, 8, 485, 485, 462}, // idx == 485
	{230, 8, 486, 486, 464}, // idx == 486
	{231, 8, 487, 487, 465}, // idx == 487
	{232, 8, 488, 488, 469}, // idx == 488
	{233, 8, 489, 489, 470}, // idx == 489
	{234, 8, 490, 490, 472}, // idx == 490
	{235, 8, 491, 491, 473}, // idx == 491
	{236, 8, 492, 492, 476}, // idx == 492
	{237, 8, 493, 493, 477}, // idx == 493
	{238, 8, 494, 494, 479}, // idx == 494
	{239, 8, 495, 495, 480}, // idx == 495
	{240, 8, 496, 496, 485}, // idx == 496
	{241, 8, 497, 497, 486}, // idx == 497
	{242, 8, 498, 498, 488}, // idx == 498
	{243, 8, 499, 499, 489}, // idx == 499
	{244, 8, 500, 500, 492}, // idx == 500
	{245, 8, 501, 501, 493}, // idx == 501
	{246, 8, 502, 502, 495}, // idx == 502
	{247, 8, 503, 503, 496}, // idx == 503
	{248, 8, 504, 504, 500}, // idx == 504
	{249, 8, 505, 505, 501}, // idx == 505
	{250, 8, 506, 506, 503}, // idx == 506
	{251, 8, 507, 507, 504}, // idx == 507
	{252, 8, 508, 508, 507}, // idx == 508
	{253, 8, 509, 509, 508}, // idx == 509
	{254, 8, 510, 510, 510}, // idx == 510
	{255, 8, 511, 511, 511}, // idx == 511
}
