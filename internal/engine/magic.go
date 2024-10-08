package engine

import (
	"endtner.dev/nChess/internal/board"
	"math"
	"math/bits"
	"math/rand"
)

type MagicEntry struct {
	Mask   uint64
	Magic  uint64
	Shift  int
	Offset int
}

func Mask(square int, pieceType uint8) uint64 {
	var mask uint64

	offsetIndexStart := 0
	offsetIndexEnd := 8

	if pieceType == board.Rook {
		offsetIndexEnd = 4
	}
	if pieceType == board.Bishop {
		offsetIndexStart = 4
	}

	for i, offset := range DirectionalOffsets[offsetIndexStart:offsetIndexEnd] {
		targetIndex := square + offset

		depth := 1
		for depth < DistanceToEdge[square][i+offsetIndexStart] {
			mask |= 1 << targetIndex

			targetIndex += offset
			depth++
		}
	}

	return mask
}

func ValidMoves(square int, pieceType uint8, occupation uint64) uint64 {
	var moveMask uint64

	offsetIndexStart := 0
	offsetIndexEnd := 8

	if pieceType == board.Rook {
		offsetIndexEnd = 4
	}
	if pieceType == board.Bishop {
		offsetIndexStart = 4
	}

	for i, offset := range DirectionalOffsets[offsetIndexStart:offsetIndexEnd] {
		targetIndex := square + offset

		depth := 1
		for depth <= DistanceToEdge[square][i+offsetIndexStart] {
			moveMask |= 1 << targetIndex

			if (occupation & (1 << targetIndex)) != 0 {
				break
			}

			targetIndex += offset
			depth++
		}
	}

	return moveMask
}

func GenerateOccupancies(mask uint64) []uint64 {
	occupancies := make([]uint64, 0)

	var occ uint64 = 0
	for {
		occupancies = append(occupancies, occ)
		occ = (occ - mask) & mask
		if occ == 0 {
			break
		}
	}

	return occupancies
}

func FindMagic(square int, pieceType uint8) (MagicEntry, int) {
	mask := Mask(square, pieceType)

	occupancies := GenerateOccupancies(mask)
	references := make([]uint64, len(occupancies))

	for i, occ := range occupancies {
		references[i] = ValidMoves(square, pieceType, occ)
	}

	shift := bits.OnesCount64(mask)

	for {
		magic := rand.Uint64() & rand.Uint64() & rand.Uint64()
		if bits.OnesCount64(magic*mask) < 6 {
			continue
		}

		table := make([]uint64, 1<<shift)
		fail := false
		maxIndex := ^int(0)

		for i, occ := range occupancies {
			index := int((occ * magic) >> (64 - shift))
			maxIndex = int(math.Max(float64(index), float64(maxIndex)))

			if table[index] == 0 {
				table[index] = references[i]
				continue
			} else if references[i] != table[index] {
				fail = true
				break
			}
		}

		if !fail {
			return MagicEntry{mask, magic, shift, 0}, maxIndex
		}
	}
}

