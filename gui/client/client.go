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
	"github.com/adrg/xdg"
	"github.com/emersion/go-autostart"
	"github.com/friendsofgo/errors"
	"github.com/google/uuid"
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
	Version = "dev"
)

var (
	gloveActiveEmoji   = fyne.NewStaticResource("glove-active", gloveActive)
	gloveInactiveEmoji = fyne.NewStaticResource("glove-inactive", gloveInactive)
	glovePendingEmoji  = fyne.NewStaticResource("glove-pending", glovePending)
)

type AppState struct {
	app          fyne.App
	desk         desktop.App
	autostartApp *autostart.App
	execPath     string
	events       chan any
	gui          *Gui
	apiKey       string
	apiKeyURI    fyne.URI

	hpCtx          context.Context
	hpCtxCancel    context.CancelFunc
	isHolePunching bool

	generatedApiKey          bool
	startAfterStop           bool
	isFirstStart             bool
	showedConfirmationDialog bool
}

type Gui struct {
	sysTrayMenu                   *fyne.Menu
	menuItemStatus                *fyne.MenuItem
	menuItemStartStopHolePunching *fyne.MenuItem
	menuItemSetApiKey             *fyne.MenuItem
	menuItemAutoStart             *fyne.MenuItem
	menuItemResults               *fyne.MenuItem
	menuItemVersionInfo           *fyne.MenuItem
	apiKeyDialog                  *fyne.Window
	confirmationDialog            *fyne.Window
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
			menuItemVersionInfo:           fyne.NewMenuItem("", nil),
			menuItemResults:               fyne.NewMenuItem("Your Personal Results", nil),
		},
	}, nil
}

type EvtOnStarted struct{}
type EvtToggleHolePunching struct{}
type EvtShowAPIKeyDialog struct{}
type EvtCloseAPIKeyDialog struct{}
type EvtCloseConfirmationDialog struct{}
type EvtToggleAutoStart struct{}
type EvtSaveApiKey struct{ apiKey string }
type EvtSaveConfirmation struct{ launchOnLogin bool }
type EvtStoppedHolePunching struct{}
type EvtOpenGrafanaResults struct{}

func (as *AppState) Init() error {

	apiKeyURI, err := xdg.ConfigFile("punchr/api-key.txt")
	if err != nil {
		apiKeyURI = path.Join(path.Dir(as.execPath), "api-key.txt")
	}
	as.apiKeyURI = storage.NewFileURI(apiKeyURI)

	if err := as.LoadApiKey(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.WithError(err).Warnln("Error loading API Key")
	}

	if as.apiKey == "" {
		if err = as.SaveApiKey(uuid.NewString()); err != nil {
			log.WithError(err).Warnln("Error generating and saving API Key")
		}
		as.generatedApiKey = true
	}
	as.isFirstStart = as.generatedApiKey && !as.autostartApp.IsEnabled()

	as.gui.menuItemStatus.Disabled = true
	as.gui.menuItemStartStopHolePunching.Label = "🔴 Hole Punching Stopped"
	as.gui.menuItemVersionInfo.Label = fmt.Sprintf("GUI: %s | CLI: %s", Version, client.App.Version)
	as.gui.menuItemVersionInfo.Disabled = true
	as.gui.menuItemStartStopHolePunching.Action = func() { as.events <- &EvtToggleHolePunching{} }
	as.gui.menuItemSetApiKey.Action = func() { as.events <- &EvtShowAPIKeyDialog{} }
	as.gui.menuItemAutoStart.Action = func() { as.events <- &EvtToggleAutoStart{} }
	as.gui.menuItemResults.Action = func() { as.events <- &EvtOpenGrafanaResults{} }
	as.gui.sysTrayMenu = fyne.NewMenu("Punchr", as.gui.menuItemStatus, fyne.NewMenuItemSeparator(), as.gui.menuItemSetApiKey, fyne.NewMenuItemSeparator(), as.gui.menuItemStartStopHolePunching, as.gui.menuItemAutoStart, fyne.NewMenuItemSeparator(), as.gui.menuItemResults, fyne.NewMenuItemSeparator(), as.gui.menuItemVersionInfo)

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
		go func() {
			as.events <- &EvtOnStarted{}
		}()
	})

	return nil
}

