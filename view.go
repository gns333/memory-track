package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"sort"
)

const (
	Menu   = "MenuView"
	Main   = "MainView"
	Detail = "DetailView"

	MenuWidth = 24
	MainWidth = 48
)

var menuSelectIndex int = 0
var mainSelectIndex int = 0

const (
	MallocTopByte           = 0
	MallocTopCount          = 1
	MallocTopByteAfterFree  = 2
	MallocTopCountAfterFree = 3
)

var MenuDescription = []string{
	"Malloc Top Byte",
	"Malloc Top Count",
	"Malloc Top Byte After Free",
	"Malloc Top Count After Free",
}
var mallocTopByteSlice []MallocStat
var mallocTopCountSlice []MallocStat
var mallocTopByteAfterFreeSlice []MallocStat
var mallocTopCountAfterFreeSlice []MallocStat

func ShowReportUI() error {
	prepareData()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return fmt.Errorf("new cui error: %w", err)
	}
	defer g.Close()

	g.Cursor = false
	g.Highlight = true
	g.SelFgColor = gocui.ColorBlue
	g.BgColor = gocui.ColorBlack
	g.FgColor = gocui.ColorWhite
	g.SetManagerFunc(layout)

	err = initViews(g)
	if err != nil {
		return fmt.Errorf("cui init view error: %w", err)
	}

	err = keyBinding(g)
	if err != nil {
		return fmt.Errorf("cui key binding error: %w", err)
	}
	_, _ = g.SetCurrentView(Menu)
	err = g.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		return fmt.Errorf("cui main loop error: %w", err)
	}
	return nil
}

func prepareData() {
	for _, v := range mallocStatMap {
		mallocTopByteSlice = append(mallocTopByteSlice, *v)
		mallocTopCountSlice = append(mallocTopCountSlice, *v)
	}
	sort.SliceStable(mallocTopByteSlice, func(i, j int) bool {
		return mallocTopByteSlice[i].byte > mallocTopByteSlice[j].byte
	})
	sort.SliceStable(mallocTopCountSlice, func(i, j int) bool {
		return mallocTopCountSlice[i].count > mallocTopCountSlice[j].count
	})

	remainMallocStatMap := make(map[uint32]*MallocStat)
	for _, v := range remainMallocOpMap {
		if _, ok := remainMallocStatMap[v.stackHash]; ok {
			remainMallocStatMap[v.stackHash].count += 1
			remainMallocStatMap[v.stackHash].byte += v.byte
		} else {
			remainMallocStatMap[v.stackHash] = &MallocStat{
				byte:  v.byte,
				count: 1,
				stack: v.stack,
			}
		}
	}
	for _, v := range remainMallocStatMap {
		mallocTopByteAfterFreeSlice = append(mallocTopByteAfterFreeSlice, *v)
		mallocTopCountAfterFreeSlice = append(mallocTopCountAfterFreeSlice, *v)
	}
	sort.SliceStable(mallocTopByteAfterFreeSlice, func(i, j int) bool {
		return mallocTopByteAfterFreeSlice[i].byte > mallocTopByteAfterFreeSlice[j].byte
	})
	sort.SliceStable(mallocTopCountAfterFreeSlice, func(i, j int) bool {
		return mallocTopCountAfterFreeSlice[i].count > mallocTopCountAfterFreeSlice[j].count
	})
}

