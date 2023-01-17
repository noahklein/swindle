// Handy bitboards and operations.
package bitboard

import (
	"fmt"
	"strings"

	"github.com/noahklein/dragon"
)

func Flip(sq uint8) uint8 { return sq ^ 56 }
func File(sq uint8) uint8 { return sq & 7 }
func Rank(sq uint8) uint8 { return sq >> 3 }

func RankFileToSquare(r, f int) int { return 8*r + f }

func Up(b uint64) uint64    { return b << 8 }
func Down(b uint64) uint64  { return b >> 8 }
func Right(b uint64) uint64 { return b & ^dragon.FileMasks[7] << 1 }
func Left(b uint64) uint64  { return (b & ^dragon.FileMasks[0]) >> 1 }

func UpRight(b uint64) uint64   { return Up(Right(b)) }
func UpLeft(b uint64) uint64    { return Up(Left(b)) }
func DownRight(b uint64) uint64 { return Down(Right(b)) }
func DownLeft(b uint64) uint64  { return Down(Left(b)) }

func UpFill(gen uint64) uint64 {
	gen |= (gen << 8)
	gen |= (gen << 16)
	gen |= (gen << 32)
	return gen
}

func DownFill(gen uint64) uint64 {
	gen |= (gen >> 8)
	gen |= (gen >> 16)
	gen |= (gen >> 32)
	return gen
}

func String(b uint64) string {
	var str strings.Builder
	str.WriteByte('\n')

	for rank := 7; rank >= 0; rank-- {
		str.WriteString(fmt.Sprintf(" %v  ", rank+1))
		for file := 0; file <= 7; file++ {
			sq := RankFileToSquare(rank, file)
			if b&(1<<sq) >= 1 {
				str.WriteString(" 1 ")
			} else {
				str.WriteString(" 0 ")
			}
		}
		str.WriteByte('\n')
	}

	str.WriteByte('\n')
	str.WriteString("     A  B  C  D  E  F  G  H")

	return str.String()
}
