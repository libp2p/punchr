package main

import (
	"context"
	_ "embed"
	"io"
	"os"
	"path"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/emersion/go-autostart"
	"github.com/friendsofgo/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/client"
)

//go:embed glove-active.png
var gloveActive []byte

//go:embed glove-inactive.png
var gloveInactive []byte

var (
	gloveActiveEmoji   = fyne.NewStaticResource("glove-active", gloveActive)
	gloveInactiveEmoji = fyne.NewStaticResource("glove-inactive", gloveInactive)

	sysTrayMenu    *fyne.Menu
	menuItemStatus = fyne.NewMenuItem("", nil)
	menuItemToggle = fyne.NewMenuItem("", nil)
	menuItemApiKey = fyne.NewMenuItem("Set API Key", nil)
	menuItemLogin  = fyne.NewMenuItem("", nil)
)

func main() {
	a := app.New()

	p, err := NewPunchr(a)
	if err != nil {
		log.WithError(err).Errorln("Could not instantiate new punchr")
		os.Exit(1)
	}

	autostartApp := &autostart.App{
		Name:        "com.protocol.ai.punchr",
		DisplayName: "Punchr Client",
		Exec:        []string{p.execPath},
	}

	a.SetIcon(gloveInactiveEmoji)

	desk, isDesktopApp := a.(desktop.App)
	if !isDesktopApp {
		log.Errorln("Can only operate as a Desktop application.")
		return
	}

	menuItemStatus.Disabled = true
	menuItemToggle.Label = "Start Hole Punching"
	menuItemToggle.Action = func() {
		if p.isHolePunching.Load() {
			go p.StopHolePunching()
		} else {
			go p.StartHolePunching()
		}
	}
	menuItemApiKey.Action = func() {
		p.ShowApiKeyDialog()
	}
	menuItemLogin.Action = func() {
		if autostartApp.IsEnabled() {
			if err := autostartApp.Disable(); err != nil {
				log.WithError(err).Warnln("error")
			}
			menuItemLogin.Label = "🔴 Launch on Login: Disabled"
		} else {
			log.Println("Enabling app...")
			if err := autostartApp.Enable(); err != nil {
				log.WithError(err).Warnln("error")
			}
			menuItemLogin.Label = "🟢 Launch on Login: Enabled"
		}
		sysTrayMenu.Refresh()
	}

	sysTrayMenu = fyne.NewMenu("Punchr", menuItemStatus, fyne.NewMenuItemSeparator(), menuItemToggle, menuItemApiKey, menuItemLogin)
	desk.SetSystemTrayMenu(sysTrayMenu)

	if p.apiKey == "" {
		menuItemStatus.Label = "No API-Key"
		menuItemToggle.Disabled = true
	} else {
		go p.StartHolePunching()
	}

	if autostartApp.IsEnabled() {
		menuItemLogin.Label = "🟢 Launch on Login: Enabled"
	} else {
		menuItemLogin.Label = "🔴 Launch on Login: Disabled"
	}

	sysTrayMenu.Refresh()

	a.Run()
}

type Punchr struct {
	app            fyne.App
	hpCtx          context.Context
	hpCtxCancel    context.CancelFunc
	isHolePunching *atomic.Bool
	apiKey         string
	execPath       string
	apiKeyURI      fyne.URI
}

func NewPunchr(app fyne.App) (*Punchr, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "determine executable path")
	}

	apiKeyURI := storage.NewFileURI(path.Join(path.Dir(execPath), "api-key.txt"))

	apiKey, err := loadApiKey(apiKeyURI)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.WithError(err).Warnln("Error loading API Key")
	}

	return &Punchr{
		app:            app,
		isHolePunching: &atomic.Bool{},
		apiKey:         apiKey,
		execPath:       execPath,
		apiKeyURI:      apiKeyURI,
	}, nil
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
	rwc, err := storage.Writer(p.apiKeyURI)
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
	if p.isHolePunching.Swap(true) {
		return
	}

	desk := p.app.(desktop.App)

	desk.SetSystemTrayIcon(gloveActiveEmoji)
	menuItemToggle.Label = "Stop Hole Punching"

	ctx, cancel := context.WithCancel(context.Background())
	p.hpCtx = ctx
	p.hpCtxCancel = cancel

LOOP:
	for {
		menuItemStatus.Label = "🟢 Running..."
		sysTrayMenu.Refresh()

		err := client.App.RunContext(p.hpCtx, []string{
			"punchrclient",
			"--api-key", p.apiKey,
			"--key-file", path.Join(path.Dir(p.execPath), "punchrclient.keys"),
			"telemetry-host", "",
			"telemetry-port", "",
		})

		if err == nil || errors.Is(p.hpCtx.Err(), context.Canceled) {
			menuItemStatus.Label = "API-Key: " + p.apiKey
			break
		}

		menuItemStatus.Label = "🟠 Retrying: " + err.Error()
		sysTrayMenu.Refresh()

		select {
		case <-time.After(10 * time.Second):
			continue
		case <-p.hpCtx.Done():
			menuItemStatus.Label = "🔴 Error: " + err.Error()
			break LOOP
		}
	}

	p.hpCtx = nil
	p.hpCtxCancel = nil

	desk.SetSystemTrayIcon(gloveInactiveEmoji)
	menuItemToggle.Label = "Start Hole Punching"
	sysTrayMenu.Refresh()

	p.isHolePunching.Swap(false)
}

func (p *Punchr) StopHolePunching() {
	if p.hpCtxCancel != nil {
		p.hpCtxCancel()
	}
}

func loadApiKey(apiKeyURI fyne.URI) (string, error) {
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
