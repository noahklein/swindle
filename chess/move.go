package chess

func (b *Board) Move(move Move) {

	var (
		from, to = move.From(), move.To()
		moving   = move.MovingPiece()
		captured = move.CapturedPiece()
	)

}
