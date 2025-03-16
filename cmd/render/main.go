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
.money {
	text-align: right;
	font-family: monospace;
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
<h1>Budget Overview</h1>
<table>
	<thead>
		<tr>
			<th>Division</th>
			<th>Department</th>
			<th>Budget Amount</th>
			<th style="min-width: 20em;">Usage</th>
			<th>Available</th>
		</tr>
	</thead>
	<tbody>
{{ range . }}
		<tr>
			<td>{{ .Division }}</td>
			<td>{{ .DepartmentDescription }}</td>
			<td><div class="money">{{ formatMoney .BudgetAmount }}</div></td>
			<td><div style="width: 100%;" class="budget-bar"><div class="expended" style="width: {{ div ( mul 100 .ExpendedAmount ) .BudgetAmount }}%;"></div><div class="encumbered" style="width: {{ div ( mul 100.0 .EncumberedAmount ) .BudgetAmount }}%;"></div><div class="available" style="flex: 1;"></div></td>
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

		allHTML += w.String()
	}

	allHTML += "</body>"
	allHTML += "</html>"

	err = os.WriteFile(outputDirectory+string(filepath.Separator)+"report.html", []byte(allHTML), 0644)
	if err != nil {
		panic(err)
	}
}
