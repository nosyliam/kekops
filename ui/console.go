package ui

import (
	"fmt"
	tb "github.com/gdamore/tcell/termbox"
	"github.com/google/gxui/math"
	ui "kekops/termui/v3"
	"kekops/termui/v3/widgets"
	"strings"
	"time"
)

type LogType int
type Category int

const (
	LOG_DEBUG LogType = iota
	LOG_INFO
	LOG_WARN
	LOG_ERROR
	LOG_CONSOLE
	LOG_NOLEVEL
	LOG_TRACE
)

const (
	CATEGORY_CONSOLE Category = iota
	CATEGORY_MAIN
	CATEGORY_ROBLOX
	CATEGORY_TRACKER
	CATEGORY_OPERATIONS
	CATEGORY_NONE
)

var logColors = map[LogType]string{
	LOG_DEBUG:   "blue",
	LOG_INFO:    "white",
	LOG_WARN:    "yellow",
	LOG_ERROR:   "red",
	LOG_CONSOLE: "magenta",
	LOG_TRACE:   "blue",
}

var ExitChan chan bool
var inputDisabled bool = false
var TraceEnabled bool

type Log struct {
	category  string
	level     string
	levelType LogType
	data      string
	length    int
}

func (l *Log) String() string {
	return fmt.Sprintf("[:%s](fg:green)[%s](fg:%s) %s\n", l.category, l.level, logColors[l.levelType], l.data)
}

func (l *Log) Update(data string) {
	l.data = data
	Console.Render()
}

type KekConsole struct {
	KekUi
	widgets.Paragraph

	Logs     []*Log
	commands map[string]func(...string)

	commandHistory []string
	scrollDelta    int
	historyDelta   int
	rendering      bool
}

func NewConsole() *KekConsole {
	console := &KekConsole{Paragraph: *widgets.NewParagraph(), commands: make(map[string]func(...string)), historyDelta: -1}
	console.Log(CATEGORY_NONE, LOG_NOLEVEL, "KEKOPS 1.0 by [nosyliam](mod:bold)\n")
	return console
}

func (k *KekConsole) Render() {
	defer func() {
		if r := recover(); r != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Caught panic!")
		}
	}()

	if k.rendering {
		time.Sleep(10 * time.Millisecond)
		k.Render()
		return
	}
	k.rendering = true
	var renderedLogs []string
	var lineCount int
	k.Text = ""

	for n := 0 + k.scrollDelta; n < 28+k.scrollDelta; n++ {
		if n > len(k.Logs)-1 {
			break
		}
		log := k.Logs[n]
		str := log.String()
		if log.length == -1 {
			cells := ui.ParseStyles(str, k.TextStyle)
			cells = ui.WrapCells(cells, uint(k.Inner.Dx()))
			rows := ui.SplitCells(cells, '\n')
			log.length = len(rows)
		}
		lineCount += log.length
		renderedLogs = append(renderedLogs, str)
		if lineCount > 28 {
			break
		}
	}
	for n := len(renderedLogs) - 1; n >= 0; n-- {
		k.Text += renderedLogs[n]
	}
	ui.Render(k)
	k.rendering = false
}

func (k *KekConsole) ProcessCommand(input string) {
	k.Log(CATEGORY_CONSOLE, LOG_CONSOLE, input)
	k.commandHistory = append([]string{input}, k.commandHistory...)
	args := strings.Split(input, " ")
	if len(args) < 1 {
		goto err
	}
	if fn, ok := k.commands[args[0]]; ok {
		fn(args[1:]...)
		return
	} else {
		goto err
	}
err:
	k.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid command.\n")
}

func (k *KekConsole) RegisterCommand(cmd string, fn func(...string)) {
	k.commands[cmd] = fn
}

func (k *KekConsole) Log(category Category, level LogType, str string) *Log {
	var categoryString, levelString string
	switch category {
	case CATEGORY_CONSOLE:
		categoryString = "Console"
	case CATEGORY_MAIN:
		categoryString = "Main"
	case CATEGORY_ROBLOX:
		categoryString = "ROBLOX"
	case CATEGORY_TRACKER:
		categoryString = "Tracker"
	case CATEGORY_OPERATIONS:
		categoryString = "Operations"
	}
	switch level {
	case LOG_DEBUG:
		levelString = " [Debug]"
	case LOG_INFO:
		levelString = " [Info]"
	case LOG_WARN:
		levelString = " [Warning]"
	case LOG_ERROR:
		levelString = " [Error]"
	case LOG_TRACE:
		levelString = " [Trace]"
		if !TraceEnabled {
			return nil
		}
	}
	log := &Log{categoryString, levelString, level, str, -1}
	k.Logs = append([]*Log{log}, k.Logs...)
	if len(k.Logs) == 201 {
		k.Logs[200] = &Log{}
		k.Logs = k.Logs[0:199]
	}
	go k.Render()
	return log
}

func (k *KekConsole) UpdateInput() {
	cells := ui.ParseStyles(Input.Text, Input.TextStyle)
	cells = ui.WrapCells(cells, uint(Input.Inner.Dx()))
	rows := ui.SplitCells(cells, '\n')
	x, y := 1+len(rows[len(rows)-1]), 31+len(rows)
	if x > 48 {
		x = 1
		y = 33
	}
	tb.SetCursor(x, y)
	ui.Render(Input)
}

func (k *KekConsole) HandleEvent(e ui.Event) {
	if inputDisabled {
		return
	}
	switch e.ID {
	case "<Up>":
		if len(k.commandHistory) == 0 {
			return
		}
		k.historyDelta = math.Min(len(k.commandHistory)-1, k.historyDelta+1)
		Input.Text = "> " + k.commandHistory[k.historyDelta]
		k.UpdateInput()
	case "<Down>":
		k.historyDelta = math.Max(-1, k.historyDelta-1)
		if k.historyDelta == -1 {
			Input.Text = "> "
		} else {
			Input.Text = "> " + k.commandHistory[k.historyDelta]
		}
		k.UpdateInput()
	case "<MouseWheelUp>":
		k.scrollDelta = math.Max(0, math.Min(len(k.Logs)-20, k.scrollDelta+1))
		k.Render()
	case "<MouseWheelDown>":
		k.scrollDelta = math.Max(0, k.scrollDelta-1)
		k.Render()
	case "<C-c>":
		go func() {
			ExitChan <- true
		}()
		return
	default:
		switch true {
		case len([]byte(e.ID)) == 1:
			Input.Text += string([]byte(e.ID)[0])
		case e.ID == "<Space>":
			Input.Text += " "
		case e.ID == "<Enter>":
			go k.ProcessCommand(Input.Text[2:len(Input.Text)])
			k.historyDelta = -1
			Input.Text = "> "
		case e.ID == "<Backspace>" || e.ID == "<C-<Backspace>>":
			if len(Input.Text) > 2 {
				Input.Text = Input.Text[0 : len(Input.Text)-1]
			}
		}
		k.UpdateInput()
	}
}

func init() {
	ExitChan = make(chan bool)
}