func (as *AppState) Loop() {
	for event := range as.events {
		log.Infof("Handle gui event %T\n", event)
		switch evt := event.(type) {
		case *EvtOnStarted:
			if as.generatedApiKey {
				as.ShowApiKeyDialog()
			} else if as.isFirstStart {
				as.ShowConfirmationDialog()
			} else {
				time.Sleep(time.Second)
				SetActivationPolicy()
			}
		case *EvtToggleHolePunching:
			if as.isHolePunching {
				as.StopHolePunching()
			} else {
				as.StartHolePunching()
			}
		case *EvtStoppedHolePunching:
			as.hpCtx = nil
			as.hpCtxCancel = nil
			as.desk.SetSystemTrayIcon(gloveInactiveEmoji)
			as.gui.menuItemStartStopHolePunching.Label = "🔴 Hole Punching Stopped"
			as.gui.sysTrayMenu.Refresh()
			as.isHolePunching = false

			if as.startAfterStop {
				as.startAfterStop = false
				as.StartHolePunching()
			}
		case *EvtOpenGrafanaResults:
			u, err := url.ParseRequestURI("https://punchr.dtrautwein.eu/results?apiKey=" + as.apiKey)
			if err != nil {
				log.WithError(err).Errorln("Could not parse results URI")
				break
			}
			if err = as.app.OpenURL(u); err != nil {
				log.WithError(err).Warnln("Could not open URL")
			}
		case *EvtShowAPIKeyDialog:
			if as.gui.apiKeyDialog == nil {
				as.ShowApiKeyDialog()
			}
		case *EvtToggleAutoStart:
			if as.autostartApp.IsEnabled() {
				as.DisableAutoStart()
			} else {
				as.EnableAutoStart()
			}
		case *EvtSaveApiKey:
			isNewApiKey := as.apiKey != evt.apiKey
			if err := as.SaveApiKey(evt.apiKey); err != nil {
				log.WithError(err).WithField("apiKey", evt.apiKey).Warnln("Could not save API-Key")
				return
			}
			if !as.isHolePunching {
				as.gui.menuItemStatus.Label = "API-Key: " + as.apiKey
				go as.gui.sysTrayMenu.Refresh()
			}

			if as.gui.apiKeyDialog != nil {
				go (*as.gui.apiKeyDialog).Close()
			}

			if isNewApiKey {
				as.StopHolePunching()
				as.startAfterStop = true
			}
		case *EvtCloseAPIKeyDialog:
			as.gui.apiKeyDialog = nil

			if as.isFirstStart && !as.showedConfirmationDialog {
				as.ShowConfirmationDialog()
			} else {
				SetActivationPolicy()
			}
		case *EvtSaveConfirmation:
			if evt.launchOnLogin {
				as.EnableAutoStart()
			} else {
				as.DisableAutoStart()
			}
			if as.gui.confirmationDialog != nil {
				go (*as.gui.confirmationDialog).Close()
			}
			SetActivationPolicy()

		case *EvtCloseConfirmationDialog:
			as.gui.confirmationDialog = nil
		}
	}
}

func (as *AppState) ShowConfirmationDialog() {
	window := as.app.NewWindow("All set!")
	as.gui.confirmationDialog = &window

	l := widget.NewLabel("You're all set! Hole punching has already started.\nYou can see the status in your menu bar/system tray (usually in the top right corner).\nDo you want the application to automatically launch on startup?")
	btnYes := widget.NewButton("Yes", func() {
		as.events <- &EvtSaveConfirmation{launchOnLogin: true}
	})
	btnNo := widget.NewButton("No", func() {
		as.events <- &EvtSaveConfirmation{launchOnLogin: false}
	})
	window.SetOnClosed(func() { as.events <- &EvtCloseConfirmationDialog{} })
	window.SetContent(container.New(layout.NewVBoxLayout(), l, container.New(layout.NewHBoxLayout(), btnYes, btnNo)))

	go window.Show()
	as.showedConfirmationDialog = true
}

func (as *AppState) ShowApiKeyDialog() {

	window := as.app.NewWindow("Punchr API-Key")
	as.gui.apiKeyDialog = &window

	l := widget.NewLabel("Do you have a personal API-Key? If so, enter it below.\nOtherwise you can continue with a generated one.")
	apiKeyLabel := widget.NewLabel("Your current API-Key: " + as.apiKey + "\nIf you didn't enter one it was generated for you.")
	apiKeyLabel.TextStyle = fyne.TextStyle{Italic: true}
	urlBtn := widget.NewButton("Request\npersonal API-Key", func() {
		u, err := url.ParseRequestURI("https://forms.gle/h1ABCpS87jYmg9a48")
		if err != nil {
			panic(err)
		}
		if err = as.app.OpenURL(u); err != nil {
			log.WithError(err).Warnln("Could not open URL")
		}
	})

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Please enter your API-Key")

	btn := widget.NewButton("Save API-Key", func() { as.events <- &EvtSaveApiKey{apiKey: strings.TrimSpace(entry.Text)} })
	btn.Disable()

	btnCancel := widget.NewButton("Continue", func() { window.Close() })
	window.SetOnClosed(func() { as.events <- &EvtCloseAPIKeyDialog{} })
	entry.OnChanged = func(s string) {
		s = strings.TrimSpace(s)
		if isValidUUID(s) {
			btn.Enable()
		} else {
			btn.Disable()
		}
	}

	window.SetContent(container.New(layout.NewVBoxLayout(), container.New(layout.NewHBoxLayout(), l, urlBtn), apiKeyLabel, entry, btn, btnCancel))

	go window.Show()
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

	go as.holepunchLoop()
}

func (as *AppState) holepunchLoop() {
LOOP:
	for {
		as.gui.menuItemStatus.Label = "🟢 Running..."
		as.desk.SetSystemTrayIcon(gloveActiveEmoji)
		as.gui.sysTrayMenu.Refresh()

		err := client.App.RunContext(as.hpCtx, []string{
			"punchrclient",
			"--api-key", as.apiKey,
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
			break
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

	as.events <- &EvtStoppedHolePunching{}
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

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
