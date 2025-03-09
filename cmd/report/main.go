package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/tekkamanendless/cboc-tools/dataservicecenter"
)

type Config struct {
	District         string
	DelawarePassword string
	DelawareUsername string
	DSCUsername      string
	DSCPassword      string
	ERPPassword      string
	ERPUsername      string
	BaseDirectory    string
	TargetYear       int
	TargetMonth      int
}

func main() {
	var devTools bool
	var headless bool
	var slowMotion time.Duration
	var config Config
	flag.BoolVar(&devTools, "dev-tools", false, "Show the dev tools.")
	flag.BoolVar(&headless, "headless", true, "Set the headless mode.  If true, no browser will be shown.")
	flag.DurationVar(&slowMotion, "slow-motion", 0, "Set the delay between actions.")
	flag.StringVar(&config.BaseDirectory, "base-directory", "", "The location to save the results.")
	flag.StringVar(&config.District, "district", "Christina", "The district.")
	flag.StringVar(&config.DelawareUsername, "delaware-username", "", "The username.")
	flag.StringVar(&config.DelawarePassword, "delaware-password", "", "The password.")
	flag.StringVar(&config.DSCUsername, "dsc-username", "", "The username.")
	flag.StringVar(&config.DSCPassword, "dsc-password", "", "The password.")
	flag.StringVar(&config.ERPUsername, "erp-username", "", "The username.")
	flag.StringVar(&config.ERPPassword, "erp-password", "", "The password.")
	flag.IntVar(&config.TargetYear, "target-year", 0, "The target year.")
	flag.IntVar(&config.TargetMonth, "target-month", 0, "The target month.")

	flag.Parse()

	// The DSC username is the same as the Delaware username, but without the domain.
	if config.DSCUsername == "" {
		config.DSCUsername = config.DelawareUsername
		if i := strings.Index(config.DSCUsername, "@"); i > 0 {
			config.DSCUsername = config.DSCUsername[0:i]
		}
	}
	// The DSC password is the same as the Delaware password.
	if config.DSCPassword == "" {
		config.DSCPassword = config.DelawarePassword
	}

	if config.TargetYear == 0 {
		targetDate := time.Now().AddDate(0, -1, 0)
		config.TargetYear = targetDate.Year()
	}
	if config.TargetMonth == 0 {
		targetDate := time.Now().AddDate(0, -1, 0)
		config.TargetMonth = int(targetDate.Month())
	}

	if config.BaseDirectory == "" {
		config.BaseDirectory = os.TempDir()
	}

	l := launcher.New().
		Headless(headless).
		Devtools(devTools).
		Set("--disable-web-security").
		Set("--start-maximised")

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

	err := doTheThing(browser, config)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	fmt.Printf("Sleeping...\n")
	time.Sleep(10 * time.Minute)
}

func doTheThing(browser *rod.Browser, config Config) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			debug.PrintStack()
			time.Sleep(10 * time.Minute)
		}
	}()

	if config.District != "" && config.DSCUsername != "" && config.DSCPassword != "" {
		dscInstance := dataservicecenter.New(browser)
		err := dscInstance.Login(config.District, config.DSCUsername, config.DSCPassword)
		if err != nil {
			return err
		}

		fsf, err := dscInstance.FSF()
		if err != nil {
			return err
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-program-summary.csv"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsf.DownloadOperatingUnitProgramSummaryReport(config.TargetYear, config.TargetMonth, false, "csv")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-program-summary.pdf"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsf.DownloadOperatingUnitProgramSummaryReport(config.TargetYear, config.TargetMonth, false, "pdf")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-expenditure-summary.csv"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsf.DownloadOperatingUnitExpenditureSummaryReport(config.TargetYear, config.TargetMonth, []string{"33", "51", "56", "60"}, "csv")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-expenditure-summary.pdf"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsf.DownloadOperatingUnitExpenditureSummaryReport(config.TargetYear, config.TargetMonth, []string{"33", "51", "56", "60"}, "pdf")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
	}

	if config.DelawareUsername != "" && config.DelawarePassword != "" {
		err := loginToDelawareDotGov(browser, config.DelawareUsername, config.DelawarePassword)
		if err != nil {
			return err
		}
	}

	if config.ERPUsername != "" && config.ERPPassword != "" {
		err := loginToERPPortal(browser, config.ERPUsername, config.ERPPassword)
		if err != nil {
			return err
		}
	}

	return nil
}

func loginToDelawareDotGov(browser *rod.Browser, username string, password string) error {
	fmt.Printf("Logging in to Delaware.gov...\n")

	// Create a new page
	page := browser.MustPage("https://id.delaware.gov").MustWaitStable()

	formElement := page.MustElement(`form`) // Could be "#form19"

	usernameInput := formElement.MustElement(`input[autocomplete="username"]`)
	usernameInput.MustInput(username)

	passwordInput := formElement.MustElement(`input[type="password"]`)
	passwordInput.MustInput(password)

	signinButton := formElement.MustElement(`input[type="submit"]`)
	signinButton.MustClick()

	page.MustWaitStable()

	if page.MustInfo().URL == "https://id.delaware.gov/app/UserHome" {
		return fmt.Errorf("could not log in")
	}

	return nil
}

