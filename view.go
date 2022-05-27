package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"sort"
	"strconv"
)

const (
	Menu   = "MenuView"
	Main   = "MainView"
	Detail = "DetailView"

	MenuWidth         = 32
	MainWidth         = 60
	MainFunctionWidth = MainWidth - 15
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
	"Top Byte (Malloc)",
	"Top Count (Malloc)",
	"Top Byte (MallocAfterFree)",
	"Top Count (MallocAfterFree)",
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
	g.SelFgColor = gocui.ColorMagenta
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

	drawMenuView(g)
	drawMainView(g)
	drawDetailView(g)

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
		return mallocTopByteSlice[i].Byte > mallocTopByteSlice[j].Byte
	})
	sort.SliceStable(mallocTopCountSlice, func(i, j int) bool {
		return mallocTopCountSlice[i].Count > mallocTopCountSlice[j].Count
	})

	remainMallocStatMap := make(map[uint32]*MallocStat)
	for _, v := range remainMallocOpMap {
		if _, ok := remainMallocStatMap[v.StackHash]; ok {
			remainMallocStatMap[v.StackHash].Count += 1
			remainMallocStatMap[v.StackHash].Byte += v.Byte
		} else {
			remainMallocStatMap[v.StackHash] = &MallocStat{
				Byte:  v.Byte,
				Count: 1,
				Stack: v.Stack,
			}
		}
	}
	for _, v := range remainMallocStatMap {
		mallocTopByteAfterFreeSlice = append(mallocTopByteAfterFreeSlice, *v)
		mallocTopCountAfterFreeSlice = append(mallocTopCountAfterFreeSlice, *v)
	}
	sort.SliceStable(mallocTopByteAfterFreeSlice, func(i, j int) bool {
		return mallocTopByteAfterFreeSlice[i].Byte > mallocTopByteAfterFreeSlice[j].Byte
	})
	sort.SliceStable(mallocTopCountAfterFreeSlice, func(i, j int) bool {
		return mallocTopCountAfterFreeSlice[i].Count > mallocTopCountAfterFreeSlice[j].Count
	})
}

func initViews(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	menuView, err := g.SetView(Menu, 0, 0, MenuWidth, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	menuView.Title = "Menu"
	menuView.Highlight = true
	menuView.FgColor = gocui.ColorCyan
	menuView.SelBgColor = gocui.ColorBlue
	menuView.SelFgColor = gocui.ColorBlack

	mainView, err := g.SetView(Main, MenuWidth+1, 0, MenuWidth+MainWidth+1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	mainView.Title = "Main Window"
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
	detailView.Title = "Detail Window"
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
	err = g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, quitReportUI)
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
			str := expandStyleString(elem.Stack[0], MainFunctionWidth, strconv.FormatInt(elem.Byte, 10))
			_, _ = fmt.Fprintf(mainV, "%s\n", str)
		}
	} else if menuSelectIndex == MallocTopCount {
		for _, elem := range mallocTopCountSlice {
			str := expandStyleString(elem.Stack[0], MainFunctionWidth, strconv.FormatInt(int64(elem.Count), 10))
			_, _ = fmt.Fprintf(mainV, "%s\n", str)
		}
	} else if menuSelectIndex == MallocTopByteAfterFree {
		for _, elem := range mallocTopByteAfterFreeSlice {
			str := expandStyleString(elem.Stack[0], MainFunctionWidth, strconv.FormatInt(elem.Byte, 10))
			_, _ = fmt.Fprintf(mainV, "%s\n", str)
		}
	} else if menuSelectIndex == MallocTopCountAfterFree {
		for _, elem := range mallocTopCountAfterFreeSlice {
			str := expandStyleString(elem.Stack[0], MainFunctionWidth, strconv.FormatInt(int64(elem.Count), 10))
			_, _ = fmt.Fprintf(mainV, "%s\n", str)
		}
	}
	_ = mainV.SetCursor(0, mainSelectIndex)
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
		for _, elem := range mainSlice[mainSelectIndex].Stack {
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

func expandStyleString(s1 string, s1width int, s2 string) string {
	var ret string
	if len(s1) >= s1width {
		ret += s1[:s1width-7]
		ret += "..."
		ret += s1[len(s1)-3:]
	} else {
		ret += s1
	}
	for i := len(ret); i < s1width; i++ {
		ret += " "
	}
	ret += s2
	return ret
}
