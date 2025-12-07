package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func List(file *os.File, db DB) {
	app := tview.NewApplication()
	mainView := tview.NewTreeNode("Welcome to Notabena!").SetColor(tcell.ColorMediumPurple).SetSelectable(false)
	noteTree := tview.NewTreeView().SetRoot(mainView).SetCurrentNode(mainView)

	for _, v := range db.GetNotes() {
		stringId := strconv.FormatUint(uint64(v.Id), 10)
		node := tview.NewTreeNode(v.Name + " [grey]#" + stringId + "[white]")
		node.SetReference(v.Id).SetExpanded(false)
		mainView.AddChild(node)
		node.AddChild(
			tview.NewTreeNode("Edit").SetReference("EDT+" + stringId).SetColor(tcell.ColorLightCyan),
		)
		node.AddChild(
			tview.NewTreeNode("View").SetReference("VWR+" + stringId),
		)
		node.AddChild(
			tview.NewTreeNode("Delete").SetReference("DEL+" + stringId).SetColor(tcell.ColorRed),
		)
	}

	mainView.AddChild(
		tview.NewTreeNode("Create a Note!").SetReference("NEW").SetColor(tcell.ColorBlue),
	)
	mainView.AddChild(
		tview.NewTreeNode("Exit Notabena!").SetReference("EXT").SetColor(tcell.ColorRed),
	)

	noteTree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == "NEW" {
			app.Stop()
			Create(file, db, 0)
			return
		}
		if reference == "EXT" {
			app.Stop()
			return
		}
		str, ok := reference.(string)
		if ok {
			num, err := strconv.ParseUint(strings.Split(str, "+")[1], 10, 32)
			if err != nil {
				panic(err)
			}
			app.Stop()
			switch strings.Split(str, "+")[0] {
			case "DEL":
				db.DeleteNote(uint32(num))
				List(file, db)
			case "EDT":
				Create(file, db, uint32(num))
			case "VWR":
				View(file, db, uint32(num))
			}
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
	})

	if err := app.SetRoot(noteTree, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		log.Fatalf("Error while starting Notabena: %s", err)
	}
}