func loginToERPPortal(browser *rod.Browser, username string, password string) error {
	// Create a new page
	page := browser.MustPage("https://portal.erp.state.de.us") //.MustWaitStable()
	time.Sleep(2 * time.Second)

	formElement := page.MustElement(`form[name="login"]`)

	usernameInput := formElement.MustElement(`input[name="userid"]`)
	usernameInput.MustInput(username)

	passwordInput := formElement.MustElement(`input[type="password"]`)
	passwordInput.MustInput(password)

	agreeInput := formElement.MustElement(`input[name="agree"]`)
	agreeInput.MustClick()

	signinButton := formElement.MustElement(`input[type="submit"]`)
	signinButton.MustClick()

	page.MustWaitStable()

	/*
		if page.MustInfo().URL == "https://id.delaware.gov/app/UserHome" {
			return fmt.Errorf("could not log in")
		}
	*/
	// TODO: End up here: https://portal.erp.state.de.us/psc/ps92pd/EMPLOYEE/EMPL/c/NUI_FRAMEWORK.PT_LANDINGPAGE.GBL?

	// And then Mobius

	var mobiusLinkElement *rod.Element
	for _, element := range page.MustElements(".ps_groupleth") {
		if element.MustText() == "Mobius View" {
			mobiusLinkElement = element
			break
		}
	}
	if mobiusLinkElement == nil {
		return fmt.Errorf("could not find Mobius View")
	}
	mobiusLinkElement.MustClick()
	time.Sleep(2 * time.Second)

	pages := browser.MustPages()
	for _, page := range pages {
		fmt.Printf("page: %s\n", page.MustInfo().URL)
	}

	page = pages.MustFindByURL("viewerpreports.dti")
	fmt.Printf("page: %s\n", page.MustInfo().URL)
	page.MustElement("button#continue").MustClick()

	err := doTheMobiusWork(page)
	if err != nil {
		return err
	}

	return nil
}

func doTheMobiusWork(page *rod.Page) error {
	page.MustWaitStable()
	page.WaitDOMStable(5*time.Second, 10)

	err := clickMobiusItem(page, "First State Financials")
	if err != nil {
		return err
	}
	page.WaitDOMStable(5*time.Second, 10)

	err = clickMobiusItem(page, "Reports")
	if err != nil {
		return err
	}
	page.WaitDOMStable(5*time.Second, 10)

	err = searchMobiusItems(page, "DGL060")
	if err != nil {
		return err
	}

	page.WaitDOMStable(5*time.Second, 10)

	err = clickMobiusItem(page, "DGL060")
	if err != nil {
		return err
	}
	page.WaitDOMStable(5*time.Second, 10)

	err = clickMobiusItem(page, "Jan 31, 2025 11:43:10 PM")
	if err != nil {
		return err
	}
	page.WaitDOMStable(5*time.Second, 10)

	// "953300"
	// 33, 51, 56, 60

	err = clickMobiusItem(page, "953300")
	if err != nil {
		return err
	}
	page.WaitDOMStable(5*time.Second, 10)

	// Extract
	{
		page.MustElement(`app-mobius-view-docviewer mobius-toolbar div[title="Extract"]`).MustClick()
		page.WaitDOMStable(5*time.Second, 10)
		page.WaitDOMStable(5*time.Second, 10)
		page.WaitDOMStable(5*time.Second, 10)

		err = clickMobiusItem(page, "DGL060")
		if err != nil {
			return err
		}
		page.WaitDOMStable(5*time.Second, 10)

		page.MustElement(`app-mobius-view-extract-results mobius-toolbar div[title="Export"]`).MustClick()
		page.WaitDOMStable(5*time.Second, 10)

		page.MustElement(`ngb-modal-window [placeholder="File Name"]`).MustType(input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace, input.Backspace)
		page.MustElement(`ngb-modal-window [placeholder="File Name"]`).MustInput("DLG060_953300")
		page.WaitDOMStable(5*time.Second, 10)

		page.MustElement(`ngb-modal-window mobius-ui-checkbox#dontZipDownloadFile a`).MustClick()
		page.WaitDOMStable(5*time.Second, 10)

		fmt.Printf("Waiting for download.\n")
		download := page.Browser().MustWaitDownload()

		page.MustElement(`ngb-modal-window button.btn-submit`).MustClick()
		page.WaitDOMStable(5*time.Second, 10)

		fmt.Printf("Downloading...\n")
		contents := download()
		fmt.Printf("Downloaded %d bytes.\n", len(contents))
		os.WriteFile("/tmp/DLG060_95330.csv", contents, 0644)
	}

	// mobius-ui-content-breadcrumb
	/*
			<a _ngcontent-c40="" class="breadcrumb-item" href="#">
		                DGL060
		            </a>
	*/

	return nil
}

func clickMobiusItem(page *rod.Page, name string) error {
	itemMap, err := getMobiusItems(page)
	if err != nil {
		return err
	}
	targetItemElement := itemMap[name]
	if targetItemElement == nil {
		return fmt.Errorf("could not find item: %s", name)
	}
	targetItemElement.MustClick()

	return nil
}

func getMobiusItems(page *rod.Page) (map[string]*rod.Element, error) {
	output := map[string]*rod.Element{}
	itemElements := page.MustElements("app-mobius-view-content-list mobius-content-list mobius-content-item .content-item-label")
	for _, itemElement := range itemElements {
		text := strings.TrimSpace(itemElement.MustText())
		fmt.Printf("getMobiusItems: item: %s\n", text)
		output[text] = itemElement
	}
	return output, nil
}

func searchMobiusItems(page *rod.Page, input string) error {
	inputElement := page.MustElement("app-mobius-view-content-list mobius-content-list mobius-content-filter input")
	inputElement.MustInput(input)
	return nil
}
