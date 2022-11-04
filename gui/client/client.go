package client

import (
	"C"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
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

//go:embed glove-pending.png
var glovePending []byte

var (
	gloveActiveEmoji   = fyne.NewStaticResource("glove-active", gloveActive)
	gloveInactiveEmoji = fyne.NewStaticResource("glove-inactive", gloveInactive)
	glovePendingEmoji  = fyne.NewStaticResource("glove-pending", glovePending)
)

type AppState struct {
	app            fyne.App
	desk           desktop.App
	autostartApp   *autostart.App
	execPath       string
	isHolePunching bool
	events         chan any
	gui            *Gui
	apiKey         string
	apiKeyURI      fyne.URI

	hpCtx       context.Context
	hpCtxCancel context.CancelFunc
}

type Gui struct {
	sysTrayMenu                   *fyne.Menu
	menuItemStatus                *fyne.MenuItem
	menuItemStartStopHolePunching *fyne.MenuItem
	menuItemSetApiKey             *fyne.MenuItem
	menuItemAutoStart             *fyne.MenuItem
	apiKeyDialog                  *fyne.Window
}

func NewAppState() (*AppState, error) {
	fapp := app.New()
	fapp.SetIcon(gloveInactiveEmoji)

	desk, isDesktopApp := fapp.(desktop.App)
	if !isDesktopApp {
		return nil, fmt.Errorf("no desktop app")
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "determine executable path")
	}

	autostartApp := &autostart.App{
		Name:        "com.protocol.ai.punchr",
		DisplayName: "Punchr Client",
		Exec:        []string{execPath},
	}

	return &AppState{
		app:            fapp,
		desk:           desk,
		autostartApp:   autostartApp,
		isHolePunching: false,
		events:         make(chan any),
		execPath:       execPath,
		gui: &Gui{
			menuItemStatus:                fyne.NewMenuItem("", nil),
			menuItemStartStopHolePunching: fyne.NewMenuItem("", nil),
			menuItemSetApiKey:             fyne.NewMenuItem("Set API Key", nil),
			menuItemAutoStart:             fyne.NewMenuItem("", nil),
		},
	}, nil
}

type EvtToggleHolePunching struct{}
type EvtShowAPIKeyDialog struct{}
type EvtCloseAPIKeyDialog struct{}
type EvtToggleAutoStart struct{}
type EvtSaveApiKey struct{ apiKey string }

func (as *AppState) Init() error {

	as.apiKeyURI = storage.NewFileURI(path.Join(path.Dir(as.execPath), "api-key.txt"))

	if err := as.LoadApiKey(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.WithError(err).Warnln("Error loading API Key")
	}

	as.gui.menuItemStatus.Disabled = true
	as.gui.menuItemStartStopHolePunching.Label = "🔴 Hole Punching Stopped"
	as.gui.menuItemStartStopHolePunching.Action = func() { as.events <- &EvtToggleHolePunching{} }
	as.gui.menuItemSetApiKey.Action = func() { as.events <- &EvtShowAPIKeyDialog{} }
	as.gui.menuItemAutoStart.Action = func() { as.events <- &EvtToggleAutoStart{} }
	as.gui.sysTrayMenu = fyne.NewMenu("Punchr", as.gui.menuItemStatus, fyne.NewMenuItemSeparator(), as.gui.menuItemSetApiKey, fyne.NewMenuItemSeparator(), as.gui.menuItemStartStopHolePunching, as.gui.menuItemAutoStart)

	as.desk.SetSystemTrayMenu(as.gui.sysTrayMenu)

	if as.apiKey == "" {
		as.gui.menuItemStatus.Label = "Please set an API-Key below"
		as.gui.menuItemStartStopHolePunching.Disabled = true
	} else {
		go as.StartHolePunching()
	}

	if as.autostartApp.IsEnabled() {
		as.gui.menuItemAutoStart.Label = "🟢 Launch on Login: Enabled"
	} else {
		as.gui.menuItemAutoStart.Label = "🔴 Launch on Login: Disabled"
	}

	as.gui.sysTrayMenu.Refresh()

	as.app.Lifecycle().SetOnStarted(func() {
		if as.apiKey == "" {
			go as.ShowApiKeyDialog()
		}

		go func() {
			time.Sleep(time.Second)
			SetActivationPolicy()
		}()
	})

	return nil
}

func (as *AppState) Loop() {
	for event := range as.events {
		switch evt := event.(type) {
		case *EvtToggleHolePunching:
			if as.isHolePunching {
				go as.StopHolePunching()
			} else {
				go as.StartHolePunching()
			}
		case *EvtShowAPIKeyDialog:
			if as.gui.apiKeyDialog == nil {
				go as.ShowApiKeyDialog()
			}
		case *EvtToggleAutoStart:
			if as.autostartApp.IsEnabled() {
				as.DisableAutoStart()
			} else {
				as.EnableAutoStart()
			}
		case *EvtSaveApiKey:
			if err := as.SaveApiKey(evt.apiKey); err != nil {
				log.WithError(err).WithField("apiKey", evt.apiKey).Warnln("Could not save API-Key")
				return
			}

			as.gui.menuItemStatus.Label = "API-Key: " + as.apiKey
			as.gui.menuItemStartStopHolePunching.Disabled = false
			as.gui.sysTrayMenu.Refresh()

			if as.gui.apiKeyDialog != nil {
				(*as.gui.apiKeyDialog).Close()
			}
		case *EvtCloseAPIKeyDialog:
			as.gui.apiKeyDialog = nil
		}
	}
}

