package mobius

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

type Mobius struct {
	page *rod.Page
}

func New(page *rod.Page) *Mobius {
	return &Mobius{
		page: page,
	}
}

func (m *Mobius) GoToReport(path []string) error {
	breadcrumbs := m.breadcrumbs()
	var lastBreadcrumb *Breadcrumb
	remainingPath := append([]string{}, path...)
	for _, breadcrumb := range breadcrumbs {
		fmt.Printf("Breadcrumb: %s\n", breadcrumb.Name)
		for pathIndex, pathPart := range path {
			if strings.EqualFold(breadcrumb.Name, pathPart) {
				lastBreadcrumb = &breadcrumb
				remainingPath = remainingPath[pathIndex+1:]
			}
		}
	}
	if lastBreadcrumb == nil {
		return fmt.Errorf("could not find a breadcrumb for the path: %v", path)
	}

	for _, pathPart := range remainingPath {
		err := m.SearchItems(pathPart)
		if err != nil {
			return err
		}

		err = m.ClickItem(pathPart)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Mobius) ExtractReport(name string) ([]byte, error) {
	err := m.ClickItem(name)
	if err != nil {
		return nil, err
	}

	page := m.page

	// Extract
	page.MustElement(`app-mobius-view-docviewer mobius-toolbar div[title="Extract"]`).MustClick()
	page.WaitDOMStable(5*time.Second, 10)
	page.WaitDOMStable(5*time.Second, 10)
	page.WaitDOMStable(5*time.Second, 10)

	err = m.ClickItem(name)
	if err != nil {
		return nil, err
	}
	page.WaitDOMStable(5*time.Second, 10)

	page.MustElement(`app-mobius-view-extract-results mobius-toolbar div[title="Export"]`).MustClick()
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

	// TODO: Close the file preview?

	return contents, nil
}

type Breadcrumb struct {
	Name    string
	Element *rod.Element
}

func (m *Mobius) breadcrumbs() []Breadcrumb {
	/*
			<a _ngcontent-c40="" class="breadcrumb-item" href="#">
		                DGL060
		            </a>
	*/

	var output []Breadcrumb
	breadcrumbElements := m.page.MustElements(`mobius-ui-content-breadcrumb a.breadcrumb-item`)
	for _, breadcrumbElement := range breadcrumbElements {
		breadcrumb := Breadcrumb{
			Name:    strings.TrimSpace(breadcrumbElement.MustText()),
			Element: breadcrumbElement,
		}
		output = append(output, breadcrumb)
	}
	return output
}

func (m *Mobius) ClickItem(name string) error {
	itemMap, err := m.GetItems()
	if err != nil {
		return err
	}
	targetItemElement := itemMap[name]
	if targetItemElement == nil {
		return fmt.Errorf("could not find item: %s", name)
	}
	targetItemElement.MustClick()

	m.page.WaitDOMStable(5*time.Second, 10)

	return nil
}

func (m *Mobius) GetItems() (map[string]*rod.Element, error) {
	output := map[string]*rod.Element{}
	itemElements := m.page.MustElements("app-mobius-view-content-list mobius-content-list mobius-content-item .content-item-label")
	for _, itemElement := range itemElements {
		text := strings.TrimSpace(itemElement.MustText())
		fmt.Printf("getItems: item: %s\n", text)
		output[text] = itemElement
	}
	return output, nil
}

func (m *Mobius) SearchItems(input string) error {
	inputElement := m.page.MustElement("app-mobius-view-content-list mobius-content-list mobius-content-filter input")
	inputElement.MustInput(input)

	m.page.WaitDOMStable(5*time.Second, 10)
	return nil
}
