package bitboard

import "fmt"

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
	files = [8]uint64{
		FileAMask, FileBMask, FileCMask, FileDMask,
		FileEMask, FileFMask, FileGMask, FileHMask,
	}
	ranks = [8]uint64{
		Rank1Mask, Rank2Mask, Rank3Mask, Rank4Mask,
		Rank5Mask, Rank6Mask, Rank7Mask, Rank8Mask,
	}
)

var (
	FileMask [64]uint64
	RankMask [64]uint64

	AdjacentMask [64]uint64
	// [White, Black]
	PassedMask [2][64]uint64
)

func init() {
	for sq := uint8(0); sq < 64; sq++ {
		rank, file := Rank(sq), File(sq)
		FileMask[sq] = files[file]
		RankMask[sq] = ranks[rank]

		AdjacentMask[sq] = Left(files[file]) | Right(files[file])

		up := UpFill(1<<sq) & ^ranks[rank]
		PassedMask[0][sq] = Left(up) | up | Right(up)

		down := DownFill(1<<sq) & ^ranks[rank]
		PassedMask[1][sq] = Left(down) | down | Right(down)
	}

	// diagnose(42)
}

// For debugging.
func diagnose(sq int) {
	fmt.Println("File", String(FileMask[sq]))
	fmt.Println("Rank", String(RankMask[sq]))
	fmt.Println()

	fmt.Println("Passed", String(PassedMask[0][sq]))
	fmt.Println("Passed", String(PassedMask[1][sq]))
	fmt.Println("Adjacent", String(AdjacentMask[sq]))
}
