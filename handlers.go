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
		'6.SELECTION-2' , '7.TUBI' ,'9.VENEER', '', 'Used') order by area`

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
