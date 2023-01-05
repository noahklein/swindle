module github.com/noahklein/chess

go 1.18

require (
	github.com/fatih/color v1.13.0
	github.com/noahklein/dragon v0.0.0-20230102065222-742c8db549d9
	github.com/rodaine/table v1.1.0
)

require (
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
)

replace github.com/noahklein/dragon => ../dragon
