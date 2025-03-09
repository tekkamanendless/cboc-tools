package erp

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/tekkamanendless/cboc-tools/mobius"
)

type ERP struct {
	browser *rod.Browser
	page    *rod.Page
}

func New(browser *rod.Browser) *ERP {
	return &ERP{
		browser: browser,
	}
}

func (e *ERP) Login(username string, password string) error {
	fmt.Printf("Logging in to the ERP portal...\n")

	page := e.browser.MustPage("https://portal.erp.state.de.us") //.MustWaitStable()
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

	e.page = page

	return nil
}

func (e *ERP) Mobius() (*mobius.Mobius, error) {
	if e.page == nil {
		return nil, fmt.Errorf("not logged in")
	}

	// TODO: Should we navigate to the main page again first?

	var mobiusLinkElement *rod.Element
	for _, element := range e.page.MustElements(".ps_groupleth") {
		if element.MustText() == "Mobius View" {
			mobiusLinkElement = element
			break
		}
	}
	if mobiusLinkElement == nil {
		return nil, fmt.Errorf("could not find Mobius View")
	}
	mobiusLinkElement.MustClick()
	time.Sleep(2 * time.Second)

	pages := e.browser.MustPages()
	for _, page := range pages {
		fmt.Printf("page: %s\n", page.MustInfo().URL)
	}

	page := pages.MustFindByURL("viewerpreports.dti")
	fmt.Printf("page: %s\n", page.MustInfo().URL)
	page.MustElement("button#continue").MustClick()

	page.MustWaitStable()
	page.WaitDOMStable(5*time.Second, 10)

	e.page = page

	return mobius.New(e.page), nil
}
