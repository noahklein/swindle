package chess

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const StartPos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

const (
	rankNames = "12345678"
	fileNames = "abcdefgh"
)

func ParseFen(fen string) (Board, error) {
	fields := strings.Fields(fen)
	if len(fields) < 6 {
		return Board{}, fmt.Errorf("invalid fen: not enough fields want 6, got %v, %v", len(fields), fields)
	}
	var (
		pieces    = fields[0]
		side      = fields[1]
		castling  = fields[2]
		enPassant = fields[3]
		halfMove  = fields[4]
		fullMove  = fields[5]
	)

	var pieceColors [64]pieceColor
	var i = 0
	for _, r := range pieces {
		if unicode.IsDigit(r) {
			n, _ := strconv.Atoi(string(r))
			i += n
			continue
		} else if unicode.IsLetter(r) {
			pieceColors[Flip(i)] = pieceFromRune(r)
			i++
		}
	}

	var board Board
	for sq, pc := range pieceColors {
		var m = squareMask[sq]

		if pc.side {
			board.White ^= m
		} else {
			board.Black ^= m
		}

		switch pc.piece {
		case Pawn:
			board.Pawns ^= m
		case Knight:
			board.Knights ^= m
		case Bishop:
			board.Bishops ^= m
		case Rook:
			board.Rooks ^= m
		case Queen:
			board.Queens ^= m
		case King:
			board.Kings ^= m
		}

		// TODO: zobrist hashing.
	}

	// Side to move.
	board.WhiteToMove = side == "w"

	// Castling
	if strings.Contains(castling, "K") {
		board.Castle |= CastleWK
	}
	if strings.Contains(castling, "Q") {
		board.Castle |= CastleWQ
	}
	if strings.Contains(castling, "k") {
		board.Castle |= CastleBK
	}
	if strings.Contains(castling, "q") {
		board.Castle |= CastleBQ
	}

	// En passant square, e.g. e3.
	board.EnPassantSq = SquareNone
	if enPassant != "-" {
		rank := strings.Index(rankNames, enPassant[0:1])
		file := strings.Index(rankNames, enPassant[1:2])
		board.EnPassantSq = rankFileToSquare(rank, file)
	}

	board.HalfMove, _ = strconv.Atoi(halfMove)
	board.Ply, _ = strconv.Atoi(fullMove)
	board.Ply *= 2

	return board, nil
}

type pieceColor struct {
	piece int
	side  bool
}

func pieceFromRune(r rune) pieceColor {
	switch r {
	// White
	case 'P':
		return pieceColor{Pawn, true}
	case 'N':
		return pieceColor{Knight, true}
	case 'B':
		return pieceColor{Bishop, true}
	case 'R':
		return pieceColor{Rook, true}
	case 'Q':
		return pieceColor{Queen, true}
	case 'K':
		return pieceColor{King, true}
	// Black
	case 'p':
		return pieceColor{Pawn, false}
	case 'n':
		return pieceColor{Knight, false}
	case 'b':
		return pieceColor{Bishop, false}
	case 'r':
		return pieceColor{Rook, false}
	case 'q':
		return pieceColor{Queen, false}
	case 'k':
		return pieceColor{King, false}
	}
	return pieceColor{Nothing, false}
}

var pieceStrings = [2][7]string{
	{".", "♙", "♘", "♗", "♖", "♕", "♔"},
	{".", "♟", "♞", "♝", "♜", "♛", "♚"},
}

func (b *Board) String() string {
	var s strings.Builder
	s.WriteByte('\n')

	for rank := Rank8; rank >= Rank1; rank-- {
		s.WriteString(fmt.Sprintf(" %v  ", rank+1))
		for file := FileA; file <= FileH; file++ {
			sq := rankFileToSquare(rank, file)
			piece, isWhite := b.PieceColor(sq)
			if piece < Pawn {
				s.WriteString(" . ")
				continue
			}

			color := 0
			if !isWhite {
				color = 1
			}

			char := pieceStrings[color][piece]
			s.WriteString(" " + char + " ")
		}
		s.WriteByte('\n')
	}

	s.WriteByte('\n')
	s.WriteString("     A  B  C  D  E  F  G  H")
	return s.String()
}

func PrintBitboard(b uint64) string {
	var str strings.Builder
	str.WriteByte('\n')

	for rank := Rank8; rank >= Rank1; rank-- {
		str.WriteString(fmt.Sprintf(" %v  ", rank+1))
		for file := FileA; file <= FileH; file++ {
			sq := rankFileToSquare(rank, file)
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
