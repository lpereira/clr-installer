package clui

import (
	xs "github.com/huandu/xstrings"
	term "github.com/nsf/termbox-go"
)

type TextReader struct {
	BaseControl
	colorized bool
	topLine   int
	lineCount int

	onDrawLine        func(int) string
	onPositionChanged func(int, int)
}

func CreateTextReader(parent Control, width, height int, scale int) *TextReader {
	l := new(TextReader)
	l.BaseControl = NewBaseControl()

	if height == AutoSize {
		height = 10
	}
	if width == AutoSize {
		width = 20
	}

	l.SetSize(width, height)
	l.SetConstraints(width, height)
	l.parent = parent

	l.SetScale(scale)

	if parent != nil {
		parent.AddChild(l)
	}

	l.onDrawLine = nil
	l.onPositionChanged = nil

	return l
}

func (l *TextReader) drawText() {
	if l.onDrawLine == nil {
		return
	}

	PushAttributes()
	defer PopAttributes()

	bg, fg := RealColor(l.bg, ColorEditBack), RealColor(l.fg, ColorEditText)
	if l.Active() {
		bg, fg = RealColor(l.bg, ColorEditActiveBack), RealColor(l.fg, ColorEditActiveText)
	}
	SetTextColor(fg)
	SetBackColor(bg)

	ind := 0
	for ind < l.height {
		var str string
		if ind+l.topLine < l.lineCount {
			str = l.onDrawLine(ind + l.topLine)
		} else {
			if ind+l.topLine == l.lineCount+5 {
				str = xs.Center("--- THE END ---", l.width, " ")
			} else {
				str = ""
			}
		}

		if str != "" {
			str = SliceColorized(str, 0, l.width)
			DrawText(l.x, l.y+ind, str)
		}

		ind++
	}
}

// Repaint draws the control on its View surface
func (l *TextReader) Draw() {
	if l.hidden {
		return
	}

	PushAttributes()
	defer PopAttributes()

	x, y := l.Pos()
	w, h := l.Size()

	bg, fg := RealColor(l.bg, ColorEditBack), RealColor(l.fg, ColorEditText)
	if l.Active() {
		bg, fg = RealColor(l.bg, ColorEditActiveBack), RealColor(l.fg, ColorEditActiveText)
	}

	SetTextColor(fg)
	SetBackColor(bg)
	FillRect(x, y, w, h, ' ')
	l.drawText()
}

func (l *TextReader) home() {
	if l.topLine != 0 {
		l.topLine = 0

		if l.onPositionChanged != nil {
			l.onPositionChanged(l.topLine, l.lineCount)
		}
	}
}

func (l *TextReader) end() {
	if l.lineCount > 0 && l.topLine != l.lineCount-1 {
		l.topLine = l.lineCount - 1

		if l.onPositionChanged != nil {
			l.onPositionChanged(l.topLine, l.lineCount)
		}
	}
}

func (l *TextReader) moveUp(count int) {
	if l.topLine != 0 {
		l.topLine -= count
		if l.topLine < 0 {
			l.topLine = 0
		}

		if l.onPositionChanged != nil {
			l.onPositionChanged(l.topLine, l.lineCount)
		}
	}
}

func (l *TextReader) moveDown(count int) {
	if l.lineCount > 0 && l.topLine != l.lineCount-1 {
		l.topLine += count
		if l.topLine > l.lineCount-1 {
			l.topLine = l.lineCount - 1
		}

		if l.onPositionChanged != nil {
			l.onPositionChanged(l.topLine, l.lineCount)
		}
	}
}

func (l *TextReader) processMouseClick(ev Event) bool {
	if ev.Key != term.MouseLeft {
		return false
	}

	dy := ev.Y - l.y
	ww := l.height

	if dy < l.height/2 {
		l.moveUp(ww - 1)
	} else {
		l.moveDown(ww - 1)
	}

	return true
}

/*
ProcessEvent processes all events come from the control parent. If a control
processes an event it should return true. If the method returns false it means
that the control do not want or cannot process the event and the caller sends
the event to the control parent
*/
func (l *TextReader) ProcessEvent(event Event) bool {
	if !l.Active() || !l.Enabled() {
		return false
	}

	switch event.Type {
	case EventKey:
		switch event.Key {
		case term.KeyHome:
			l.home()
			return true
		case term.KeyEnd:
			l.end()
			return true
		case term.KeyArrowUp:
			l.moveUp(1)
			return true
		case term.KeyArrowDown:
			l.moveDown(1)
			return true
		case term.KeyPgup:
			l.moveUp(l.height - 1)
			return true
		case term.KeyPgdn, term.KeySpace:
			l.moveDown(l.height - 1)
			return true
		}

		switch event.Ch {
		case 'k', 'K':
			l.moveUp(1)
			return true
		case 'j', 'J':
			l.moveDown(1)
			return true
		case 'u', 'U':
			l.moveUp(l.height - 1)
			return true
		case 'd', 'D':
			l.moveDown(l.height - 1)
			return true
		default:
			return false
		}
	case EventMouse:
		return l.processMouseClick(event)
	}

	return false
}

// OnDrawLine is called every time the reader is going to display a line
// the argument of the function is the line number to display
func (l *TextReader) OnDrawLine(fn func(int) string) {
	l.onDrawLine = fn
}

// OnPositionChanged is called every time the reader changes the top line or
// the total number of lines is changed
// Callback gets two numbers: the current top line, and the total number of
// lines. Top line number starts from 0.
func (l *TextReader) OnPositionChanged(fn func(int, int)) {
	l.onPositionChanged = fn
}

func (l *TextReader) LineCount() int {
	return l.lineCount
}

func (l *TextReader) SetLineCount(lineNo int) {
	if l.topLine == lineNo {
		return
	}

	if lineNo < l.topLine-1 {
		l.topLine = lineNo - 1
	}
	l.lineCount = lineNo

	if l.onPositionChanged != nil {
		l.onPositionChanged(l.topLine, l.lineCount)
	}
}

func (l *TextReader) TopLine() int {
	return l.topLine
}

func (l *TextReader) SetTopLine(top int) {
	if top < l.lineCount {
		l.topLine = top

		if l.onPositionChanged != nil {
			l.onPositionChanged(l.topLine, l.lineCount)
		}
	}
}
