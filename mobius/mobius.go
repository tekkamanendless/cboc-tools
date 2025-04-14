package mobius

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
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
	fmt.Printf("GoToReport: %v\n", path)

	breadcrumbs := m.breadcrumbs()
	fmt.Printf("breadcrumbs: %+v\n", breadcrumbs)

	var lastBreadcrumb *Breadcrumb
	remainingPath := append([]string{}, path...)
	for _, breadcrumb := range breadcrumbs {
		fmt.Printf("Breadcrumb: %s\n", breadcrumb.Name)
		for _, pathPart := range path {
			if strings.EqualFold(breadcrumb.Name, pathPart) {
				lastBreadcrumb = &breadcrumb
				remainingPath = remainingPath[1:]
			}
		}
	}
	if lastBreadcrumb == nil {
		return fmt.Errorf("could not find a breadcrumb for the path: %v", path)
	}
	fmt.Printf("Last breadcrumb: %+v\n", *lastBreadcrumb)
	fmt.Printf("Remaining path: %v\n", remainingPath)

	{
		err := lastBreadcrumb.Element.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			if strings.Contains(err.Error(), "pointer-events is none") {
				fmt.Printf("Not clicking because we can't.\n")
			} else {
				fmt.Printf("Could not click: %v\n", err)
			}
		} else {
			m.page.WaitDOMStable(5*time.Second, 10)
		}
	}

	{
		for _, pathPart := range remainingPath {
			itemMap, err := m.GetItems()
			if err != nil {
				return fmt.Errorf("could not get items: %w", err)
			}
			if _, ok := itemMap[pathPart]; !ok {
				err := m.SearchItems(pathPart)
				if err != nil {
					return err
				}
			}

			err = m.ClickItem(pathPart)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Mobius) ExtractReport(reportName string, division string) ([]byte, error) {
	fmt.Printf("ExtractReport: reportName=%s division=%s\n", reportName, division)

	itemMap, err := m.GetItems()
	if err != nil {
		return nil, fmt.Errorf("could not get items: %w", err)
	}
	if _, ok := itemMap[division]; !ok {
		err := m.SearchItems(division)
		if err != nil {
			return nil, err
		}
	}

	err = m.ClickItem(division)
	if err != nil {
		return nil, err
	}

	page := m.page

	// Extract
	page.MustElement(`app-mobius-view-docviewer mobius-toolbar div[title="Extract"]`).MustClick()
	page.WaitDOMStable(5*time.Second, 10)
	page.WaitDOMStable(5*time.Second, 10)
	page.WaitDOMStable(5*time.Second, 10)

	err = m.ClickItem(reportName)
	if err != nil {
		return nil, err
	}
	page.WaitDOMStable(5*time.Second, 10)

	page.MustElement(`app-mobius-view-extract-results mobius-toolbar div[title="Export"]`).MustClick()
	page.WaitDOMStable(5*time.Second, 10)

	{
		dontZipElement := page.MustElement(`ngb-modal-window mobius-ui-checkbox#dontZipDownloadFile`)
		checked := false
		{
			dontZipCheckboxElement := dontZipElement.MustElement(`.basicCheckbox`)
			classNamesString := dontZipCheckboxElement.MustAttribute("class")
			if classNamesString != nil {
				classNames := strings.Split(*classNamesString, " ")
				checked = slices.Contains(classNames, "checked")
			}
		}
		if !checked {
			dontZipElement.MustElement(`a`).MustClick()
			page.WaitDOMStable(5*time.Second, 10)
		}
	}

	fmt.Printf("Waiting for download.\n")
	download := page.Browser().MustWaitDownload()

	page.MustElement(`ngb-modal-window button.btn-submit`).MustClick()
	page.WaitDOMStable(5*time.Second, 10)

	fmt.Printf("Downloading...\n")
	contents := download()
	fmt.Printf("Downloaded %d bytes.\n", len(contents))

	// Close the file preview.
	page.MustElement(`app-mobius-view-extract-results mobius-ui-dv-close`).MustClick()
	page.WaitDOMStable(5*time.Second, 10)

	// TODO: Close the file preview?
	/*
			<mobius-ui-dv-close _ngcontent-c34="" class="mx-1 ng-star-inserted" _nghost-c69=""><div _ngcontent-c69="" class="ccontainer">
		  <a _ngcontent-c69="" title="Close">
		    <mobius-icon _ngcontent-c69="" _nghost-c6=""><!----><!---->
		<!---->
		<!----><div _ngcontent-c6="" class="inline basicSvgIcon smallHighlightIcon ng-star-inserted">
		    <svg _ngcontent-c6="" xmlns:xlink="http://www.w3.org/1999/xlink" height="100%" version="1.1" viewBox="0 0 16 16" width="100%">
		        <use _ngcontent-c6="" xlink:href="#close"></use>
		    </svg>
		</div>
		<!---->

		</mobius-icon>
		  </a>
		</div>
		</mobius-ui-dv-close>
	*/

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
	breadcrumbElement := m.page.MustElement(`mobius-ui-content-breadcrumb`) // Only get the first one.
	breadcrumbElements := breadcrumbElement.MustElements(`a.breadcrumb-item`)
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
	fmt.Printf("ClickItem: %s\n", name)

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

	itemMap, err = m.GetItems()
	if err != nil {
		return err
	}
	targetItemElement = itemMap[name]
	if targetItemElement != nil {
		fmt.Printf("No dice; trying again.\n")

		targetItemElement.MustClick()

		m.page.WaitDOMStable(5*time.Second, 10)
	}

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

func (m *Mobius) GetItemsN(limit int) (map[string]*rod.Element, error) {
	output := map[string]*rod.Element{}

	for {
		originalLength := len(output)
		newOutput, err := m.GetItems()
		if err != nil {
			return nil, err
		}
		for key, value := range newOutput {
			output[key] = value
		}
		newLength := len(output)

		if newLength == originalLength {
			break
		}
		if newLength >= limit {
			break
		}

		err = m.page.Mouse.Scroll(0, 500, 4)
		if err != nil {
			fmt.Printf("Could not scroll: %v\n", err)
		}
		m.page.WaitDOMStable(1*time.Second, 10)
	}

	return output, nil
}

func (m *Mobius) SearchItems(searchText string) error {
	fmt.Printf("SearchItems: %s\n", searchText)

	//panic("oops")

	t, err := time.Parse("Jan 2, 2006 3:04:05 PM", searchText)
	if err == nil {
		searchText = t.Format("20060102150405")
	}

	inputElement := m.page.MustElement("app-mobius-view-content-list mobius-content-list mobius-content-filter input")
	inputElement.Type(slices.Repeat([]input.Key{input.Backspace}, 30)...)
	inputElement.MustInput(searchText)

	m.page.WaitDOMStable(5*time.Second, 10)
	return nil
}
