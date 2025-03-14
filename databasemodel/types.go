package databasemodel

import "time"

type FSFOperatingUnitExpenditureSummary struct {
	District                 string  `gorm:"column:district"`
	Division                 string  `gorm:"column:division"`
	RecordType               string  `gorm:"column:record_type"`
	SubType                  string  `gorm:"column:sub_type"`
	OperatingUnit            string  `gorm:"column:operating_unit"`
	OperatingUnitDescription string  `gorm:"column:operating_unit_description"`
	BudgetedAmount           float64 `gorm:"column:budget_amount"`
	EncumberedAmount         float64 `gorm:"column:encumbered_amount"`
	ExpendedAmount           float64 `gorm:"column:expended_amount"`
}

type FSFOperatingUnitProgramSummary struct {
	District                 string  `gorm:"column:district"`
	Division                 string  `gorm:"column:division"`
	RecordType               string  `gorm:"column:record_type"`
	OperatingUnit            string  `gorm:"column:operating_unit"`
	OperatingUnitDescription string  `gorm:"column:operating_unit_description"`
	ProgramCode              string  `gorm:"column:program_code"`
	ProgramCodeDescription   string  `gorm:"column:program_code_description"`
	BudgetedAmount           float64 `gorm:"column:budget_amount"`
	EncumberedAmount         float64 `gorm:"column:encumbered_amount"`
	ExpendedAmount           float64 `gorm:"column:expended_amount"`
}

type MobiusDGL060 struct {
	Division                 string    `gorm:"column:division"`
	AsOfDate                 time.Time `gorm:"column:as_of_date"`
	DepartmentID             string    `gorm:"column:department_id"`
	DepartmentDescription    string    `gorm:"column:department_description"`
	FiscalYear               int       `gorm:"column:fiscal_year"`
	Fund                     string    `gorm:"column:fund"`
	Appropriation            string    `gorm:"column:appropriation"`
	AppropriationType        string    `gorm:"column:appropriation_type"`
	AppropriationDescription string    `gorm:"column:appropriation_description"`
	EndDate                  time.Time `gorm:"column:end_date"`
	AvailableAmount          float64   `gorm:"column:available_amount"` // This is the total amount of money available.
	EncumberedAmount         float64   `gorm:"column:encumbered_amount"`
	CurrentYearExpenses      float64   `gorm:"column:current_year_expenses"`
	PriorYearExpenses        float64   `gorm:"column:prior_year_expenses"`
	RemainingAmount          float64   `gorm:"column:remaining_spend_authorized"`
}

type MobiusDGL114 struct {
	Division                  string    `gorm:"column:division"`
	AsOfDate                  time.Time `gorm:"column:as_of_date"`
	DepartmentID              string    `gorm:"column:department_id"`
	DepartmentDescription     string    `gorm:"column:department_description"`
	BudgetYear                int       `gorm:"column:budget_year"`
	Fund                      string    `gorm:"column:fund"`
	Appropriation             string    `gorm:"column:appropriation"`
	AppropriationType         string    `gorm:"column:appropriation_type"`
	RevenueAccount            string    `gorm:"column:revenue_account"`
	RevenueAccountDescription string    `gorm:"column:revenue_account_description"`
	LocalFundsCurrent         float64   `gorm:"column:local_funds_current"`
	LocalFundsYearToDate      float64   `gorm:"column:local_funds_year_to_date"`
	StateFundsCurrent         float64   `gorm:"column:state_funds_current"`
	StateFundsYearToDate      float64   `gorm:"column:state_funds_year_to_date"`
}

type MobiusDGL115 struct {
	Division              string  `gorm:"column:division"`
	DepartmentID          string  `gorm:"column:department_id"`
	DepartmentDescription string  `gorm:"column:department_description"`
	FiscalYear            int     `gorm:"column:fiscal_year"`
	AccountPeriod         int     `gorm:"column:account_period"`
	Account               string  `gorm:"column:account"`
	AccountDescription    string  `gorm:"column:account_description"`
	LocalFundsMonthToDate float64 `gorm:"column:local_funds_month_to_date"`
	StateFundsMonthToDate float64 `gorm:"column:state_funds_month_to_date"`
	TotalFundsMonthToDate float64 `gorm:"column:total_funds_month_to_date"`
	LocalFundsYearToDate  float64 `gorm:"column:local_funds_year_to_date"`
	StateFundsYearToDate  float64 `gorm:"column:state_funds_year_to_date"`
	TotalFundsYearToDate  float64 `gorm:"column:total_funds_year_to_date"`
}
