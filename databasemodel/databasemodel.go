package databasemodel

import (
	"fmt"

	"gorm.io/gorm"
)

func Apply(db *gorm.DB) error {
	err := db.AutoMigrate(
		&FSFOperatingUnitExpenditureSummary{},
		&FSFOperatingUnitProgramSummary{},
		&MobiusDGL060{},
		&MobiusDGL114{},
		&MobiusDGL115{},
	)
	if err != nil {
		return fmt.Errorf("could not migrate tables: %w", err)
	}
	return nil
}