func initViews(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	menuView, err := g.SetView(Menu, 0, 0, MenuWidth, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	menuView.Highlight = true
	menuView.FgColor = gocui.ColorCyan
	menuView.SelBgColor = gocui.ColorBlue
	menuView.SelFgColor = gocui.ColorBlack

	mainView, err := g.SetView(Main, MenuWidth+1, 0, MenuWidth+MainWidth+1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	mainView.Highlight = true
	mainView.Autoscroll = true
	mainView.Wrap = true
	mainView.FgColor = gocui.ColorCyan
	mainView.SelBgColor = gocui.ColorBlue
	mainView.SelFgColor = gocui.ColorBlack

	detailView, err := g.SetView(Detail, MenuWidth+MainWidth+2, 0, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	detailView.Highlight = true
	detailView.Autoscroll = true
	detailView.Wrap = true
	detailView.FgColor = gocui.ColorCyan
	detailView.SelBgColor = gocui.ColorBlue
	detailView.SelFgColor = gocui.ColorBlack
	return nil
}

func keyBinding(g *gocui.Gui) error {
	err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitReportUI)
	if err != nil {
		return err
	}
	err = g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, keyArrowUp)
	if err != nil {
		return err
	}
	err = g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, keyArrowDown)
	if err != nil {
		return err
	}
	err = g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, keyArrowLeft)
	if err != nil {
		return err
	}
	err = g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, keyArrowRight)
	if err != nil {
		return err
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	_, err := g.SetView(Menu, 0, 0, MenuWidth, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	_, err = g.SetView(Main, MenuWidth+1, 0, MenuWidth+MainWidth+1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	_, err = g.SetView(Detail, MenuWidth+MainWidth+2, 0, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	return nil
}

func quitReportUI(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func drawMenuView(g *gocui.Gui) {
	menuV, _ := g.View(Menu)
	menuV.Clear()
	for i := 0; i < len(MenuDescription); i++ {
		fmt.Fprintln(menuV, MenuDescription[i])
	}
	menuV.SetCursor(0, menuSelectIndex)
}

func drawMainView(g *gocui.Gui) {
	mainV, _ := g.View(Main)
	mainV.Clear()
	if menuSelectIndex == MallocTopByte {
		for _, elem := range mallocTopByteSlice {
			_, _ = fmt.Fprintf(mainV, "%s\t%d\n", elem.stack[0], elem.byte)
		}
	} else if menuSelectIndex == MallocTopCount {
		for _, elem := range mallocTopCountSlice {
			_, _ = fmt.Fprintf(mainV, "%s\t%d\n", elem.stack[0], elem.count)
		}
	} else if menuSelectIndex == MallocTopByteAfterFree {
		for _, elem := range mallocTopByteAfterFreeSlice {
			_, _ = fmt.Fprintf(mainV, "%s\t%d\n", elem.stack[0], elem.byte)
		}
	} else if menuSelectIndex == MallocTopCountAfterFree {
		for _, elem := range mallocTopCountAfterFreeSlice {
			_, _ = fmt.Fprintf(mainV, "%s\t%d\n", elem.stack[0], elem.count)
		}
	}
	mainV.SetCursor(0, mainSelectIndex)
}

func getMainViewSlice() []MallocStat {
	if menuSelectIndex == MallocTopByte {
		return mallocTopByteSlice
	} else if menuSelectIndex == MallocTopCount {
		return mallocTopCountSlice
	} else if menuSelectIndex == MallocTopByteAfterFree {
		return mallocTopByteAfterFreeSlice
	} else if menuSelectIndex == MallocTopCountAfterFree {
		return mallocTopCountAfterFreeSlice
	}
	return nil
}

func drawDetailView(g *gocui.Gui) {
	detailV, _ := g.View(Detail)
	detailV.Clear()
	mainSlice := getMainViewSlice()
	if mainSlice != nil {
		for _, elem := range mainSlice[mainSelectIndex].stack {
			_, _ = fmt.Fprintf(detailV, "%s\n", elem)
		}
	}
}

func keyArrowUp(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == Menu {
		if menuSelectIndex > 0 {
			menuSelectIndex--
			mainSelectIndex = 0
			drawMenuView(g)
			drawMainView(g)
			drawDetailView(g)
		}
	} else if v.Name() == Main {
		if mainSelectIndex > 0 {
			mainSelectIndex--
			drawMainView(g)
			drawDetailView(g)
		}
	}
	return nil
}

func keyArrowDown(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == Menu {
		if menuSelectIndex < len(MenuDescription)-1 {
			menuSelectIndex++
			mainSelectIndex = 0
			drawMenuView(g)
			drawMainView(g)
			drawDetailView(g)
		}
	} else if v.Name() == Main {
		if mainSelectIndex < len(getMainViewSlice())-1 {
			mainSelectIndex++
			drawMainView(g)
			drawDetailView(g)
		}
	}
	return nil
}

func keyArrowLeft(g *gocui.Gui, v *gocui.View) error {
	if g.CurrentView().Name() == Main {
		_, _ = g.SetCurrentView(Menu)
	}
	return nil
}

func keyArrowRight(g *gocui.Gui, v *gocui.View) error {
	if g.CurrentView().Name() == Menu {
		_, _ = g.SetCurrentView(Main)
	}
	return nil
}
