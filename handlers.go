package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func (s *Server) efficiencyHandler(c *fiber.Ctx) error {
	fromdate := time.Now().AddDate(0, 0, -15).Format("2006-01-02")
	return c.Render("efficiency/main", fiber.Map{
		"fromdate": fromdate,
	}, "layout")
}

func (s *Server) efficiencyWithdateHandler(c *fiber.Ctx) error {
	fromdate := c.Query("fromdate")
	return c.Render("efficiency/main", fiber.Map{
		"fromdate": fromdate,
	})
}

func (s *Server) prodvalueChartHandler(c *fiber.Ctx) error {
	fromdate := c.Query("fromdate")

	var labels []string
	var quanity []float64
	var laborrate []float64
	var actual_target float64
	var targets []float64
	var target float64
	var units = map[string]string{
		"Production Value": "Amount($)",
		"CUTTING":          "Quanity(cmb)",
		"LAMINATION":       "Quanity(m2)",
		"REEDEDLINE":       "Quanity(m2)",
		"VENEERLAMINATION": "Quanity(m2)",
		"PANELCNC":         "Quanity(sheet)",
		"ASSEMBLY":         "Amount($)",
		"WOODFINISHING":    "Amount($)",
		"PACKING":          "Amount($)",
	}

	rows, err := s.db.Query(`select actual_target, target from efficienct_workcenter 
		where workcenter = 'PACKING'`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&actual_target, &target)
	}

	var totalManhrBydate = map[string]float64{}
	rows, err = s.db.Query(`SELECT date, sum(manhr) from efficienct_reports 
		group by date having date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var a string
		var b float64
		rows.Scan(&a, &b)
		a = strings.Split(a, "T")[0]
		totalManhrBydate[a] = b
	}

	rows, err = s.db.Query(`SELECT date, work_center, sum(qty), sum(manhr) from 
		efficienct_reports group by date, work_center having work_center = 'PACKING' 
		and date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		laborrate = append(laborrate, math.Round(c/totalManhrBydate[a]))
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")

		labels = append(labels, a)
		quanity = append(quanity, c)
		targets = append(targets, target)
	}
	randColor := fmt.Sprintf("rgba(%d, %d, %d, 0.4)", rand.Intn(255), rand.Intn(255), rand.Intn(255))

	return c.Render("efficiency/chart", fiber.Map{
		"workcenter":  "Production Value",
		"labels":      labels,
		"quanity":     quanity,
		"efficiency":  laborrate,
		"targets":     targets,
		"chartLabels": []string{"Quanity", "labor rate($/manhr)", "Target"},
		"bg_color":    randColor,
		"units":       units,
	})
}

func (s *Server) importreededfHangler(c *fiber.Ctx) error {

	return c.Render("efficiency/import_reeded_file", fiber.Map{}, "layout")
}

func (s *Server) proccess_reeded_excelfilePostHandler(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		panic(err)
	}
	fi, err := file.Open()
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	f, err := excelize.OpenReader(fi)
	if err != nil {
		panic(err)
	}

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		log.Println(err)
	}

	sql := `insert into reeded_reports(date, area, qty) values `

	for i := 3; i < len(rows); i++ {
		for j := 1; j < len(rows[i]); j++ {
			sql += `('` + rows[i][0] + `', '` + rows[1][j] + `', ` + rows[i][j] + `),`
		}
	}
	sql = sql[:len(sql)-1] + ` ON CONFLICT (date, area) DO UPDATE SET qty = EXCLUDED.qty;`
	_, err = s.db.Exec(sql)
	if err != nil {
		panic(err)
	}

	return c.Redirect("/efficiency", fiber.StatusFound)
}

func (s *Server) reededcahrtHandler(c *fiber.Ctx) error {
	// }
	fromdate := c.Query("fromdate")

	sql := `select area, sum(qty), avg(qty)	from reeded_reports where date >= '` + fromdate + `' 
		group by area having area in ('1.SLICE', '2.SELECTION', '3.LAMINATION', '4.DRYING', '5.REEDING' ,
		'6.SELECTION-2' , '7.TUBI' ,'9.VENEER') order by area`

	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}

	var labels []string
	var totals []float64
	var avgs []float64
	for rows.Next() {
		var a string
		var sumqty, avgaty float64
		rows.Scan(&a, &sumqty, &avgaty)
		a = strings.SplitAfter(a, ".")[1]
		labels = append(labels, a)
		totals = append(totals, math.Round(sumqty))
		avgs = append(avgs, math.Round(avgaty))
	}
	randColor := fmt.Sprintf("rgba(%d, %d, %d, 0.4)", rand.Intn(255), rand.Intn(255), rand.Intn(255))
	return c.Render("efficiency/reeded_chart", fiber.Map{
		"labels":   labels,
		"totals":   totals,
		"avgs":     avgs,
		"bg_color": randColor,
	})
}

func (s *Server) woodrecoveryHandler(c *fiber.Ctx) error {
	var dates []string
	var recoveries []float64
	var targets []float64

	sql := `select date, recovery, target from wood_recovery where date >= '` + c.FormValue("fromdate") + `' order by date`

	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println("fail to get data from wood_recovery")
		panic(err)
	}

	for rows.Next() {
		var a string
		var b, c float64
		rows.Scan(&a, &b, &c)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")
		dates = append(dates, a)
		recoveries = append(recoveries, b)
		targets = append(targets, c)
	}

	return c.Render("efficiency/wood_recover", fiber.Map{
		"dates":      dates,
		"recoveries": recoveries,
		"targets":    targets,
	})
}

