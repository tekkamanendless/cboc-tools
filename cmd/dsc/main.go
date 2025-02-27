package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	var devTools bool
	var district string
	var headless bool
	var password string
	var slowMotion time.Duration
	var username string
	flag.BoolVar(&devTools, "dev-tools", false, "Show the dev tools.")
	flag.BoolVar(&headless, "headless", true, "Set the headless mode.  If true, no browser will be shown.")
	flag.DurationVar(&slowMotion, "slow-motion", 0, "Set the delay between actions.")
	flag.StringVar(&district, "district", "Christina", "The district.")
	flag.StringVar(&username, "username", "", "The username.")
	flag.StringVar(&password, "password", "", "The password.")

	flag.Parse()

	l := launcher.New().
		Headless(headless).
		Devtools(devTools)

	defer l.Cleanup()

	controlURL := l.MustLaunch()

	// Launch a new browser with default options, and connect to it.
	browser := rod.New().
		ControlURL(controlURL).
		Trace(true).
		SlowMotion(slowMotion).
		MustConnect()

	// Even you forget to close, rod will close it after main process ends.
	defer browser.MustClose()

	err := login(browser, district, username, password)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sleeping...\n")
	time.Sleep(10 * time.Second)
}

func login(browser *rod.Browser, district string, username string, password string) error {
	// Create a new page
	page := browser.MustPage("https://secure.dataservice.org/Logon/").MustWaitStable()

	formElement := page.MustElement(`form#loginForm`)
	districtSelect := formElement.MustElement(`select[name="Input.District"]`)
	districtSelect.MustSelect(district)

	usernameInput := formElement.MustElement(`input[name="Input.Username"]`)
	usernameInput.MustInput(username)

	passwordInput := formElement.MustElement(`input[name="Input.Password"]`)
	passwordInput.MustInput(password)

	signinButton := formElement.MustElement(`button[type="submit"]`)
	signinButton.MustClick()

	page.MustWaitStable()

	if page.MustInfo().URL == "https://secure.dataservice.org/Logon/" {
		return fmt.Errorf("could not log in")
	}

	return nil
}
