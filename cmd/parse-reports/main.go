package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tekkamanendless/cboc-tools/database"
	"github.com/tekkamanendless/cboc-tools/databasemodel"
)

func main() {
	var baseDirectory string
	var databaseFile string
	flag.StringVar(&baseDirectory, "base-directory", "", "The location to save the results.")
	flag.StringVar(&databaseFile, "database-file", "", "The database file.")

	flag.Parse()

	db, err := database.New("file:" + databaseFile)
	if err != nil {
		panic(err)
	}

	err = databasemodel.Apply(db)
	if err != nil {
		panic(err)
	}

	{
		filename := baseDirectory + string(filepath.Separator) + "fsf.operating-unit-expenditure-summary.csv"
		fmt.Printf("filename: %s\n", filename)

		if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", filename)
			panic(err)
		} else {
			fileHandle, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			csvReader := csv.NewReader(fileHandle)
			rows, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}

			if len(rows) == 0 {
				fmt.Printf("No rows found in the CSV file.\n")
			} else {
				fmt.Printf("Rows: %d\n", len(rows))

				header := rows[0]
				rows = rows[1:]

				headerMap := map[string]int{}
				for i, column := range header {
					column = strings.ToLower(column)
					column = strings.TrimSpace(column)
					headerMap[column] = i
				}

				var records []databasemodel.FSFOperatingUnitExpenditureSummary
				for r, row := range rows {
					for c := range row {
						row[c] = strings.TrimSpace(row[c])
					}
					record := databasemodel.FSFOperatingUnitExpenditureSummary{
						District:                 row[headerMap["district"]],
						Division:                 row[headerMap["div"]],
						RecordType:               row[headerMap["recordtype"]],
						SubType:                  row[headerMap["subtype"]],
						OperatingUnit:            row[headerMap["operatingunit"]],
						OperatingUnitDescription: row[headerMap["descr"]],
					}
					if row[headerMap["budgetamt"]] != "" {
						v, err := strconv.ParseFloat(row[headerMap["budgetamt"]], 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing budgetamt: %v\n", r+1, err)
							continue
						}
						record.BudgetedAmount = v
					}
					if row[headerMap["encumberedamt"]] != "" {
						v, err := strconv.ParseFloat(row[headerMap["encumberedamt"]], 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing encumberedamt: %v\n", r+1, err)
							continue
						}
						record.EncumberedAmount = v
					}
					if row[headerMap["expendedamt"]] != "" {
						v, err := strconv.ParseFloat(row[headerMap["expendedamt"]], 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing expendedamt: %v\n", r+1, err)
							continue
						}
						record.ExpendedAmount = v
					}
					records = append(records, record)
				}

				err := db.CreateInBatches(records, 100).Error
				if err != nil {
					panic(err)
				}
			}
		}
	}

	{
		filename := baseDirectory + string(filepath.Separator) + "fsf.operating-unit-program-summary.csv"
		fmt.Printf("filename: %s\n", filename)

		if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", filename)
			panic(err)
		} else {
			fileHandle, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			csvReader := csv.NewReader(fileHandle)
			rows, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}

			if len(rows) == 0 {
				fmt.Printf("No rows found in the CSV file.\n")
			} else {
				fmt.Printf("Rows: %d\n", len(rows))

				header := rows[0]
				rows = rows[1:]

				headerMap := map[string]int{}
				for i, column := range header {
					column = strings.ToLower(column)
					headerMap[column] = i
				}

				var records []databasemodel.FSFOperatingUnitProgramSummary
				for r, row := range rows {
					record := databasemodel.FSFOperatingUnitProgramSummary{
						District:                 row[headerMap["district"]],
						Division:                 row[headerMap["div"]],
						RecordType:               row[headerMap["recordtype"]],
						OperatingUnit:            row[headerMap["operatingunit"]],
						OperatingUnitDescription: row[headerMap["operatingunitdesc"]],
						ProgramCode:              row[headerMap["programcode"]],
						ProgramCodeDescription:   row[headerMap["programcodedesc"]],
					}
					if row[headerMap["budgetamt"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["budgetamt"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing budgetamt: %v\n", r+1, err)
							continue
						}
						record.BudgetedAmount = v
					}
					if row[headerMap["encumberedamt"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["encumberedamt"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing encumberedamt: %v\n", r+1, err)
							continue
						}
						record.EncumberedAmount = v
					}
					if row[headerMap["expendedamt"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["expendedamt"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing expendedamt: %v\n", r+1, err)
							continue
						}
						record.ExpendedAmount = v
					}

					records = append(records, record)
				}

				err := db.CreateInBatches(records, 100).Error
				if err != nil {
					panic(err)
				}
			}
		}
	}

	files, err := filepath.Glob(baseDirectory + string(filepath.Separator) + "mobius.DGL060.*.csv")
	if err != nil {
		panic(err)
	}
	for _, filename := range files {
		division := strings.TrimSuffix(strings.TrimPrefix(filename, baseDirectory+"/mobius.DGL060."), ".csv")
		fmt.Printf("filename: %s\n", filename)

		if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", filename)
			panic(err)
		} else {
			fileHandle, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			csvReader := csv.NewReader(fileHandle)
			rows, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}

			if len(rows) == 0 {
				fmt.Printf("No rows found in the CSV file.\n")
			} else {
				fmt.Printf("Rows: %d\n", len(rows))

				header := rows[0]
				rows = rows[1:]

				headerMap := map[string]int{}
				for i, column := range header {
					column = strings.ToLower(column)
					headerMap[column] = i
				}

				var records []databasemodel.MobiusDGL060
				for r, row := range rows {
					row = processFormulas(row)

					record := databasemodel.MobiusDGL060{
						Division:                 division,
						DepartmentID:             row[headerMap["dept_id"]],
						DepartmentDescription:    row[headerMap["dept_desc"]],
						Fund:                     row[headerMap["fund"]],
						Appropriation:            row[headerMap["appr"]],
						AppropriationType:        row[headerMap["type"]],
						AppropriationDescription: row[headerMap["appr_descr"]],
					}
					{
						v, err := strconv.ParseInt(row[headerMap["fy"]], 10, 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing fy: %v\n", r+1, err)
							continue
						}
						if v < 100 {
							v += 2000
						}
						record.FiscalYear = int(v)
					}
					{
						v, err := time.Parse("01/02/06", row[headerMap["rpt_asof_date"]])
						if err != nil {
							fmt.Printf("Row %d: error parsing rpt_asof_date: %v\n", r+1, err)
							continue
						}
						record.AsOfDate = v
					}
					{
						v, err := time.Parse("01/02/06", row[headerMap["end_date"]])
						if err != nil {
							fmt.Printf("Row %d: error parsing end_date: %v\n", r+1, err)
							continue
						}
						record.EndDate = v
					}
					if row[headerMap["available_funds"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["available_funds"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing available_funds: %v\n", r+1, err)
							continue
						}
						record.AvailableAmount = v
					}
					if row[headerMap["encumbrances"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["encumbrances"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing encumbrances: %v\n", r+1, err)
							continue
						}
						record.EncumberedAmount = v
					}
					if row[headerMap["curr_yr_expen"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["curr_yr_expen"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing curr_yr_expen: %v\n", r+1, err)
							continue
						}
						record.CurrentYearExpenses = v
					}
					if row[headerMap["prior_yr_expen"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["prior_yr_expen"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing prior_yr_expen: %v\n", r+1, err)
							continue
						}
						record.PriorYearExpenses = v
					}
					if row[headerMap["remain_spend_auth"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["remain_spend_auth"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing remain_spend_auth: %v\n", r+1, err)
							continue
						}
						record.RemainingSpendAuthorized = v
					}

					records = append(records, record)
				}

				err := db.CreateInBatches(records, 100).Error
				if err != nil {
					panic(err)
				}
			}
		}
	}

	files, err = filepath.Glob(baseDirectory + string(filepath.Separator) + "mobius.DGL114.*.csv")
	if err != nil {
		panic(err)
	}
	for _, filename := range files {
		division := strings.TrimSuffix(strings.TrimPrefix(filename, baseDirectory+"/mobius.DGL114."), ".csv")
		fmt.Printf("filename: %s\n", filename)

		if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", filename)
			panic(err)
		} else {
			fileHandle, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			csvReader := csv.NewReader(fileHandle)
			rows, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}

			if len(rows) == 0 {
				fmt.Printf("No rows found in the CSV file.\n")
			} else {
				fmt.Printf("Rows: %d\n", len(rows))

				header := rows[0]
				rows = rows[1:]

				headerMap := map[string]int{}
				for i, column := range header {
					column = strings.ToLower(column)
					headerMap[column] = i
				}

				var records []databasemodel.MobiusDGL114
				for r, row := range rows {
					row = processFormulas(row)

					record := databasemodel.MobiusDGL114{
						Division:                  division,
						DepartmentID:              row[headerMap["deptid"]],
						DepartmentDescription:     row[headerMap["deptdesc"]],
						Fund:                      row[headerMap["fund"]],
						Appropriation:             row[headerMap["apprcode"]],
						AppropriationType:         row[headerMap["apprtype"]],
						RevenueAccount:            row[headerMap["revaccount"]],
						RevenueAccountDescription: row[headerMap["revdescr"]],
					}
					{
						v, err := strconv.ParseInt(row[headerMap["budref"]], 10, 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing budref: %v\n", r+1, err)
							continue
						}
						if v < 100 {
							v += 2000
						}
						record.BudgetYear = int(v)
					}
					{
						v, err := time.Parse("01/02/2006", row[headerMap["rptasofdate"]])
						if err != nil {
							fmt.Printf("Row %d: error parsing rptasofdate: %v\n", r+1, err)
							continue
						}
						record.AsOfDate = v
					}
					if row[headerMap["gf_current"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["gf_current"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing gf_current: %v\n", r+1, err)
							continue
						}
						record.LocalFundsCurrent = v
					}
					if row[headerMap["gf_ytd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["gf_ytd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing gf_ytd: %v\n", r+1, err)
							continue
						}
						record.LocalFundsYearToDate = v
					}
					if row[headerMap["sf_current"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["sf_current"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing sf_current: %v\n", r+1, err)
							continue
						}
						record.StateFundsCurrent = v
					}
					if row[headerMap["sf_ytd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["sf_ytd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing sf_ytd: %v\n", r+1, err)
							continue
						}
						record.StateFundsYearToDate = v
					}

					records = append(records, record)
				}

				err := db.CreateInBatches(records, 100).Error
				if err != nil {
					panic(err)
				}
			}
		}
	}

	files, err = filepath.Glob(baseDirectory + string(filepath.Separator) + "mobius.DGL115.*.csv")
	if err != nil {
		panic(err)
	}
	for _, filename := range files {
		division := strings.TrimSuffix(strings.TrimPrefix(filename, baseDirectory+"/mobius.DGL115."), ".csv")
		fmt.Printf("filename: %s\n", filename)

		if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", filename)
			panic(err)
		} else {
			fileHandle, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			csvReader := csv.NewReader(fileHandle)
			rows, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}

			if len(rows) == 0 {
				fmt.Printf("No rows found in the CSV file.\n")
			} else {
				fmt.Printf("Rows: %d\n", len(rows))

				header := rows[0]
				rows = rows[1:]

				headerMap := map[string]int{}
				for i, column := range header {
					column = strings.ToLower(column)
					headerMap[column] = i
				}

				var records []databasemodel.MobiusDGL115
				for r, row := range rows {
					row = processFormulas(row)

					record := databasemodel.MobiusDGL115{
						Division:              division,
						DepartmentID:          row[headerMap["deptid"]],
						DepartmentDescription: row[headerMap["dept_descr"]],
						Account:               row[headerMap["account"]],
						AccountDescription:    row[headerMap["acct_descr"]],
					}
					{
						v, err := strconv.ParseInt(row[headerMap["fy"]], 10, 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing fy: %v\n", r+1, err)
							continue
						}
						if v < 100 {
							v += 2000
						}
						record.FiscalYear = int(v)
					}
					{
						v, err := strconv.ParseInt(row[headerMap["acct_period"]], 10, 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing acct_period: %v\n", r+1, err)
							continue
						}
						record.AccountPeriod = int(v)
					}
					if row[headerMap["gf_mtd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["gf_mtd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing gf_mtd: %v\n", r+1, err)
							continue
						}
						record.LocalFundsMonthToDate = v
					}
					if row[headerMap["sf_mtd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["sf_mtd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing sf_mtd: %v\n", r+1, err)
							continue
						}
						record.StateFundsMonthToDate = v
					}
					if row[headerMap["totl_mtd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["totl_mtd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing totl_mtd: %v\n", r+1, err)
							continue
						}
						record.TotalFundsMonthToDate = v
					}
					if row[headerMap["gf_ytd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["gf_ytd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing gf_ytd: %v\n", r+1, err)
							continue
						}
						record.LocalFundsYearToDate = v
					}
					if row[headerMap["sf_ytd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["sf_ytd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing sf_ytd: %v\n", r+1, err)
							continue
						}
						record.StateFundsYearToDate = v
					}
					if row[headerMap["totl_ytd"]] != "" {
						v, err := strconv.ParseFloat(strings.ReplaceAll(row[headerMap["totl_ytd"]], ",", ""), 64)
						if err != nil {
							fmt.Printf("Row %d: error parsing totl_ytd: %v\n", r+1, err)
							continue
						}
						record.TotalFundsYearToDate = v
					}

					records = append(records, record)
				}

				err := db.CreateInBatches(records, 100).Error
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func processFormulas(row []string) []string {
	for c, column := range row {
		column = strings.TrimSpace(column)
		if strings.HasPrefix(column, "=") {
			column = strings.TrimPrefix(column, "=")
			if strings.HasPrefix(column, `"`) && strings.HasSuffix(column, `"`) {
				column = strings.TrimPrefix(column, `"`)
				column = strings.TrimSuffix(column, `"`)
			}
			column = strings.TrimSpace(column)
		}
		if strings.HasPrefix(column, `(`) && strings.HasSuffix(column, `)`) {
			column = strings.TrimPrefix(column, `(`)
			column = strings.TrimSuffix(column, `)`)
			column = "-" + column
		}
		row[c] = column
	}
	return row
}