var RookMagics = []MagicEntry{
	MagicEntry{0x101010101017e, 0x180102480004002, 64 - 12, 0},
	MagicEntry{0x202020202027c, 0xc0200040001001, 64 - 11, 4096},
	MagicEntry{0x404040404047a, 0x4800880a0009000, 64 - 11, 6144},
	MagicEntry{0x8080808080876, 0x100100044210008, 64 - 11, 8192},
	MagicEntry{0x1010101010106e, 0xa0010080a002005, 64 - 11, 10240},
	MagicEntry{0x2020202020205e, 0x4600280514020010, 64 - 11, 12288},
	MagicEntry{0x4040404040403e, 0x880010000800200, 64 - 11, 14336},
	MagicEntry{0x8080808080807e, 0xc100044090230002, 64 - 12, 16384},
	MagicEntry{0x1010101017e00, 0x800140008420, 64 - 11, 20480},
	MagicEntry{0x2020202027c00, 0x800802000904000, 64 - 10, 22528},
	MagicEntry{0x4040404047a00, 0x4001808020001000, 64 - 10, 23552},
	MagicEntry{0x8080808087600, 0x2022002210c20008, 64 - 10, 24576},
	MagicEntry{0x10101010106e00, 0x181000801000c10, 64 - 10, 25600},
	MagicEntry{0x20202020205e00, 0x2001002000d48, 64 - 10, 26624},
	MagicEntry{0x40404040403e00, 0x802100800200, 64 - 10, 27648},
	MagicEntry{0x80808080807e00, 0x9806800046801100, 64 - 11, 28672},
	MagicEntry{0x10101017e0100, 0xa262848001400020, 64 - 11, 30720},
	MagicEntry{0x20202027c0200, 0x1020828020024008, 64 - 10, 32768},
	MagicEntry{0x40404047a0400, 0x6000820020b20040, 64 - 10, 33792},
	MagicEntry{0x8080808760800, 0x40050100100008a1, 64 - 10, 34816},
	MagicEntry{0x101010106e1000, 0x1004808004020800, 64 - 10, 35840},
	MagicEntry{0x202020205e2000, 0x11808006000400, 64 - 10, 36864},
	MagicEntry{0x404040403e4000, 0x2240001104886, 64 - 10, 37888},
	MagicEntry{0x808080807e8000, 0x80020001008044, 64 - 11, 38912},
	MagicEntry{0x101017e010100, 0xd400080008020, 64 - 11, 40960},
	MagicEntry{0x202027c020200, 0x1100400080200880, 64 - 10, 43008},
	MagicEntry{0x404047a040400, 0x500100180200080, 64 - 10, 44032},
	MagicEntry{0x8080876080800, 0x810a10100081000, 64 - 10, 45056},
	MagicEntry{0x1010106e101000, 0x100080080040081, 64 - 10, 46080},
	MagicEntry{0x2020205e202000, 0x14000202001008, 64 - 10, 47104},
	MagicEntry{0x4040403e404000, 0x4a9002100160004, 64 - 10, 48128},
	MagicEntry{0x8080807e808000, 0x8208082000420c9, 64 - 11, 49152},
	MagicEntry{0x1017e01010100, 0x10c401682800020, 64 - 11, 51200},
	MagicEntry{0x2027c02020200, 0x4501401000402004, 64 - 10, 53248},
	MagicEntry{0x4047a04040400, 0x40812000801000, 64 - 10, 54272},
	MagicEntry{0x8087608080800, 0x200100080800800, 64 - 10, 55296},
	MagicEntry{0x10106e10101000, 0x82802400800800, 64 - 10, 56320},
	MagicEntry{0x20205e20202000, 0x1046000802000490, 64 - 10, 57344},
	MagicEntry{0x40403e40404000, 0x1042000412000108, 64 - 10, 58368},
	MagicEntry{0x80807e80808000, 0x40100440a000089, 64 - 11, 59392},
	MagicEntry{0x17e0101010100, 0x8508a0100420020, 64 - 11, 61440},
	MagicEntry{0x27c0202020200, 0x232c2a010004000, 64 - 10, 63488},
	MagicEntry{0x47a0404040400, 0x800a00011050042, 64 - 10, 64512},
	MagicEntry{0x8760808080800, 0x2002004010a20008, 64 - 10, 65536},
	MagicEntry{0x106e1010101000, 0x23000802250010, 64 - 10, 66560},
	MagicEntry{0x205e2020202000, 0x2032008004008002, 64 - 10, 67584},
	MagicEntry{0x403e4040404000, 0x10108108a140021, 64 - 10, 68608},
	MagicEntry{0x807e8080808000, 0x1081044a0004, 64 - 11, 69632},
	MagicEntry{0x7e010101010100, 0x48fffe99fecfaa00, 64 - 10, 71680},
	MagicEntry{0x7c020202020200, 0x48fffe99fecfaa00, 64 - 9, 72704},
	MagicEntry{0x7a040404040400, 0x497fffadff9c2e00, 64 - 9, 73216},
	MagicEntry{0x76080808080800, 0x613fffddffce9200, 64 - 9, 73728},
	MagicEntry{0x6e101010101000, 0xffffffe9ffe7ce00, 64 - 9, 74240},
	MagicEntry{0x5e202020202000, 0xfffffff5fff3e600, 64 - 9, 74752},
	MagicEntry{0x3e404040404000, 0x3ff95e5e6a4c0, 64 - 9, 75264},
	MagicEntry{0x7e808080808000, 0x510ffff5f63c96a0, 64 - 10, 75776},
	MagicEntry{0x7e01010101010100, 0xebffffb9ff9fc526, 64 - 11, 76800},
	MagicEntry{0x7c02020202020200, 0x61fffeddfeedaeae, 64 - 10, 78848},
	MagicEntry{0x7a04040404040400, 0x53bfffedffdeb1a2, 64 - 10, 79872},
	MagicEntry{0x7608080808080800, 0x127fffb9ffdfb5f6, 64 - 10, 80896},
	MagicEntry{0x6e10101010101000, 0x411fffddffdbf4d6, 64 - 10, 81920},
	MagicEntry{0x5e20202020202000, 0x88150004000802c5, 64 - 11, 82944},
	MagicEntry{0x3e40404040404000, 0x3ffef27eebe74, 64 - 10, 84992},
	MagicEntry{0x7e80808080808000, 0x7645fffecbfea79e, 64 - 11, 86016},
}

