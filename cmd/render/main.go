package main

import (
	"flag"

	"github.com/tekkamanendless/cboc-tools/database"
)

func main() {
	var outputDirectory string
	var databaseFile string
	flag.StringVar(&outputDirectory, "output-directory", "", "The location to save the results.")
	flag.StringVar(&databaseFile, "database-file", "", "The database file.")

	flag.Parse()

	db, err := database.New("file:" + databaseFile)
	if err != nil {
		panic(err)
	}

	{
		type Row struct {
			Division         string  `gorm:"column:division"`
			BudgetAmount     float64 `gorm:"column:budget_amount"`
			EncumberedAmount float64 `gorm:"column:encumbered_amount"`
			ExpendedAmount   float64 `gorm:"column:expended_amount"`
		}
		var rows []Row
		err := db.Raw(`SELECT division, SUM(budget_amount) AS budget_amount, SUM(encumbered_amount) AS encumbered_amount, SUM(expended_amount) AS expended_amount FROM fsf_operating_unit_expenditure_summaries GROUP BY division`).
			Find(&rows).
			Error
		if err != nil {
			panic(err)
		}
	}
}
