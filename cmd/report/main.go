package main

import (
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/tekkamanendless/cboc-tools/dataservicecenter"
	"github.com/tekkamanendless/cboc-tools/delawaregov"
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

	divisions := []string{"33", "51", "56", "60"}
	fmt.Printf("Divisions: %v\n", divisions)

	if config.District != "" && config.DSCUsername != "" && config.DSCPassword != "" {
		fmt.Printf("Doing: DSC\n")

		dscInstance := dataservicecenter.New(browser)
		err := dscInstance.Login(config.District, config.DSCUsername, config.DSCPassword)
		if err != nil {
			return err
		}

		fsfInstance, err := dscInstance.FSF()
		if err != nil {
			return err
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-program-summary.csv"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsfInstance.DownloadOperatingUnitProgramSummaryReport(config.TargetYear, config.TargetMonth, false, "csv")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-program-summary.pdf"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsfInstance.DownloadOperatingUnitProgramSummaryReport(config.TargetYear, config.TargetMonth, false, "pdf")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-expenditure-summary.csv"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsfInstance.DownloadOperatingUnitExpenditureSummaryReport(config.TargetYear, config.TargetMonth, divisions, "csv")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
		{
			fileName := config.BaseDirectory + string(filepath.Separator) + "fsf.operating-unit-expenditure-summary.pdf"
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				contents, err := fsfInstance.DownloadOperatingUnitExpenditureSummaryReport(config.TargetYear, config.TargetMonth, divisions, "pdf")
				if err != nil {
					return err
				}

				os.WriteFile(fileName, contents, 0644)
			}
		}
	}

	if config.DelawareUsername != "" && config.DelawarePassword != "" {
		fmt.Printf("Doing: Delaware.gov\n")

		delawareGovInstance := delawaregov.New(browser)
		err := delawareGovInstance.Login(config.DelawareUsername, config.DelawarePassword)
		if err != nil {
			return err
		}

		if config.ERPUsername != "" && config.ERPPassword != "" {
			fmt.Printf("Doing: ERP\n")

			erpInstance, err := delawareGovInstance.ERP()
			if err != nil {
				return err
			}

			err = erpInstance.Login(config.ERPUsername, config.ERPPassword)
			if err != nil {
				return err
			}

			mobiusInstance, err := erpInstance.Mobius()
			if err != nil {
				return err
			}

			reportNames := []string{
				"DGL060",
				"DGL114",
				"DGL115",
			}
			for _, reportName := range reportNames {
				path := []string{"Repositories", "First State Financials", "Reports", reportName}

				err := mobiusInstance.GoToReport(path)
				if err != nil {
					return err
				}

				var dateFile string
				{
					itemMap, err := mobiusInstance.GetItemsN(400)
					if err != nil {
						return err
					}

					lastDateOfMonth := time.Date(config.TargetYear, time.Month(config.TargetMonth)+1, 1, 0, 0, -1, 0, time.Local)
					var lastDate time.Time
					var lastDateFile string
					var firstDateAfterMonth time.Time
					var firstDateAfterMonthFile string
					for dateName := range maps.Keys(itemMap) {
						t, err := time.Parse("Jan 2, 2006 3:04:05 PM", dateName)
						if err != nil {
							fmt.Printf("Could not parse date %q: %v\n", dateName, err)
							continue
						}
						if t.After(lastDateOfMonth) && (firstDateAfterMonth.IsZero() || t.Before(firstDateAfterMonth)) {
							firstDateAfterMonth = t
							firstDateAfterMonthFile = dateName
						}
						if t.Year() != config.TargetYear || int(t.Month()) != config.TargetMonth {
							continue
						}
						if lastDate.IsZero() || t.After(lastDate) {
							lastDate = t
							lastDateFile = dateName
						}
					}

					if !firstDateAfterMonth.IsZero() {
						dateFile = firstDateAfterMonthFile
					}
					if !lastDate.IsZero() {
						if lastDate.AddDate(0, 0, 1).Month() != lastDate.Month() {
							dateFile = lastDateFile
						}
					}
				}
				if dateFile == "" {
					return fmt.Errorf("could not find a date file")
				}

				/*
					err = mobiusInstance.ClickItem(dateFile)
					if err != nil {
						return err
					}
				*/

				pathWithDate := append([]string{}, path...)
				pathWithDate = append(pathWithDate, dateFile)

				for _, division := range divisions {
					fmt.Printf("Exporting report %s for division %s.\n", reportName, division)

					reportFile := fmt.Sprintf("95%s00", division)

					err = mobiusInstance.GoToReport(path)
					if err != nil {
						return err
					}

					err = mobiusInstance.GoToReport(pathWithDate)
					if err != nil {
						return err
					}

					fileName := config.BaseDirectory + string(filepath.Separator) + "mobius." + reportName + "." + division + ".csv"
					if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
						contents, err := mobiusInstance.ExtractReport(reportName, reportFile)
						if err != nil {
							return err
						}

						os.WriteFile(fileName, contents, 0644)
					}
				}
			}
		}
	}

	return nil
}
