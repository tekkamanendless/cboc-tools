package database

import (
	"fmt"
	"net/url"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func New(connectionString string) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		TranslateError:           true, // Ensure that errors are properly translated into the Gorm built-in ones.
		DisableNestedTransaction: true, // Do not use SAVEPOINT statements for nested transactions; just have a single, overarching transaction.
	}
	var dialector gorm.Dialector

	{
		parsedURL, err := url.Parse(connectionString)
		if err != nil {
			return nil, err
		}
		fmt.Printf("URL: Query: %v\n", parsedURL.Query())

		targetValues := map[string][]string{
			"_pragma":      {"foreign_keys(1)", `encoding("UTF-8")`, "journal_mode(MEMORY)"},
			"_time_format": {"sqlite"}, // This is the only supported format: YYYY-MM-DDTHH:MM:SS.SSS
			"cache":        {"shared"},
		}

		currentValues := parsedURL.Query()
		for key, values := range targetValues {
			currentValue := currentValues.Get(key)
			for _, value := range values {
				if currentValue != value {
					if currentValue == "" {
						fmt.Printf("Setting query parameter %q to %q.\n", key, value)
					} else {
						fmt.Printf("Overriding query parameter %q to %q.\n", key, value)
					}
					currentValues.Add(key, value)
				}
			}
		}
		parsedURL.RawQuery = currentValues.Encode()
		connectionString = parsedURL.String()

		dialector = sqlite.Dialector{DSN: connectionString}
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}
