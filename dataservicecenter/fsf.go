package dataservicecenter

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

type FSF struct {
	page  *rod.Page
	items []FSFItem
}

type FSFItem struct {
	Category string
	Name     string
	URL      string
}

func (f *FSF) loadItems() error {
	itemLists := f.page.MustElements(`td ol`)
	for _, itemList := range itemLists {
		// TODO: Get the category.

		items := itemList.MustElements(`li a`)
		for _, item := range items {
			itemName := item.MustText()
			itemURL := item.MustProperty("href").String()
			f.items = append(f.items, FSFItem{
				// TODO: Category: category,
				Name: itemName,
				URL:  itemURL,
			})
		}
	}

	fmt.Printf("Items: %+v\n", f.items)
	return nil
}

func (f *FSF) Items() []FSFItem {
	return f.items
}

func (f *FSF) Item(name string) (FSFItem, error) {
	for _, item := range f.items {
		if strings.EqualFold(item.Name, name) {
			return item, nil
		}
	}
	return FSFItem{}, fmt.Errorf("item not found")
}

func (f *FSF) DownloadOperatingUnitProgramSummaryReport(year int, month int, totalsOnly bool, format string) ([]byte, error) {
	item, err := f.Item("Operating Unit/Program Expenditure Summary")
	if err != nil {
		return nil, err
	}

	f.page.MustNavigate(item.URL)
	f.page.MustWaitStable()

	f.page.MustElement(`select[name="ddlFiscalYear"]`).MustSelect(fmt.Sprintf("%d", year))
	f.page.MustElement(`select[name="ddlFiscalMonth"]`).MustSelect(time.Month(month).String())
	if totalsOnly {
		f.page.MustElement(`input[name="chkOperatingUnitTotals"]`).MustClick()
	}
	formatElement := f.page.MustElement(`select[name="ddlFormat"]`)
	{
		selected := false
		formatOptions := formatElement.MustElements(`option`)
		for _, formatOption := range formatOptions {
			if strings.Contains(strings.ToLower(formatOption.MustText()), strings.ToLower(format)) {
				fmt.Printf("Found the format: %s\n", formatOption.MustText())
				formatElement.MustSelect(formatOption.MustText())
				selected = true
				break
			}
		}
		if !selected {
			return nil, fmt.Errorf("format not found")
		}
	}

	fmt.Printf("Waiting for download.\n")
	download := f.page.Browser().MustWaitDownload()

	f.page.MustElement(`input[type="submit"]`).MustClick()

	fmt.Printf("Downloading...\n")
	contents := download()
	fmt.Printf("Downloaded %d bytes.\n", len(contents))

	return contents, nil
}

func (f *FSF) DownloadOperatingUnitExpenditureSummaryReport(year int, month int, divisions []string, format string) ([]byte, error) {
	item, err := f.Item("Operating Unit Expenditure Summary")
	if err != nil {
		return nil, err
	}

	f.page.MustNavigate(item.URL)
	f.page.MustWaitStable()

	f.page.MustElement(`select[name="ddlFiscalYear"]`).MustSelect(fmt.Sprintf("%d", year))
	f.page.MustElement(`select[name="ddlFiscalMonth"]`).MustSelect(time.Month(month).String())
	if divisions != nil {
		divisionMap := map[string]bool{}
		for _, division := range divisions {
			divisionMap[division] = true
		}

		divisionInputs := f.page.MustElements(`#cblDivision input[type="checkbox"]`)
		for _, divisionInput := range divisionInputs {
			var division string
			{
				labelElement := divisionInput.MustParent().MustElement(`label`)
				label := labelElement.MustText()
				parts := strings.Split(label, " ")
				division = parts[0]
			}

			var checked bool
			{
				checkedValue := divisionInput.MustAttribute("checked")
				if checkedValue != nil {
					checked = *checkedValue == "checked"
				}
			}

			if divisionMap[division] != checked {
				divisionInput.MustClick()
			}
		}
	}
	formatElement := f.page.MustElement(`select[name="ddlFormat"]`)
	{
		selected := false
		formatOptions := formatElement.MustElements(`option`)
		for _, formatOption := range formatOptions {
			if strings.Contains(strings.ToLower(formatOption.MustText()), strings.ToLower(format)) {
				fmt.Printf("Found the format: %s\n", formatOption.MustText())
				formatElement.MustSelect(formatOption.MustText())
				selected = true
				break
			}
		}
		if !selected {
			return nil, fmt.Errorf("format not found")
		}
	}

	fmt.Printf("Waiting for download.\n")
	download := f.page.Browser().MustWaitDownload()

	f.page.MustElement(`input[type="submit"]`).MustClick()

	fmt.Printf("Downloading...\n")
	contents := download()
	fmt.Printf("Downloaded %d bytes.\n", len(contents))

	return contents, nil
}

