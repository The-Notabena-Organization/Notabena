package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rivo/tview"
)

func Create(file *os.File, db DB, edit uint32) {
	app := tview.NewApplication()
	textArea := tview.NewTextArea().SetWrap(true).SetPlaceholder("Write all your thoughts here! :D")
	var prevNote Note
	var noteNum uint32
	if edit != 0 {
		prevNote = db.GetNote(edit)
		noteNum = uint32(prevNote.Id)
		textArea.SetText(prevNote.Content, false)
		textArea.SetTitle(fmt.Sprintf("%s [gray]#%d", prevNote.Name, noteNum)).SetBorder(true)
	} else {
		prevNote = Note{
			Id: 0,
		}
		noteNum = uint32(len(db.GetNotes()) + 1)
		textArea.SetTitle(fmt.Sprintf("New note [gray]#%d", noteNum)).SetBorder(true)
	}

	info := tview.NewTextView().SetDynamicColors(true).SetText("[::r]^X[::-] Save [::r]^Q[::-] Quit [::r]F1[::-] Help")
	position := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)
	pages := tview.NewPages()
	updateInfo := func() {
		fromRow, fromColumn, toRow, toColumn := textArea.GetCursor()
		if fromRow == toRow && fromColumn == toColumn {
			position.SetText(
				fmt.Sprintf("[::r]%d:%d[::-] ", fromRow+1, fromColumn+1),
			)
		} else {
			position.SetText(
				fmt.Sprintf("[red]From[white] [::r]%d:%d[::-] [red]To[white] [::r]%d:%d[::-] ", fromRow+1, fromColumn+1, toRow+1, toColumn+1),
			)
		}
	}

	textArea.SetMovedFunc(updateInfo)
	updateInfo()
	mainView := tview.NewGrid().SetRows(0, 1).AddItem(textArea, 0, 0, 1, 2, 0, 0, true).AddItem(info, 1, 0, 1, 1, 0, 0, false).AddItem(position, 1, 1, 1, 1, 0, 0, false)
	mdpage1 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[white:#294cff]Notabena! supports simple markdown.[white:-]
Syntax: [green][<foreground>:<background>:<attributes>:<url>][white]
[green]<foreground>[white]: Change the color of the words.
[green]<background>[white]: Change the color of the background behind the words.
[green]<attributes>[white]: Alter the text. (italic, bold, etc)
[green]<url>[white]: Create a masked link with the text.

List of attributes (uppercase the letter to negate!):
- [yellow]b[white]: Bold
- [yellow]i[white]: Italic
- [yellow]l[white]: Blink
- [yellow]d[white]: Dim
- [yellow]r[white]: Reverse (switches foreground and background colors)
- [yellow]u[white]: Underline
- [yellow]s[white]: Strikethrough

[::r]Enter[::-] Proceed [::r]Escape[::-] Return`)
	mdpage2 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[white:#294cff]Here are some rules :D[white:-]
[white:green]1. You can skip parts of the syntax.[white:-]
You can write things like [green::b][::b[][white::-] to skip the previous parts of the syntax while keeping your notes short and stylish.

[white:green]2. Always type closing brackets after you're done.
Else you'll have this issue.[white:-] If you write something like [green:white:b[] and not close it with [white:-:-[] to return to the normal style, the style that you applied for that bit of text will apply for everything else after it.

[white:green]3. - matters.[white:-]
If you don't want to specify what to return to, simply type -.

[::r]Enter[::-] View examples [::r]Escape[::-] Return`)
	mdpage3 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[white:#294cff]Examples[white:-]
[::d]Small thing to mention: the syntax is visible because we're showing how to achieve such styling, when viewing your notes the syntax would not be visible.[::-]

[::i][::i[]Italic and [::I][::I[]not italic
[:purple:bi][:purple:bi[]Purple background, bold and italic![:purple:Bi][:purple:Bi[]oh and now I'm not important :(
[-:-:-:-][-:-:-:-[]At the end of the day, simple is good :D
[yellow][yellow[]Some yellow text, what could possibly go[::r[][::r]wrong...?
WHO CAUSED THIS???? Is this note doomed to live in yellow forever? The quick yellow fox doesn't exist in the first place to jump over the...[-:-:-[][-:-:-]re we go! [::r[][::r]Oh, not again..
[blue][blue[]I'm blue da ba dee da ba[-][-[]WHY D:

[::r]Enter[::-] Back to start [::r]Escape[::-] Return`)

	md := tview.NewFrame(mdpage1).SetBorders(1, 1, 0, 0, 2, 2)
	md.SetBorder(true).
		SetTitle("Markdown").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				pages.SwitchToPage("main")
				return nil
			} else if event.Key() == tcell.KeyEnter {
				switch {
				case md.GetPrimitive() == mdpage1:
					md.SetPrimitive(mdpage2)
				case md.GetPrimitive() == mdpage2:
					md.SetPrimitive(mdpage3)
				case md.GetPrimitive() == mdpage3:
					md.SetPrimitive(mdpage1)
				}
				return nil
			}
			return event
		})

	saved := tview.NewTextView().SetDynamicColors(true).SetText(`[green]Ready to go!
[blue]Please give your note a name.`)

	savedPopup := tview.NewGrid()
	savedPopup.SetBorder(true).SetTitle("Success")
	savedPopup.AddItem(saved, 0, 0, 1, 2, 0, 0, false)
	titleInput := tview.NewTextArea().SetWrap(false).SetPlaceholder("Write your note's title here")
	titleInput.SetTitle("Title your note").SetBorder(true)
	savedPopup.AddItem(titleInput, 1, 0, 5, 3, 0, 0, true)
	savedPopup.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			app.Stop()
			notes := []*Note{}
			sqlscan.Select(context.Background(), db.Db, &notes, "SELECT id FROM saved_notes;")
			note := Note{
				Id:      uint32(len(notes)) + 1,
				Name:    titleInput.GetText(),
				Content: textArea.GetText(),
				Created: time.Now().Format("2006-01-02 15:04")}
			note.Save(file.Name())
			List(file, db)
			return nil
		} else if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})

	pages.AddAndSwitchToPage("main", mainView, true).AddPage("saved", tview.NewGrid().SetColumns(0, 64, 0).SetRows(0, 22, 0).AddItem(savedPopup, 1, 1, 1, 1, 0, 0, true), true, false).AddPage("md", tview.NewGrid().SetColumns(0, 64, 0).SetRows(0, 22, 0).AddItem(md, 1, 1, 1, 1, 0, 0, true), true, false)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			if prevNote.Id != 0 {
				app.Stop()
				notes := []*Note{}
				sqlscan.Select(context.Background(), db.Db, &notes, "SELECT id FROM saved_notes;")
				note := Note{
					Id:      prevNote.Id,
					Name:    prevNote.Name,
					Content: textArea.GetText(),
					Created: prevNote.Created}
				note.Save(file.Name())
				List(file, db)
				return nil
			}
			pages.ShowPage("saved")
		case tcell.KeyCtrlQ:
			app.Stop()
			List(file, db)
		case tcell.KeyF1:
			pages.ShowPage("md")
		}
		return event
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		log.Fatalf("Error while starting Notabena: %s", err)
	}
}
