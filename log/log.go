package log

import (
	"fmt"
	"log"
)

type Level int

const (
	NONE Level = iota - 1
	UCI
	ERROR
	WARN
)

const (
	escape = "\x1b"
	reset  = escape + "[m"
	yellow = escape + "[33m"
	red    = escape + "[31m"
	green  = escape + "[32m"
)

type Logger struct {
	Level Level
}

// UCI communicates with the GUI.
func (l Logger) UCI(s string, a ...any) {
	if l.Level >= UCI {
		fmt.Printf(s+"\n", a...)
	}
}

// Error logging.
func (l Logger) Error(s string, a ...any) {
	if l.Level >= ERROR {
		fmt.Println(color(red, s, a...))
	}
}

// Warn logging.
func (l Logger) Warn(s string, a ...any) {
	if l.Level >= WARN {
		fmt.Println(color(yellow, s, a...))
	}
}

// Fatal error, log and exit.
func (l Logger) Fatal(s string, a ...any) {
	log.Fatalf(s, a...)
}

func color(clr string, s string, a ...any) string {
	return fmt.Sprintf(clr+s+reset, a...)
}

func Green(s string, a ...any) {
	fmt.Println(color(green, s, a...))
}

func Yellow(s string, a ...any) {
	fmt.Println(color(yellow, s, a...))
}

func Red(s string, a ...any) {
	fmt.Println(color(red, s, a...))
}
