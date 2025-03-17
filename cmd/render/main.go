package main

import (
	"bytes"
	"flag"
	"html/template"
	"os"
	"path/filepath"

	"github.com/tekkamanendless/cboc-tools/database"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

	var allHTML string
	allHTML += "<html>"
	allHTML += "<head>"
	allHTML += "<title>CBOC Report</title>"
	allHTML += `<style>
body {
	font-family: sans-serif;
}
@media screen {
	.page {
		padding-bottom: 5em;
		border-bottom: 1px solid gray;
		margin-bottom: 5em;
	}
}
@media print {
	body {
		font-size: 0.9em;
	}
	td, th {
		font-size: 0.9em;
	}
	.page {
		break-after: page;
	}
	.page-break {
		break-after: page;
	}
}
a:visited, a:active {
	color: blue;
}
.money {
	text-align: right;
	font-variant-numeric: ordinal;
	font-size: 0.9em;
}
.budget-bar {
	display: flex;
	min-height: 1em;
	height: 1em;
	background-color: #f0f0f0;
}
.budget-bar div {
	min-height: 1em;
	height: 1em;
}
.expended {
	background-color: #ff0000;
}
.encumbered {
	background-color: #ff8800;
}
.available {
	background-color: #88ff88;
}
</style>`
	allHTML += "</head>"
	allHTML += "<body>"
	allHTML += "<h1>CBOC Report</h1>"

	funcMap := template.FuncMap{
		"add": func(inputs ...float64) float64 {
			if len(inputs) == 0 {
				return 0
			}
			output := inputs[0]
			for i := 1; i < len(inputs); i++ {
				output += inputs[i]
			}
			return output
		},
		"div": func(inputs ...float64) float64 {
			if len(inputs) == 0 {
				return 0
			}
			output := inputs[0]
			for i := 1; i < len(inputs); i++ {
				output /= inputs[i]
			}
			return output
		},
		"formatMoney": func(amount float64) string {
			printer := message.NewPrinter(language.English)
			return "$" + printer.Sprintf("%0.2f", amount)
		},
		"mul": func(inputs ...float64) float64 {
			if len(inputs) == 0 {
				return 0
			}
			output := inputs[0]
			for i := 1; i < len(inputs); i++ {
				output *= inputs[i]
			}
			return output
		},
		"sub": func(inputs ...float64) float64 {
			if len(inputs) == 0 {
				return 0
			}
			output := inputs[0]
			for i := 1; i < len(inputs); i++ {
				output -= inputs[i]
			}
			return output
		},
	}

	{
		allHTML += `<div class="page">`
		allHTML += "<h1>Table of Contents</h1>"
		allHTML += "<ul>"
		allHTML += "<li><a href=\"#budget-overview\">Budget Overview</a></li>"
		allHTML += "<li><a href=\"#budget-breakdown\">Budget Breakdown</a></li>"
		allHTML += "<li><a href=\"#program-breakdown\">Program Breakdown</a></li>"
		allHTML += "</ul>"
		allHTML += `</div>`
	}

	{
		type Row struct {
			Division              string  `gorm:"column:division"`
			DepartmentDescription string  `gorm:"column:department_description"`
			BudgetAmount          float64 `gorm:"column:budget_amount"`
			EncumberedAmount      float64 `gorm:"column:encumbered_amount"`
			ExpendedAmount        float64 `gorm:"column:expended_amount"`
		}
		var rows []Row
		err := db.Raw(`
SELECT
	report.division,
	department.department_description,
	SUM(budget_amount) AS budget_amount,
	SUM(encumbered_amount) AS encumbered_amount,
	SUM(expended_amount) AS expended_amount
FROM
	fsf_operating_unit_expenditure_summaries AS report
	INNER JOIN
	(
		SELECT DISTINCT division, department_description FROM mobius_dgl115
	) AS department
		ON report.division = department.division
GROUP BY
	report.division
HAVING
	budget_amount > 0
`).
			Find(&rows).
			Error
		if err != nil {
			panic(err)
		}

		templateText := `
<a name="budget-overview">
<h1>Budget Overview</h1>
<table width="100%">
	<thead>
		<tr>
			<th width="40%">Department</th>
			<th width="10%">Budget Amount</th>
			<th width="40%">Usage</th>
			<th width="10%">Available</th>
		</tr>
	</thead>
	<tbody>
{{ range . }}
		<tr>
			<td><a href="#budget-breakdown-{{ .Division }}">{{ .Division }} - {{ .DepartmentDescription }}</a></td>
			<td><div class="money">{{ formatMoney .BudgetAmount }}</div></td>
			<td><div style="width: 100%;" class="budget-bar"><div class="expended" style="width: {{ div ( mul 100 .ExpendedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .ExpendedAmount }}"></div><div class="encumbered" style="width: {{ div ( mul 100.0 .EncumberedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .EncumberedAmount }}"></div><div class="available" style="flex: 1;" title="{{ formatMoney (sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}"></div></td>
			<td><div class="money">{{ formatMoney ( sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}</div></td>
		</tr>
{{ end }}
	</tbody>
</table>
                `
		t, err := template.New("").Funcs(funcMap).Parse(templateText)
		if err != nil {
			panic(err)
		}

		var w bytes.Buffer
		err = t.Execute(&w, rows)
		if err != nil {
			panic(err)
		}

		allHTML += `<div class="page">`
		allHTML += w.String()
		allHTML += `</div>`
	}

	{
		type Line struct {
			ProgramCode            string
			ProgramCodeDescription string
			BudgetAmount           float64
			EncumberedAmount       float64
			ExpendedAmount         float64
		}

		type Unit struct {
			UnitCode        string
			UnitDescription string
			Lines           []*Line

			BudgetAmount     float64
			EncumberedAmount float64
			ExpendedAmount   float64
		}

		type Division struct {
			Division    string
			Description string
			Units       []*Unit

			BudgetAmount     float64
			EncumberedAmount float64
			ExpendedAmount   float64
		}

		type Row struct {
			Division               string  `gorm:"column:division"`
			DepartmentDescription  string  `gorm:"column:department_description"`
			UnitCode               string  `gorm:"column:operating_unit"`
			UnitDescription        string  `gorm:"column:operating_unit_description"`
			ProgramCode            string  `gorm:"column:program_code"`
			ProgramCodeDescription string  `gorm:"column:program_code_description"`
			BudgetAmount           float64 `gorm:"column:budget_amount"`
			EncumberedAmount       float64 `gorm:"column:encumbered_amount"`
			ExpendedAmount         float64 `gorm:"column:expended_amount"`
		}
		var rows []Row
		err := db.Raw(`
SELECT
	report.division,
	department.department_description,
	operating_unit,
	operating_unit_description,
	program_code,
	program_code_description,
	SUM(budget_amount) AS budget_amount,
	SUM(encumbered_amount) AS encumbered_amount,
	SUM(expended_amount) AS expended_amount
FROM
	fsf_operating_unit_program_summaries AS report
	INNER JOIN
	(
		SELECT DISTINCT division, department_description FROM mobius_dgl115
	) AS department
		ON report.division = department.division
GROUP BY
	report.division, operating_unit, program_code
HAVING
	budget_amount > 0
`).
			Find(&rows).
			Error
		if err != nil {
			panic(err)
		}

		divisionMap := map[string]*Division{}
		unitMap := map[string]map[string]*Unit{}
		divisions := []*Division{}
		for _, row := range rows {
			division, ok := divisionMap[row.Division]
			if !ok {
				division = &Division{
					Division:    row.Division,
					Description: row.DepartmentDescription,
				}
				divisionMap[row.Division] = division
				divisions = append(divisions, division)
			}

			if _, ok := unitMap[row.Division]; !ok {
				unitMap[row.Division] = map[string]*Unit{}
			}
			unit, ok := unitMap[row.Division][row.UnitCode]
			if !ok {
				unit = &Unit{
					UnitCode:        row.UnitCode,
					UnitDescription: row.UnitDescription,
				}
				unitMap[row.Division][row.UnitCode] = unit
				division.Units = append(division.Units, unit)
			}

			line := &Line{
				ProgramCode:            row.ProgramCode,
				ProgramCodeDescription: row.ProgramCodeDescription,
				BudgetAmount:           row.BudgetAmount,
				EncumberedAmount:       row.EncumberedAmount,
				ExpendedAmount:         row.ExpendedAmount,
			}
			unit.Lines = append(unit.Lines, line)

			unit.BudgetAmount += row.BudgetAmount
			unit.EncumberedAmount += row.EncumberedAmount
			unit.ExpendedAmount += row.ExpendedAmount

			division.BudgetAmount += row.BudgetAmount
			division.EncumberedAmount += row.EncumberedAmount
			division.ExpendedAmount += row.ExpendedAmount
		}

		templateText := `
<a name="budget-breakdown">
<h1>Budget Breakdown</h1>
{{ range . }}
{{ $division := .}}
<div class="page">
<a name="budget-breakdown-{{ .Division }}">
<h2>{{ .Division }} - {{ .Description }}</h2>
<table width="100%">
	<thead>
		<tr>
			<th width="10%">Code</th>
			<th width="30%">Program</th>
			<th width="10%">Budget Amount</th>
			<th width="40%">Usage</th>
			<th width="10%">Available</th>
		</tr>
	</thead>
	<tbody>
{{ range .Units }}
		<tr>
			<td><a href="#budget-breakdown-{{ $division.Division }}-unit-{{ .UnitCode }}">{{ .UnitCode }}</a></td>
			<td><a href="#budget-breakdown-{{ $division.Division }}-unit-{{ .UnitCode }}">{{ .UnitDescription }}</a></td>
			<td><div class="money">{{ formatMoney .BudgetAmount }}</div></td>
			<td><div style="width: 100%;" class="budget-bar"><div class="expended" style="width: {{ div ( mul 100 .ExpendedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .ExpendedAmount }}"></div><div class="encumbered" style="width: {{ div ( mul 100.0 .EncumberedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .EncumberedAmount }}"></div><div class="available" style="flex: 1;" title="{{ formatMoney (sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}"></div></td>
			<td><div class="money">{{ formatMoney ( sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}</div></td>
		</tr>
{{ end }}
	</tbody>
</table>
{{ range .Units }}
 <a name="budget-breakdown-{{ $division.Division }}-unit-{{ .UnitCode }}">
<h3>{{ .UnitCode }} - {{ .UnitDescription }}</h3>
<table width="100%">
	<thead>
		<tr>
			<th width="10%">Code</th>
			<th width="30%">Program</th>
			<th width="10%">Budget Amount</th>
			<th width="40%">Usage</th>
			<th width="10%">Available</th>
		</tr>
	</thead>
	<tbody>
{{ range .Lines }}
		<tr>
			<td>{{ .ProgramCode }}</td>
			<td>{{ .ProgramCodeDescription }}</td>
			<td><div class="money">{{ formatMoney .BudgetAmount }}</div></td>
			<td><div style="width: 100%;" class="budget-bar"><div class="expended" style="width: {{ div ( mul 100 .ExpendedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .ExpendedAmount }}"></div><div class="encumbered" style="width: {{ div ( mul 100.0 .EncumberedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .EncumberedAmount }}"></div><div class="available" style="flex: 1;" title="{{ formatMoney (sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}"></div></td>
			<td><div class="money">{{ formatMoney ( sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}</div></td>
		</tr>
{{ end }}
	</tbody>
</table>
{{ end }}
</div>
{{ end }}
                `
		t, err := template.New("").Funcs(funcMap).Parse(templateText)
		if err != nil {
			panic(err)
		}

		var w bytes.Buffer
		err = t.Execute(&w, divisions)
		if err != nil {
			panic(err)
		}

		//allHTML += `<div class="page">`
		allHTML += w.String()
		//allHTML += `</div>`
	}

	{
		type Line struct {
			UnitCode            string
			UnitCodeDescription string
			BudgetAmount        float64
			EncumberedAmount    float64
			ExpendedAmount      float64
		}

		type Program struct {
			ProgramCode            string
			ProgramCodeDescription string
			Lines                  []*Line

			BudgetAmount     float64
			EncumberedAmount float64
			ExpendedAmount   float64
		}

		type Division struct {
			Division    string
			Description string
			Programs    []*Program

			BudgetAmount     float64
			EncumberedAmount float64
			ExpendedAmount   float64
		}

		type Row struct {
			Division               string  `gorm:"column:division"`
			DepartmentDescription  string  `gorm:"column:department_description"`
			UnitCode               string  `gorm:"column:operating_unit"`
			UnitDescription        string  `gorm:"column:operating_unit_description"`
			ProgramCode            string  `gorm:"column:program_code"`
			ProgramCodeDescription string  `gorm:"column:program_code_description"`
			BudgetAmount           float64 `gorm:"column:budget_amount"`
			EncumberedAmount       float64 `gorm:"column:encumbered_amount"`
			ExpendedAmount         float64 `gorm:"column:expended_amount"`
		}
		var rows []Row
		err := db.Raw(`
SELECT
	report.division,
	department.department_description,
	operating_unit,
	operating_unit_description,
	program_code,
	program_code_description,
	SUM(budget_amount) AS budget_amount,
	SUM(encumbered_amount) AS encumbered_amount,
	SUM(expended_amount) AS expended_amount
FROM
	fsf_operating_unit_program_summaries AS report
	INNER JOIN
	(
		SELECT DISTINCT division, department_description FROM mobius_dgl115
	) AS department
		ON report.division = department.division
GROUP BY
	report.division, operating_unit, program_code
HAVING
	budget_amount > 0
`).
			Find(&rows).
			Error
		if err != nil {
			panic(err)
		}

		divisionMap := map[string]*Division{}
		programMap := map[string]map[string]*Program{}
		divisions := []*Division{}
		for _, row := range rows {
			division, ok := divisionMap[row.Division]
			if !ok {
				division = &Division{
					Division:    row.Division,
					Description: row.DepartmentDescription,
				}
				divisionMap[row.Division] = division
				divisions = append(divisions, division)
			}

			if _, ok := programMap[row.Division]; !ok {
				programMap[row.Division] = map[string]*Program{}
			}
			program, ok := programMap[row.Division][row.ProgramCode]
			if !ok {
				program = &Program{
					ProgramCode:            row.ProgramCode,
					ProgramCodeDescription: row.ProgramCodeDescription,
				}
				programMap[row.Division][row.ProgramCode] = program
				division.Programs = append(division.Programs, program)
			}

			line := &Line{
				UnitCode:            row.UnitCode,
				UnitCodeDescription: row.UnitDescription,
				BudgetAmount:        row.BudgetAmount,
				EncumberedAmount:    row.EncumberedAmount,
				ExpendedAmount:      row.ExpendedAmount,
			}
			program.Lines = append(program.Lines, line)

			program.BudgetAmount += row.BudgetAmount
			program.EncumberedAmount += row.EncumberedAmount
			program.ExpendedAmount += row.ExpendedAmount

			division.BudgetAmount += row.BudgetAmount
			division.EncumberedAmount += row.EncumberedAmount
			division.ExpendedAmount += row.ExpendedAmount
		}

		templateText := `
<a name="program-breakdown">
<h1>Program Breakdown</h1>
{{ range . }}
{{ $division := .}}
<div class="page">
<a name="program-breakdown-{{ .Division }}">
<h2>{{ .Division }} - {{ .Description }}</h2>
<table width="100%">
	<thead>
		<tr>
			<th width="10%">Code</th>
			<th width="30%">Program</th>
			<th width="10%">Budget Amount</th>
			<th width="40%">Usage</th>
			<th width="10%">Available</th>
		</tr>
	</thead>
	<tbody>
{{ range .Programs }}
		<tr>
			<td><a href="#program-breakdown-{{ $division.Division }}-unit-{{ .ProgramCode }}">{{ .ProgramCode }}</a></td>
			<td><a href="#program-breakdown-{{ $division.Division }}-unit-{{ .ProgramCode }}">{{ .ProgramCodeDescription }}</a></td>
			<td><div class="money">{{ formatMoney .BudgetAmount }}</div></td>
			<td><div style="width: 100%;" class="budget-bar"><div class="expended" style="width: {{ div ( mul 100 .ExpendedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .ExpendedAmount }}"></div><div class="encumbered" style="width: {{ div ( mul 100.0 .EncumberedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .EncumberedAmount }}"></div><div class="available" style="flex: 1;" title="{{ formatMoney (sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}"></div></td>
			<td><div class="money">{{ formatMoney ( sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}</div></td>
		</tr>
{{ end }}
	</tbody>
</table>
{{ range .Programs }}
 <a name="program-breakdown-{{ $division.Division }}-unit-{{ .ProgramCode }}">
<h3>{{ .ProgramCode }} - {{ .ProgramCodeDescription }}</h3>
<table width="100%">
	<thead>
		<tr>
			<th width="10%">Code</th>
			<th width="30%">Unit</th>
			<th width="10%">Budget Amount</th>
			<th width="40%">Usage</th>
			<th width="10%">Available</th>
		</tr>
	</thead>
	<tbody>
{{ range .Lines }}
		<tr>
			<td>{{ .UnitCode }}</td>
			<td>{{ .UnitCodeDescription }}</td>
			<td><div class="money">{{ formatMoney .BudgetAmount }}</div></td>
			<td><div style="width: 100%;" class="budget-bar"><div class="expended" style="width: {{ div ( mul 100 .ExpendedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .ExpendedAmount }}"></div><div class="encumbered" style="width: {{ div ( mul 100.0 .EncumberedAmount ) .BudgetAmount }}%;" title="{{ formatMoney .EncumberedAmount }}"></div><div class="available" style="flex: 1;" title="{{ formatMoney (sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}"></div></td>
			<td><div class="money">{{ formatMoney ( sub .BudgetAmount .ExpendedAmount .EncumberedAmount ) }}</div></td>
		</tr>
{{ end }}
	</tbody>
</table>
{{ end }}
</div>
{{ end }}
                `
		t, err := template.New("").Funcs(funcMap).Parse(templateText)
		if err != nil {
			panic(err)
		}

		var w bytes.Buffer
		err = t.Execute(&w, divisions)
		if err != nil {
			panic(err)
		}

		//allHTML += `<div class="page">`
		allHTML += w.String()
		//allHTML += `</div>`
	}

	allHTML += "</body>"
	allHTML += "</html>"

	err = os.WriteFile(outputDirectory+string(filepath.Separator)+"report.html", []byte(allHTML), 0644)
	if err != nil {
		panic(err)
	}
}
