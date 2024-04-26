package main

import (
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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
	fromdate := c.Query("pfromdate")

	var labels []string
	var quanity []float64
	var laborrate []float64
	var actual_target float64
	var targets []float64
	var target float64

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
		a = t.Format("Jan 2")

		labels = append(labels, a)
		quanity = append(quanity, c)
		targets = append(targets, target)
	}

	return c.Render("efficiency/chart", fiber.Map{
		"workcenter": "prodvalue",
		"labels":     labels,
		"quanity":    quanity,
		"efficiency": laborrate,
		"targets":    targets,
	})
}
