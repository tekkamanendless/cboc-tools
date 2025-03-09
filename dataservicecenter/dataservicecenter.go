package dataservicecenter

import (
	"fmt"
	"strings"

	"github.com/go-rod/rod"
)

type DataServiceCenter struct {
	browser      *rod.Browser
	page         *rod.Page
	applications []Application
}

type Application struct {
	Name string
	URL  string
}

func New(browser *rod.Browser) *DataServiceCenter {
	return &DataServiceCenter{
		browser: browser,
	}
}

func (d *DataServiceCenter) Login(district string, username string, password string) error {
	page := d.browser.MustPage("https://secure.dataservice.org/Logon/").MustWaitStable()

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

	d.page = page

	err := d.loadApplications()
	if err != nil {
		return fmt.Errorf("could not load applications: %w", err)
	}
	return nil
}

func (d *DataServiceCenter) Applications() []Application {
	return d.applications
}

func (d *DataServiceCenter) Application(name string) (Application, error) {
	for _, application := range d.applications {
		if strings.EqualFold(application.Name, name) {
			return application, nil
		}
	}
	return Application{}, fmt.Errorf("application not found")
}

func (d *DataServiceCenter) loadApplications() error {
	if d.page == nil {
		return fmt.Errorf("not logged in")
	}

	cardHeaders := d.page.MustElements(`.card .card-header`)
	for _, cardHeader := range cardHeaders {
		if strings.ToLower(cardHeader.MustText()) != "applications" {
			continue
		}

		listElement := cardHeader.MustParent().MustElement(`.list-group`)
		listItems := listElement.MustElements(`a.list-group-item`)
		for _, listItem := range listItems {
			applicationName := listItem.MustText()
			applicationURL := listItem.MustProperty("href").String()
			d.applications = append(d.applications, Application{
				Name: applicationName,
				URL:  applicationURL,
			})
		}
	}

	fmt.Printf("Applications: %+v\n", d.applications)
	return nil
}

func (d *DataServiceCenter) FSF() (*FSF, error) {
	if d.page == nil {
		return nil, fmt.Errorf("not logged in")
	}

	application, err := d.Application("Finance Reporting (FSF)")
	if err != nil {
		return nil, err
	}

	d.page.Navigate(application.URL)
	d.page.MustWaitStable()
	fsf := &FSF{
		page: d.page,
	}
	fsf.loadItems()

	return fsf, nil
}