func (as *AppState) ShowApiKeyDialog() {

	window := as.app.NewWindow("Punchr API-Key")
	as.gui.apiKeyDialog = &window

	window.Resize(fyne.NewSize(300, 100))

	l := widget.NewLabel("If you don't have an API-Key yet please request one here:")
	urlBtn := widget.NewButton("Request API-Key", func() {
		u, err := url.ParseRequestURI("https://forms.gle/gwc4NtgdFbKcaeza9")
		if err != nil {
			panic(err)
		}
		if err = as.app.OpenURL(u); err != nil {
			log.WithError(err).Warnln("Could not open URL")
		}
	})

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Please enter your API-Key")
	entry.SetText(as.apiKey)

	btn := widget.NewButton("Save API-Key", func() { as.events <- &EvtSaveApiKey{apiKey: entry.Text} })
	window.SetOnClosed(func() { as.events <- &EvtCloseAPIKeyDialog{} })
	entry.OnChanged = func(s string) {
		if s == "" {
			btn.Disable()
		} else {
			btn.Enable()
		}
	}

	if as.apiKey == "" {
		btn.Disable()
	}

	window.SetContent(container.New(layout.NewVBoxLayout(), container.New(layout.NewHBoxLayout(), l, urlBtn), entry, btn))
	window.Show()
}

func (as *AppState) SaveApiKey(newApiKey string) error {
	rwc, err := storage.Writer(as.apiKeyURI)
	if err != nil {
		return fmt.Errorf("opening storage writer: %w", err)
	}

	if _, err = rwc.Write([]byte(newApiKey)); err != nil {
		return fmt.Errorf("writing new api key: %w", err)
	}

	as.apiKey = newApiKey

	return nil
}

func (as *AppState) LoadApiKey() error {
	r, err := storage.Reader(as.apiKeyURI)
	if err != nil {
		return errors.Wrap(err, "storage reader")
	}

	apiKeyBytes, err := io.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "read all")
	}

	as.apiKey = string(apiKeyBytes)

	return nil
}

func (as *AppState) StopHolePunching() {
	if as.hpCtxCancel != nil {
		as.hpCtxCancel()
	}
}

func (as *AppState) StartHolePunching() {
	as.desk.SetSystemTrayIcon(gloveActiveEmoji)
	as.gui.menuItemStartStopHolePunching.Label = "🟢 Hole Punching Started"

	ctx, cancel := context.WithCancel(context.Background())
	as.hpCtx = ctx
	as.hpCtxCancel = cancel
	as.isHolePunching = true

LOOP:
	for {
		as.gui.menuItemStatus.Label = "🟢 Running..."
		as.desk.SetSystemTrayIcon(gloveActiveEmoji)
		as.gui.sysTrayMenu.Refresh()

		err := client.App.RunContext(as.hpCtx, []string{
			"punchrclient",
			"--api-key", as.apiKey,
			"--key-file", path.Join(path.Dir(as.execPath), "punchrclient.keys"),
			"telemetry-host", "",
			"telemetry-port", "",
		})
		if err == nil || errors.Is(as.hpCtx.Err(), context.Canceled) {
			as.gui.menuItemStatus.Label = "API-Key: " + as.apiKey
			break
		}

		if err != nil && strings.Contains(err.Error(), "unauthorized") {
			as.app.SendNotification(fyne.NewNotification("Unauthorized", "Your API-Key may be not correct"))
			as.gui.menuItemStatus.Label = "🔴 Error: API-Key not correct"
			break LOOP
		}

		as.desk.SetSystemTrayIcon(glovePendingEmoji)
		as.gui.menuItemStatus.Label = "🟠 Retrying: " + err.Error()
		as.gui.sysTrayMenu.Refresh()

		select {
		case <-time.After(10 * time.Second):
			continue
		case <-as.hpCtx.Done():
			as.gui.menuItemStatus.Label = "🔴 Error: " + err.Error()
			break LOOP
		}
	}

	as.hpCtx = nil
	as.hpCtxCancel = nil

	as.desk.SetSystemTrayIcon(gloveInactiveEmoji)
	as.gui.menuItemStartStopHolePunching.Label = "🔴 Hole Punching Stopped"
	as.gui.sysTrayMenu.Refresh()

	as.isHolePunching = false
}

func (as *AppState) EnableAutoStart() {
	log.Println("Enabling app auto start on login...")

	if err := as.autostartApp.Enable(); err != nil {
		log.WithError(err).Warnln("Could not enable app auto start")
	}

	as.gui.menuItemAutoStart.Label = "🟢 Launch on Login: Enabled"
	as.gui.sysTrayMenu.Refresh()
}

func (as *AppState) DisableAutoStart() {
	log.Println("Disabling app auto start on login...")

	if err := as.autostartApp.Disable(); err != nil {
		log.WithError(err).Warnln("Could not disable app auto start")
	}

	as.gui.menuItemAutoStart.Label = "🔴 Launch on Login: Disabled"
	as.gui.sysTrayMenu.Refresh()
}

func Start() {
	appState, err := NewAppState()
	if err != nil {
		log.WithError(err).Errorln("Could not get new app state")
		os.Exit(1)
	}

	if err := appState.Init(); err != nil {
		log.WithError(err).Errorln("Could not init app state")
		os.Exit(1)
	}

	go appState.Loop()

	appState.app.Run()
}
