package chess

import "math/bits"

const (
	White = 0
	Black = 1
)

const (
	FileA int = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

const (
	Rank1 int = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

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
	FileMask = [8]uint64{
		FileAMask, FileBMask, FileCMask, FileDMask,
		FileEMask, FileFMask, FileGMask, FileHMask,
	}
	RankMask = [8]uint64{
		Rank1Mask, Rank2Mask, Rank3Mask, Rank4Mask,
		Rank5Mask, Rank6Mask, Rank7Mask, Rank8Mask,
	}
)

func Flip(sq int) int { return sq ^ 56 }
func File(sq int) int { return sq & 7 }
func Rank(sq int) int { return sq >> 3 }

func rankFileToSquare(r int, f int) int {
	return 8*r + f
}

func Up(b uint64) uint64    { return b << 8 }
func Down(b uint64) uint64  { return b >> 8 }
func Right(b uint64) uint64 { return (b & ^FileHMask) << 1 }
func Left(b uint64) uint64  { return (b & ^FileAMask) >> 1 }

func UpRight(b uint64) uint64   { return Up(Right(b)) }
func UpLeft(b uint64) uint64    { return Up(Left(b)) }
func DownRight(b uint64) uint64 { return Down(Right(b)) }
func DownLeft(b uint64) uint64  { return Down(Left(b)) }

func fileDistance(sq1, sq2 int) int {
	return absDelta(File(sq1), File(sq2))
}

func rankDistance(sq1, sq2 int) int {
	return absDelta(Rank(sq1), Rank(sq2))
}

func SquareDistance(sq1, sq2 int) int {
	return max(fileDistance(sq1, sq2), rankDistance(sq1, sq2))
}

func FirstOne(x uint64) int {
	return bits.TrailingZeros64(x)
}

func absDelta(x, y int) int {
	if x > y {
		return x - y
	}
	return y - x
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
