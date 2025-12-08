package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func View(file *os.File, db DB, id uint32) {
	app := tview.NewApplication()
	textArea := tview.NewTextView().SetWrap(true).SetDynamicColors(true)
	textArea.SetDisabled(false)
	note := db.GetNote(id)
	textArea.SetText(note.Content)
	textArea.SetTitle("Viewing " + note.Name).SetBorder(true)
	info := tview.NewTextView().SetDynamicColors(true).SetText("[::r]^X[::-] Exit")
	position := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)
	pages := tview.NewPages()
	mainView := tview.NewGrid().SetRows(0, 1).AddItem(textArea, 0, 0, 1, 2, 0, 0, true).AddItem(info, 1, 0, 1, 1, 0, 0, false).AddItem(position, 1, 1, 1, 1, 0, 0, false)

	pages.AddAndSwitchToPage("main", mainView, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlX {
			app.Stop()
			List(file, db)
			return nil
		}
		return event
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		log.Fatalf("Error while starting Notabena: %s", err)
	}
}
