package main

import (
	"context"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/friendsofgo/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/client"
)

var (
	gloveActiveEmoji, _   = fyne.LoadResourceFromPath("./gui/client/glove-active.png")
	gloveInactiveEmoji, _ = fyne.LoadResourceFromPath("./gui/client/glove-inactive.png")
	apiKeyURI             = storage.NewFileURI("punchr-client-api-key.txt")

	sysTrayMenu    *fyne.Menu
	menuItemStatus = fyne.NewMenuItem("", nil)
	menuItemToggle = fyne.NewMenuItem("", nil)
	menuItemApiKey = fyne.NewMenuItem("Set API Key", nil)
)

func main() {
	a := app.New()
	a.SetIcon(gloveInactiveEmoji)

	desk, isDesktopApp := a.(desktop.App)
	if !isDesktopApp {
		log.Errorln("Can only operate as a Desktop application.")
		return
	}

	p := NewPunchr(a)

	menuItemStatus.Disabled = true
	menuItemToggle.Label = "Start Hole Punching"
	menuItemToggle.Action = func() {
		if p.isHolePunching {
			go p.StopHolePunching()
		} else {
			go p.StartHolePunching()
		}
	}
	menuItemApiKey.Action = func() {
		p.ShowApiKeyDialog()
	}

	sysTrayMenu = fyne.NewMenu("Punchr", menuItemStatus, fyne.NewMenuItemSeparator(), menuItemToggle, menuItemApiKey)
	desk.SetSystemTrayMenu(sysTrayMenu)

	if p.apiKey == "" {
		menuItemStatus.Label = "No API-Key"
		menuItemToggle.Disabled = true
	} else {
		menuItemStatus.Label = "API-Key: " + p.apiKey
		menuItemToggle.Label = "Start Hole Punching"
	}

	sysTrayMenu.Refresh()

	a.Run()
}

type Punchr struct {
	app            fyne.App
	hpCtx          context.Context
	hpCtxCancel    context.CancelFunc
	isHolePunching bool
	apiKey         string
}

func NewPunchr(app fyne.App) *Punchr {
	apiKey, err := loadApiKey()
	if err != nil {
		log.WithError(err).Warnln("error loading api key")
	}

	return &Punchr{
		app:            app,
		isHolePunching: false,
		apiKey:         apiKey,
	}
}

func (p *Punchr) ShowApiKeyDialog() {
	window := p.app.NewWindow("Punchr")
	window.Resize(fyne.NewSize(300, 100))
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Please enter your API-Key")
	entry.SetText(p.apiKey)
	btn := widget.NewButton("Save", func() {
		p.SaveApiKey(entry.Text)
		menuItemToggle.Disabled = false
		sysTrayMenu.Refresh()
		window.Close()
	})
	entry.OnChanged = func(s string) {
		if s == "" {
			btn.Disable()
		} else {
			btn.Enable()
		}
	}
	if p.apiKey == "" {
		btn.Disable()
	}
	window.SetContent(container.New(layout.NewVBoxLayout(), entry, btn))
	window.Show()
}

func (p *Punchr) SaveApiKey(apiKey string) {
	rwc, err := storage.Writer(apiKeyURI)
	if err != nil {
		log.WithError(err).Warnln("error opening storage writer")
		return
	}

	if _, err = rwc.Write([]byte(apiKey)); err != nil {
		log.WithError(err).Warnln("error opening storage writer")
		return
	}

	p.apiKey = apiKey
	menuItemStatus.Label = "API-Key: " + apiKey
	sysTrayMenu.Refresh()
}

func (p *Punchr) StartHolePunching() {
	desk := p.app.(desktop.App)

	p.isHolePunching = true
	desk.SetSystemTrayIcon(gloveActiveEmoji)
	menuItemStatus.Label = "Running..."
	menuItemToggle.Label = "Stop Hole Punching"
	sysTrayMenu.Refresh()

	ctx, cancel := context.WithCancel(context.Background())
	p.hpCtx = ctx
	p.hpCtxCancel = cancel

	err := client.App.RunContext(p.hpCtx, []string{"punchrclient", "--api-key", p.apiKey})
	if err != nil && p.hpCtx.Err() != context.Canceled {
		menuItemStatus.Label = "Error: " + err.Error()
	} else {
		menuItemStatus.Label = "API-Key: " + p.apiKey
	}
	p.hpCtx = nil
	p.hpCtxCancel = nil

	p.isHolePunching = false
	desk.SetSystemTrayIcon(gloveInactiveEmoji)
	menuItemToggle.Label = "Start Hole Punching"
	sysTrayMenu.Refresh()
}

func (p *Punchr) StopHolePunching() {
	if p.hpCtxCancel != nil {
		p.hpCtxCancel()
	}
}

func loadApiKey() (string, error) {
	r, err := storage.Reader(apiKeyURI)
	if err != nil {
		return "", errors.Wrap(err, "storage reader")
	}

	apiKeyBytes, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Wrap(err, "read all")
	}

	return string(apiKeyBytes), nil
}