func (s *Server) inputputwoodrecoveryHandler(c *fiber.Ctx) error {

	return c.Render("efficiency/input_wood_recovery", fiber.Map{}, "layout")
}

func (s *Server) inputwoodrecoveryPostHandler(c *fiber.Ctx) error {
	date := c.FormValue("inputdate")
	recovery := c.FormValue("recovery")
	target := c.FormValue("wrtarget")
	log.Println(date, recovery, target)

	sql := `insert into wood_recovery(date, recovery, target) values ($1, $2, $3)`
	_, err := s.db.Exec(sql, date, recovery, target)
	if err != nil {
		log.Println("fail to insert data into wood_recovery")
		panic(err)
	}

	return c.Redirect("/inputwoodrecovery", fiber.StatusFound)
}

func (s *Server) cuttingwhHandler(c *fiber.Ctx) error {
	date := c.FormValue("fromdate")
	var dates []string
	var qty []float64
	var wh_issue []float64
	var targets []float64

	sql := `select date, work_center, sum(qty), avg(wh_issue) from efficienct_reports group by date, work_center
		having date >= '` + date + `' and work_center = 'CUTTING' order by date`

	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println("fail to get data from efficienct_reports")
		panic(err)
	}
	for rows.Next() {
		var a string
		var b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")
		dates = append(dates, a)
		qty = append(qty, c)
		wh_issue = append(wh_issue, d)
		targets = append(targets, 28)
	}

	return c.Render("efficiency/cutting_wh", fiber.Map{
		"dates":    dates,
		"qty":      qty,
		"wh_issue": wh_issue,
		"targets":  targets,
	})
}

func (s *Server) inputwhissueHandler(c *fiber.Ctx) error {

	return c.Render("efficiency/inputwhissue", fiber.Map{}, "layout")
}

func (s *Server) inputwhissuePostHandler(c *fiber.Ctx) error {
	date := c.FormValue("inputdate")
	whissue := c.FormValue("whissue")

	sql := `update efficienct_reports set wh_issue = ` + whissue + ` where work_center = 'CUTTING' 
		and date = '` + date + `'`
	_, err := s.db.Exec(sql)
	if err != nil {
		log.Println("fail to update wh_issue to efficienct_report")
		panic(err)
	}

	return c.Redirect("inputwhissue", fiber.StatusFound)
}

func (s *Server) summarytableHandler(c *fiber.Ctx) error {
	var arr [][]string

	sql := `select plan, actual, rh_act_pcs, rh_act_money, m64_act_pcs, m64_act_money 
			from packing_summary order by stt`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}

	p := message.NewPrinter(language.English)
	for rows.Next() {
		var a = []float64{0, 0, 0, 0, 0, 0}
		var as = []string{"", "", "", "", "", ""}
		rows.Scan(&a[0], &a[1], &a[2], &a[3], &a[4], &a[5])
		for i := range a {
			if a[i] != 0 {
				as[i] = p.Sprint(a[i])
			}
		}
		arr = append(arr, as)
	}

	return c.Render("efficiency/summary_body", fiber.Map{
		"arr":   arr,
		"today": time.Now().Format("02/01"),
		"nd":    time.Now().AddDate(0, 0, 1).Format("02/01"),
		"rd":    time.Now().AddDate(0, 0, 2).Format("02/01"),
	})
}

func (s *Server) inputsummaryHandler(c *fiber.Ctx) error {

	return c.Render("efficiency/inputsummary", fiber.Map{}, "layout")
}

func (s *Server) proccessforsummaryHandler(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		panic(err)
	}
	fi, err := file.Open()
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	f, err := excelize.OpenReader(fi)
	if err != nil {
		panic(err)
	}

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		log.Println(err)
	}

	for i := 1; i < len(rows); i++ {
		sql := `update packing_summary set plan =` + rows[i][1] + `, 
			actual = ` + rows[i][2] + `, rh_act_pcs = ` + rows[i][3] + `, 
			rh_act_money =` + rows[i][4] + `, m64_act_pcs = ` + rows[i][5] + `, 
			m64_act_money =` + rows[i][6] + ` where type = '` + rows[i][0] + `'`
		_, err = s.db.Exec(sql)
		if err != nil {
			panic(err)
		}
	}

	return c.Redirect("/efficiency", fiber.StatusFound)
}

func (s *Server) evaluateHandler(c *fiber.Ctx) error {

	return c.Render("worker_quality/evaluate", fiber.Map{}, "layout")
}

func (s *Server) workerbypwPostHandler(c *fiber.Ctx) error {
	pw_department := map[string]string{
		"pw1": "Mechanical",
		"pw2": "Welding",
	}
	department, ok := pw_department[c.FormValue("pw")]
	if !ok {
		return c.SendString("department not found")
	}
	log.Println(department)

	return c.Render("worker_quality/workers_by_pw", fiber.Map{})
}

func (s *Server) searchstaffPostHandler(c *fiber.Ctx) error {
	// searchWord := c.FormValue("search")

	return c.SendString("ksdfkhsf")
}