var RookMoveTable = func() []uint64 {
	var rookMoveTable = make([]uint64, 88064)

	for square := range 64 {
		entry := RookMagics[square]

		occupancies := GenerateOccupancies(entry.Mask)
		for _, occ := range occupancies {
			index := int((occ * entry.Magic) >> entry.Shift)
			rookMoveTable[entry.Offset+index] = ValidMoves(square, board.Rook, occ)
		}
	}

	return rookMoveTable
}()

func GetRookMoves(square int, mask uint64) uint64 {
	entry := RookMagics[square]
	moveIndex := ((mask & entry.Mask) * entry.Magic) >> entry.Shift

	return RookMoveTable[entry.Offset+int(moveIndex)]
}

var BishopMagics = []MagicEntry{
	MagicEntry{0x40201008040200, 0xffedf9fd7cfcffff, 64 - 5, 0},
	MagicEntry{0x402010080400, 0xfc0962854a77f576, 64 - 4, 32},
	MagicEntry{0x4020100a00, 0x20c2022e0020523a, 64 - 5, 48},
	MagicEntry{0x40221400, 0x408204040e00800, 64 - 5, 80},
	MagicEntry{0x2442800, 0x1104104004800, 64 - 5, 112},
	MagicEntry{0x204085000, 0x2202111028040405, 64 - 5, 144},
	MagicEntry{0x20408102000, 0xfc0a66c64a7ef576, 64 - 4, 176},
	MagicEntry{0x2040810204000, 0x7ffdfdfcbd79ffff, 64 - 5, 192},
	MagicEntry{0x20100804020000, 0xfc0846a64a34fff6, 64 - 4, 224},
	MagicEntry{0x40201008040000, 0xfc087a874a3cf7f6, 64 - 4, 240},
	MagicEntry{0x4020100a0000, 0x8001204a4010010, 64 - 5, 256},
	MagicEntry{0x4022140000, 0x20a02000000, 64 - 5, 288},
	MagicEntry{0x244280000, 0x8041831040014100, 64 - 5, 320},
	MagicEntry{0x20408500000, 0x808050148400015, 64 - 5, 352},
	MagicEntry{0x2040810200000, 0xfc0864ae59b4ff76, 64 - 4, 384},
	MagicEntry{0x4081020400000, 0x3c0860af4b35ff76, 64 - 4, 400},
	MagicEntry{0x10080402000200, 0x73c01af56cf4cffb, 64 - 4, 416},
	MagicEntry{0x20100804000400, 0x41a01cfad64aaffc, 64 - 4, 432},
	MagicEntry{0x4020100a000a00, 0x1200808002280, 64 - 7, 448},
	MagicEntry{0x402214001400, 0x409000801450000, 64 - 7, 576},
	MagicEntry{0x24428002800, 0x2124040280a02000, 64 - 7, 704},
	MagicEntry{0x2040850005000, 0x8204800900a02100, 64 - 7, 832},
	MagicEntry{0x4081020002000, 0x7c0c028f5b34ff76, 64 - 4, 960},
	MagicEntry{0x8102040004000, 0xfc0a028e5ab4df76, 64 - 4, 976},
	MagicEntry{0x8040200020400, 0x100c00420c0c10, 64 - 5, 992},
	MagicEntry{0x10080400040800, 0x2010100009420080, 64 - 5, 1024},
	MagicEntry{0x20100a000a1000, 0x40404488020441, 64 - 7, 1056},
	MagicEntry{0x40221400142200, 0x1402006018008020, 64 - 9, 1184},
	MagicEntry{0x2442800284400, 0xaa002016018040, 64 - 9, 1696},
	MagicEntry{0x4085000500800, 0xa030008001008880, 64 - 7, 2208},
	MagicEntry{0x8102000201000, 0x41040641040700, 64 - 5, 2336},
	MagicEntry{0x10204000402000, 0x4a008082a400, 64 - 5, 2368},
	MagicEntry{0x4020002040800, 0x1c100984052001, 64 - 5, 2400},
	MagicEntry{0x8040004081000, 0x46022048020848, 64 - 5, 2432},
	MagicEntry{0x100a000a102000, 0x301080114280040, 64 - 7, 2464},
	MagicEntry{0x22140014224000, 0x80040100100900, 64 - 9, 2592},
	MagicEntry{0x44280028440200, 0x801a008c00820020, 64 - 9, 3104},
	MagicEntry{0x8500050080400, 0x40a0208080010240, 64 - 7, 3616},
	MagicEntry{0x10200020100800, 0x448c040083044811, 64 - 5, 3744},
	MagicEntry{0x20400040201000, 0x4100810040030c00, 64 - 5, 3776},
	MagicEntry{0x2000204081000, 0xdcefd9b54bfcc09f, 64 - 4, 3808},
	MagicEntry{0x4000408102000, 0xf95ffa765afd602b, 64 - 4, 3824},
	MagicEntry{0xa000a10204000, 0x200201c02043000, 64 - 7, 3840},
	MagicEntry{0x14001422400000, 0x1014a04200880801, 64 - 7, 3968},
	MagicEntry{0x28002844020000, 0x8000208209800402, 64 - 7, 4096},
	MagicEntry{0x50005008040200, 0x8301010101000601, 64 - 7, 4224},
	MagicEntry{0x20002010080400, 0x43ff9a5cf4ca0c01, 64 - 4, 4352},
	MagicEntry{0x40004020100800, 0x4bffcd8e7c587601, 64 - 4, 4368},
	MagicEntry{0x20408102000, 0xfc0ff2865334f576, 64 - 4, 4384},
	MagicEntry{0x40810204000, 0xfc0bf6ce5924f576, 64 - 4, 4400},
	MagicEntry{0xa1020400000, 0x8400611041100000, 64 - 5, 4416},
	MagicEntry{0x142240000000, 0x412000484040100, 64 - 5, 4448},
	MagicEntry{0x284402000000, 0x292002042048000, 64 - 5, 4480},
	MagicEntry{0x500804020000, 0x8060200401020600, 64 - 5, 4512},
	MagicEntry{0x201008040200, 0xc3ffb7dc36ca8c89, 64 - 4, 4544},
	MagicEntry{0x402010080400, 0xc3ff8a54f4ca2c89, 64 - 4, 4560},
	MagicEntry{0x2040810204000, 0xfffffcfcfd79edff, 64 - 5, 4576},
	MagicEntry{0x4081020400000, 0xfc0863fccb147576, 64 - 4, 4608},
	MagicEntry{0xa102040000000, 0x200248830841020, 64 - 5, 4624},
	MagicEntry{0x14224000000000, 0x2010a0200840410, 64 - 5, 4656},
	MagicEntry{0x28440200000000, 0x6000008010221200, 64 - 5, 4688},
	MagicEntry{0x50080402000000, 0x1204842002020200, 64 - 5, 4720},
	MagicEntry{0x20100804020000, 0xfc087e8e4bb2f736, 64 - 4, 4752},
	MagicEntry{0x40201008040200, 0x43ff9e4ef4ca2c89, 64 - 5, 4768},
}

var BishopMoveTable = func() []uint64 {
	var bishopMoveTable = make([]uint64, 4800)

	for square := range 64 {
		entry := BishopMagics[square]

		occupancies := GenerateOccupancies(entry.Mask)
		for _, occ := range occupancies {
			index := int((occ * entry.Magic) >> entry.Shift)
			bishopMoveTable[entry.Offset+index] = ValidMoves(square, board.Bishop, occ)
		}
	}

	return bishopMoveTable
}()

func GetBishopMoves(square int, mask uint64) uint64 {
	entry := BishopMagics[square]
	moveIndex := ((mask & entry.Mask) * entry.Magic) >> entry.Shift

	return BishopMoveTable[entry.Offset+int(moveIndex)]
}

func GetSliderMoves(square int, mask uint64, orthogonal bool) uint64 {
	if orthogonal {
		return GetRookMoves(square, mask)
	}
	return GetBishopMoves(square, mask)
}
