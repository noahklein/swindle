package bitboard

const (
	FileAMask uint64 = 0x0101010101010101 << iota
	FileBMask
	FileCMask
	FileDMask
	FileEMask
	FileFMask
	FileGMask
	FileHMask
)

const (
	Rank1Mask uint64 = 0xFF << (8 * iota)
	Rank2Mask
	Rank3Mask
	Rank4Mask
	Rank5Mask
	Rank6Mask
	Rank7Mask
	Rank8Mask
)

var (
	ranks = [64]uint64{
		FileAMask, FileBMask, FileCMask, FileDMask,
		FileEMask, FileFMask, FileGMask, FileHMask,
	}
	files = [64]uint64{
		Rank1Mask, Rank2Mask, Rank3Mask, Rank4Mask,
		Rank5Mask, Rank6Mask, Rank7Mask, Rank8Mask,
	}
)

var (
	FileMask [64]uint64
	RankMask [64]uint64

	AdjacentMask [64]uint64
	PassedMask   [64]uint64
)

func init() {
	for sq := 0; sq < 64; sq++ {
		rank, file := Rank(sq), File(sq)
		FileMask[sq] = ranks[file]
		RankMask[sq] = files[rank]

		AdjacentMask[sq] = Left(ranks[file]) | Right(ranks[file])

		up := UpFill(1 << sq)
		PassedMask[sq] = Left(up) | up | Right(up)
	}
}

// For debugging.
// func diagnose(sq int) {
// 	fmt.Println("File", String(FileMask[sq]))
// 	fmt.Println("Rank", String(RankMask[sq]))

// 	fmt.Println("Passed", String(PassedMask[sq]))
// 	fmt.Println("Adjacent", String(AdjacentMask[sq]))
// }
