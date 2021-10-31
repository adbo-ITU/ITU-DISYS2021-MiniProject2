package main

import (
	"disysminiproject2/service"
	"log"
	"strings"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type UserUI struct {
	grid          *ui.Grid
	chatInput     *widgets.Paragraph
	chatPane      *widgets.List
	userInput     string
	uiEvents      <-chan ui.Event
	chatEvents    chan *service.UserMessage
	messageStream chan string
	renderArbiter sync.Mutex
}

func (u *UserUI) HandleChatMessages() {
	log.Println("Starting to listen for chat messages from the server")
	for {
		msg := <-u.chatEvents
		formattedMsg := FormatMessageContent(msg)
		lines := u.manuallyWrapLines(formattedMsg)
		for _, line := range lines {
			u.chatPane.Rows = append(u.chatPane.Rows, line)
			u.chatPane.ScrollDown()
		}
		u.chatPane.Rows = append(u.chatPane.Rows, "")
		u.chatPane.ScrollDown()
		log.Println(formattedMsg)
		u.Render()
	}
}

func (u *UserUI) HandleUIEvents(systemExitChan chan<- bool) {
	for {
		e := <-u.uiEvents
		switch e.ID {
		case "<Enter>":
			log.Println("Received UI event for <Enter>")
			// send a message
			if len(u.userInput) > 0 {
				buffer := u.userInput
				// clear the user input box
				u.userInput = ""
				u.chatInput.Text = ""
				u.Render()
				// send buffered message
				u.messageStream <- buffer
			}
		case "<Up>":
			u.handleScrollUp()
		case "<Down>":
			u.handleScrollDown()
		case "<Escape>", "<C-c>":
			log.Println("Received UI event for program exit")
			systemExitChan <- true
		case "<Backspace>":
			log.Println("Received UI event for <Backspace>")
			length := len(u.userInput)
			if length > 0 {
				u.userInput = u.userInput[:length-1]
			}
		case "<Space>":
			u.userInput += " "
		default:
			if IsLegalCharacter(e.ID) {
				u.userInput += e.ID
			}
		}

		u.Render()
	}
}

func (u *UserUI) Render() {
	u.renderArbiter.Lock()
	defer u.renderArbiter.Unlock()

	// set the text of the chatInput
	u.chatInput.Text = u.userInput

	// Here we manually do text wrapping on the input to fit it in the text box.
	// We also only show the last N lines that can fit in the text box.
	inputLines := u.manuallyWrapLines(u.userInput)
	maxHeightForInput := u.chatInput.Inner.Size().Y
	if len(inputLines) > maxHeightForInput {
		inputLines = inputLines[len(inputLines)-maxHeightForInput:]
	}

	u.chatInput.Text = strings.Join(inputLines, "\n")

	ui.Render(u.grid)
}

func NewUI() UserUI {
	width, height := ui.TerminalDimensions()

	// Create the boxes within the window
	chatInput := widgets.NewParagraph()
	chatInput.BorderStyle.Fg = ui.ColorBlue

	chatPane := widgets.NewList()
	chatPane.BorderStyle.Fg = ui.ColorMagenta
	chatPane.WrapText = true

	// Create the main grid and insert all the widgets to form the UI
	grid := ui.NewGrid()
	grid.SetRect(0, 0, width, height)
	grid.Set(
		ui.NewRow(0.8, chatPane),
		ui.NewRow(0.2, chatInput),
	)

	messagesChannel := make(chan (*service.UserMessage), 1000)
	messageStream := make(chan (string))

	return UserUI{grid: grid, chatInput: chatInput, chatPane: chatPane, chatEvents: messagesChannel, messageStream: messageStream}
}

func (u *UserUI) manuallyWrapLines(text string) []string {
	maxLengthForInput := u.chatInput.Inner.Size().X
	lines := strings.Split(text, "\n")
	outLines := make([]string, 0)
	for _, line := range lines {
		subLines := bigChungus([]rune(line), maxLengthForInput)
		for i := 0; i < len(subLines); i++ {
			outLines = append(outLines, string(subLines[i]))
		}
	}
	return outLines
}

// Thank you, https://freshman.tech/snippets/go/split-slice-into-chunks/
// Splits a slice into uniformly sized chunks
func bigChungus(slice []rune, chunkSize int) [][]rune {
	var chunks [][]rune
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func (u *UserUI) handleScrollUp() {
	// The scroll amounts have been deduced from experiments, as it seems that
	// the amount '1' doesn't correlate to a single line
	u.chatPane.ScrollAmount(-30)
}

func (u *UserUI) handleScrollDown() {
	u.chatPane.ScrollAmount(30)
}