func (f *FSF) DownloadDetailedActivityReport(startDate, endDate time.Time, divisions []string, format string) ([]byte, error) {
	item, err := f.Item("Detailed Activity List")
	if err != nil {
		return nil, err
	}

	f.page.MustNavigate(item.URL)
	f.page.MustWaitStable()

	if divisions != nil {
		divisionMap := map[string]bool{}
		for _, division := range divisions {
			divisionMap[division] = true
		}

		divisionInputs := f.page.MustElements(`#cblDivision input[type="checkbox"]`)
		for _, divisionInput := range divisionInputs {
			var division string
			{
				labelElement := divisionInput.MustParent().MustElement(`label`)
				label := labelElement.MustText()
				parts := strings.Split(label, " ")
				division = parts[0]
			}

			var checked bool
			{
				checkedValue := divisionInput.MustAttribute("checked")
				if checkedValue != nil {
					checked = *checkedValue == "checked"
				}
			}

			if divisionMap[division] != checked {
				divisionInput.MustClick()
			}
		}
	}
	fmt.Printf("Start date: %v\n", startDate)
	fmt.Printf("End date: %v\n", endDate)

	// The date inputs are weird; they tend to auto-select and move around when you try to mess with them.
	// We're going to backspace everything and then try to delete everything, and then we can enter the values.

	f.page.MustElement(`input[name="dbxAccountingDateStart"]`).MustClick()
	f.page.MustElement(`input[name="dbxAccountingDateStart"]`).Type(slices.Repeat([]input.Key{input.Backspace}, 30)...)
	f.page.MustElement(`input[name="dbxAccountingDateStart"]`).Type(slices.Repeat([]input.Key{input.Delete}, 30)...)
	f.page.MustElement(`input[name="dbxAccountingDateStart"]`).MustInput(startDate.Format("1/2/2006"))

	f.page.MustElement(`input[name="dbxAccountingDateEnd"]`).MustClick()
	f.page.MustElement(`input[name="dbxAccountingDateEnd"]`).Type(slices.Repeat([]input.Key{input.Backspace}, 30)...)
	f.page.MustElement(`input[name="dbxAccountingDateEnd"]`).Type(slices.Repeat([]input.Key{input.Delete}, 30)...)
	f.page.MustElement(`input[name="dbxAccountingDateEnd"]`).MustInput(endDate.Format("1/2/2006"))

	{
		input := f.page.MustElement(`input#cbBudgetRefAll`)
		var checked bool
		{
			checkedValue := input.MustAttribute("checked")
			if checkedValue != nil {
				checked = *checkedValue == "checked"
			}
		}

		if !checked {
			input.MustClick()
		}
	}
	formatElement := f.page.MustElement(`select[name="ddlFormat"]`)
	{
		selected := false
		formatOptions := formatElement.MustElements(`option`)
		for _, formatOption := range formatOptions {
			if strings.Contains(strings.ToLower(formatOption.MustText()), strings.ToLower(format)) {
				fmt.Printf("Found the format: %s\n", formatOption.MustText())
				formatElement.MustSelect(formatOption.MustText())
				selected = true
				break
			}
		}
		if !selected {
			return nil, fmt.Errorf("format not found")
		}
	}

	fmt.Printf("Waiting for download.\n")
	download := f.page.Browser().MustWaitDownload()

	f.page.MustElement(`input[type="submit"]`).MustClick()

	fmt.Printf("Downloading...\n")
	contents := download()
	fmt.Printf("Downloaded %d bytes.\n", len(contents))

	return contents, nil
}

// TODO: Total Expenditure Report

// TODO: District Revenue Report
