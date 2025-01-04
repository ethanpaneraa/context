package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FilePicker struct {
    app       *tview.Application
    list      *tview.List
    search    *tview.InputField
    files     []FileEntry
    filtered  []FileEntry
    selected  []FileEntry
    onDone    func([]FileEntry)
}

func NewFilePicker(files []FileEntry, onDone func([]FileEntry)) *FilePicker {
    picker := &FilePicker{
        app:      tview.NewApplication(),
        files:    files,
        filtered: files,
        onDone:   onDone,
    }

    picker.setupUI()
    return picker
}

func (fp *FilePicker) setupUI() {
    // Create main flex container
    flex := tview.NewFlex().SetDirection(tview.FlexRow)

    // Setup help text with more visible formatting
    helpText := tview.NewTextView().
        SetText("Controls:\n" +
            "↑/↓      : Navigate files\n" +
            "Space    : Select/deselect file\n" +
            "Tab      : Switch between search and list\n" +
            "/        : Focus search\n" +
            "Enter    : Confirm selection\n" +
            "Esc      : Cancel").
        SetTextAlign(tview.AlignLeft).
        SetDynamicColors(true)

    // Setup search input
    fp.search = tview.NewInputField().
        SetLabel("Search: ").
        SetFieldWidth(0).
        SetChangedFunc(fp.onSearch).
        SetDoneFunc(func(key tcell.Key) {
            if key == tcell.KeyEnter {
                fp.app.SetFocus(fp.list)
            }
        })

    // Setup file list
    fp.list = tview.NewList().
        ShowSecondaryText(true).
        SetHighlightFullLine(true).
        SetSelectedBackgroundColor(tcell.ColorRoyalBlue)

    // Populate initial list
    fp.updateList("")

    // Add keybindings for the list
    fp.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyEnter:
            fp.onDone(fp.selected)
            fp.app.Stop()
            return nil
        case tcell.KeyEsc:
			fp.onDone(nil)
            fp.app.Stop()
            return nil
        case tcell.KeyRune:
            if event.Rune() == '/' {
                fp.app.SetFocus(fp.search)
                return nil
            }
			if event.Rune() == ' ' {
				if idx := fp.list.GetCurrentItem(); idx >= 0 && idx < len(fp.filtered) {
					fp.toggleSelection(fp.filtered[idx])
					return nil
				}
	 		}
	}
        return event
})

    // Add global keybindings
    fp.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
		case tcell.KeyCtrlC:
			fp.onDone(nil)
			fp.app.Stop()
			return nil
        case tcell.KeyTab:
            if fp.app.GetFocus() == fp.search {
                fp.app.SetFocus(fp.list)
            } else {
                fp.app.SetFocus(fp.search)
            }
            return nil
        }
        return event
    })

    // Add components to flex with better proportions
    flex.AddItem(helpText, 7, 0, false).          // Give help text more space
        AddItem(fp.search, 1, 0, false).
        AddItem(fp.list, 0, 1, true)

    // Set root and initial focus
    fp.app.SetRoot(flex, true).SetFocus(fp.list)
}

func (fp *FilePicker) toggleSelection(file FileEntry) {
    // Check if already selected
    for i, sel := range fp.selected {
        if sel.Path == file.Path {
            // Remove from selection
            fp.selected = append(fp.selected[:i], fp.selected[i+1:]...)
            fp.updateList(fp.search.GetText())
            return
        }
    }

    fp.selected = append(fp.selected, file)
    fp.updateList(fp.search.GetText())
}

func (fp *FilePicker) onSearch(text string) {
    fp.updateList(text)
}

func (fp *FilePicker) updateList(search string) {
    fp.list.Clear()
    fp.filtered = []FileEntry{}

    search = strings.ToLower(search)
    for _, file := range fp.files {
        if search == "" || strings.Contains(strings.ToLower(file.Path), search) {
            // Check if file is selected
            isSelected := false
            for _, sel := range fp.selected {
                if sel.Path == file.Path {
                    isSelected = true
                    break
                }
            }

            // Add to filtered list
            fp.filtered = append(fp.filtered, file)

            // Create display text
            prefix := "  "
            if isSelected {
                prefix = "✓ "
            }

            fp.list.AddItem(
                prefix+file.Path,
                formatFileInfo(file),
                0,
                nil,
            )
        }
    }
}

func formatFileInfo(file FileEntry) string {
    size := formatSize(file.Size)
    ext := filepath.Ext(file.Path)
    if ext == "" {
        ext = "no extension"
    }
    
    info := fmt.Sprintf("%s, %s", size, ext)
    
    if file.TokenCount != nil {
        tokenInfo := fmt.Sprintf(", %d tokens", file.TokenCount.Count)
        if file.TokenCount.TokensPerc >= 80 {
            tokenInfo += fmt.Sprintf(" (%.1f%% of limit!)", file.TokenCount.TokensPerc)
        }
        info += tokenInfo
    }
    
    return info
}

func formatSize(size int64) string {
    const unit = 1024
    if size < unit {
        return fmt.Sprintf("%d B", size)
    }
    div, exp := int64(unit), 0
    for n := size / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func (fp *FilePicker) Run() error {
    return fp.app.Run()
}