package main

import (
	"fmt"

	"github.com/noahklein/chess/chess"
)

func main() {
	fmt.Println(chess.PrintBitboard(^chess.FileAMask))
	fmt.Println(chess.PrintBitboard(chess.FileBMask))
	fmt.Println(chess.PrintBitboard(chess.FileCMask))
}
