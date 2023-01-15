package chess

import "math/bits"

func (b *Board) MoveGen() []Move {
	var moves []Move

	us, them := b.Black, b.White
	usSide := Black
	if b.WhiteToMove {
		us, them = b.White, b.Black
		usSide = White
	}

	var all = us | them

	target := ^us
	if b.Checkers != 0 {
		kingSq := bits.TrailingZeros64(b.Kings & us)
		firstChecker := bits.TrailingZeros64(b.Checkers)
		target = b.Checkers | betweenMask[firstChecker][kingSq]
	}

	// Re-use variables to avoid needless allocations.
	var (
		fromBB, toBB uint64
		from, to     int
	)

	// En passant, holy hell.
	if b.EnPassantSq != 0 {
		for fromBB = PawnAttacks(usSide, b.EnPassantSq); fromBB != 0; fromBB &= fromBB - 1 {
			from := bits.TrailingZeros64(fromBB)
			moves = append(moves, newMove(from, b.EnPassantSq, Pawn, Pawn))
		}
	}

	// Knight moves.
	for fromBB = b.Knights & us; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = knightAttacks[from] & target; toBB != 0; toBB &= toBB - 1 {
			moves = append(moves, newMove(from, to, Knight, b.PieceType(from)))
		}
	}

	// Bishop moves.
	for fromBB = b.Bishops & us; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = BishopAttacks(from, all) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			moves = append(moves, newMove(from, to, Bishop, b.PieceType(to)))
		}
	}

	// Rook moves.
	for fromBB = b.Rooks & us; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = RookAttacks(from, all) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			moves = append(moves, newMove(from, to, Rook, b.PieceType(to)))
		}
	}

	// Queen moves.
	for fromBB = b.Queens & us; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = QueenAttacks(from, all) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			moves = append(moves, newMove(from, to, Queen, b.PieceType(to)))
		}
	}

	// King moves.
	from = FirstOne(b.Kings & us)
	for toBB = KingAttacks[from] & target; toBB != 0; toBB &= toBB - 1 {
		to = FirstOne(toBB)
		moves = append(moves, newMove(from, to, King, b.PieceType(to)))
	}

	return moves
}

func Perft(b *Board, depth int) int {
	var count int
	moves := b.MoveGen()
	for _, mv := range moves {
		b.Move(move)
		if depth > 1 {
			count += Perft(b, depth-1)
		} else {
			count++
		}
	}
	return count
}
