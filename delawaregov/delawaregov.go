package delawaregov

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/tekkamanendless/cboc-tools/erp"
)

type DelawareGov struct {
	browser *rod.Browser
	page    *rod.Page
}

func New(browser *rod.Browser) *DelawareGov {
	return &DelawareGov{
		browser: browser,
	}
}

func (g *DelawareGov) Login(username string, password string) error {
	fmt.Printf("Logging in to Delaware.gov...\n")

	// Create a new page
	page := g.browser.MustPage("https://id.delaware.gov").MustWaitStable()

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

	g.page = page

	return nil
}

func (g *DelawareGov) ERP() (*erp.ERP, error) {
	if g.page == nil {
		return nil, fmt.Errorf("not logged in")
	}

	return erp.New(g.browser), nil
}
