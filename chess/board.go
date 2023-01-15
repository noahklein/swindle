package chess

const SquareNone = -1

type Board struct {
	Hash        uint64
	WhiteToMove bool

	// Bitboards.
	White, Black, Checkers  uint64
	Pawns, Knights, Bishops uint64
	Rooks, Queens, Kings    uint64

	Castle int
	// -1 if no en passant available.
	EnPassantSq int

	HalfMove int
	Ply      int

	LastMove Move
}

// PieceType gets the type of piece on a given square.
func (b *Board) PieceType(sq int) int {
	bb := squareMask[sq]
	if (b.White|b.Black)&bb == 0 {
		return Nothing
	}
	if b.Pawns&bb != 0 {
		return Pawn
	}
	if b.Knights&bb != 0 {
		return Knight
	}
	if b.Bishops&bb != 0 {
		return Bishop
	}
	if b.Rooks&bb != 0 {
		return Rook
	}
	if b.Queens&bb != 0 {
		return Queen
	}
	if b.Kings&bb != 0 {
		return King
	}

	// Invalid square.
	return -1
}

func (b *Board) PieceColor(sq int) (int, bool) {
	bb := squareMask[sq]
	return b.PieceType(sq), b.White&bb != 0
}

// Castling rights bits.
const (
	CastleWK = 1 << iota
	CastleWQ
	CastleBK
	CastleBQ
)

// Move is a chess move stored bitwise.
//      Promotion CapturedPiece MovingPiece  From     To
// MSB     000         000           000    000000  000000   LSB
type Move int32

func (m Move) To() int {
	return int(m & 63) // 63 = 111111
}

func (m Move) From() int {
	return int((m >> 6) & 63)
}

func (m Move) MovingPiece() int {
	return int((m >> 12) & 7)
}

func (m Move) CapturedPiece() int {
	return int((m >> 15) & 7)
}

func (m Move) Promotion() int {
	return int((m >> 18) & 7)
}

func (m Move) withPromotion(promotion int) Move {
	return m ^ Move(promotion)<<18
}

func newMove(from, to, moving, captured int) Move {
	return Move(to ^ (from << 6) ^ (moving << 12) ^ (captured << 15))
}

func newPawnMove(from, to, captured, promotion int) Move {
	return Move(to ^ (from << 6) ^ (Pawn << 12) ^ (captured << 15) ^ (promotion << 18))
}

const (
	Nothing int = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)
