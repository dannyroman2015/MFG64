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
	var targetUnits = map[string]string{
		"Production Value": "$", "CUTTING": "cmb", "LAMINATION": "m2", "REEDEDLINE": "m2", "VENEERLAMINATION": "m2", "PANELCNC": "sheet", "ASSEMBLY": "$", "WOODFINISHING": "$", "PACKING": "$",
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
		if d == 0 {
			laborrate = append(laborrate, 0)
		} else {
			laborrate = append(laborrate, math.Round(c/totalManhrBydate[a]))
		}
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")

		labels = append(labels, a)
		quanity = append(quanity, c)
		targets = append(targets, target)
	}

	var latestCreated string
	rows, err = s.db.Query(`select created_datetime from efficienct_reports where work_center 
		= 'PACKING' order by id desc limit 1`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		err := rows.Scan(&latestCreated)
		if err != nil {
			latestCreated = ""
			// panic(err)
		} else {
			t, err := time.Parse("2006-01-02T15:04:05.999999999Z", latestCreated)
			if err != nil {
				panic(err)
			}
			latestCreated = t.Add(time.Hour * 7).Format("15:04")
		}
	}

	return c.Render("efficiency/chart", fiber.Map{
		"workcenter":    "Production Value",
		"labels":        labels,
		"quanity":       quanity,
		"efficiency":    laborrate,
		"targets":       targets,
		"chartLabels":   []string{"Quanity", "labor rate($/manhr)", "Target"},
		"units":         units,
		"latestCreated": latestCreated,
		"targetUnits":   targetUnits,
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

	sql = `select date from reeded_reports order by date desc limit 1`
	row := s.db.QueryRow(sql)
	var latestDate string

	row.Scan(&latestDate)
	latestDate = strings.Split(latestDate, "T")[0]

	randColor := fmt.Sprintf("rgba(%d, %d, %d, 0.4)", rand.Intn(255), rand.Intn(255), rand.Intn(255))
	return c.Render("efficiency/reeded_chart", fiber.Map{
		"labels":     labels,
		"totals":     totals,
		"avgs":       avgs,
		"bg_color":   randColor,
		"latestDate": latestDate,
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
		d = math.Round(d*100) / 100
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

	for rows.Next() {
		var as = []string{"", "", "", "", "", ""}
		rows.Scan(&as[0], &as[1], &as[2], &as[3], &as[4], &as[5])
		for i := 0; i < len(as); i++ {
			if as[i] == "0" {
				as[i] = ""
			}
		}
		arr = append(arr, as)
	}

	return c.Render("efficiency/summary_body", fiber.Map{
		"arr": arr,
		// "today": time.Now().Format("02/01"),
		// "nd":    time.Now().AddDate(0, 0, 1).Format("02/01"),
		// "rd":    time.Now().AddDate(0, 0, 2).Format("02/01"),
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
		sql := `update packing_summary set plan ='` + rows[i][1] + `', 
			actual = '` + rows[i][2] + `', rh_act_pcs = '` + rows[i][3] + `', 
			rh_act_money ='` + rows[i][4] + `', m64_act_pcs = '` + rows[i][5] + `', 
			m64_act_money ='` + rows[i][6] + `' where type = '` + rows[i][0] + `'`
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

func (s *Server) qualityInputHandler(c *fiber.Ctx) error {
	sql := `select date_issue, section_code, qty_check, qty_fail from quatity_report order by
		date_issue desc, section_code limit 15`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	var data [][]string
	for rows.Next() {
		var a = make([]string, 4)
		var t string
		err := rows.Scan(&t, &a[1], &a[2], &a[3])
		if err != nil {
			panic(err)
		}
		t = strings.Split(t, "T")[0]
		a[0] = t
		data = append(data, a)
	}

	return c.Render("efficiency/qualityInput", fiber.Map{
		"data": data,
	}, "layout")
}

func (s *Server) qualityInputPostHandler(c *fiber.Ctx) error {
	section := c.FormValue("section_code")
	date_issue := c.FormValue("date_issue")
	qty_check := c.FormValue("qty_check")
	qty_fail := c.FormValue("qty_fail")
	error_note := c.FormValue("error_note")

	sql := `insert into quatity_report (section_code, date_issue, qty_check, qty_fail, error_notes)
		values ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(sql, section, date_issue, qty_check, qty_fail, error_note)
	if err != nil {
		panic(err)
	}

	return c.Redirect("/qualityinput", fiber.StatusSeeOther)
}

func (s *Server) qulityChartHandler(c *fiber.Ctx) error {
	fromdate := c.FormValue("fromdate")
	sql := `select distinct date_issue from quatity_report group by date_issue having date_issue >= '` + fromdate + `'`

	r, _ := s.db.Exec(sql)
	numberOfDate, _ := r.RowsAffected()

	sql = `select date_issue, section_code, sum(qty_check), sum(qty_fail) from quatity_report
		group by date_issue, section_code having date_issue >= '` + fromdate + `' order by date_issue`

	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	var dates = []string{}

	var data = map[string][]int{
		"M-FIN":     make([]int, numberOfDate),
		"FIN-2":     make([]int, numberOfDate),
		"FIN-1":     make([]int, numberOfDate),
		"M-WELD":    make([]int, numberOfDate),
		"UPH":       make([]int, numberOfDate),
		"ASS-1":     make([]int, numberOfDate),
		"FIT-1":     make([]int, numberOfDate),
		"ASS-2":     make([]int, numberOfDate),
		"WW-FM":     make([]int, numberOfDate),
		"PAC-2":     make([]int, numberOfDate),
		"OBA-1":     make([]int, numberOfDate),
		"WW-VN":     make([]int, numberOfDate),
		"WW-RM":     make([]int, numberOfDate),
		"PAC-1":     make([]int, numberOfDate),
		"M-MACHINE": make([]int, numberOfDate),
	}
	lastdate := ""
	i := -1
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")
		if a != lastdate {
			i++
			dates = append(dates, a)
			lastdate = a
		}
		data[b][i] = int(math.Round(d * 100 / c))
	}

	return c.Render("efficiency/quality_chart", fiber.Map{
		"dates": dates,
		"data":  data,
	})
}

func (s *Server) inputmanhrHandler(c *fiber.Ctx) error {

	return c.Render("efficiency/inputmanhr", fiber.Map{}, "layout")
}

func (s *Server) inputmanhrPostHandler(c *fiber.Ctx) error {
	inputdate := c.FormValue("inputdate")
	manhrRaw := c.FormValue("manhr")
	manhrs := strings.Split(manhrRaw, " ")

	if len(manhrs) != 8 {
		return c.SendString("Loi: phải nhập đủ 8 số cho 8 công đoạn")
	}

	wc := []string{"CUTTING", "LAMINATION", "REEDEDLINE", "VENEERLAMINATION", "PANELCNC",
		"ASSEMBLY", "WOODFINISHING", "PACKING"}

	sql := `insert into efficienct_reports(work_center, date, qty, manhr, created_datetime) values `

	for i := 0; i < 8; i++ {
		sql += `('` + wc[i] + `', '` + inputdate + `', 0, ` + manhrs[i] + `, current_timestamp),`
	}
	sql = sql[:len(sql)-1]
	_, err := s.db.Exec(sql)
	if err != nil {
		panic(err)
	}

	return c.Redirect("/efficiency", fiber.StatusSeeOther)
}

func (s *Server) qualityquickinputHandler(c *fiber.Ctx) error {
	sql := `select date_issue, section_code, qty_check, qty_fail from quatity_report order by
		date_issue desc, section_code limit 15`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	var data [][]string
	for rows.Next() {
		var a = make([]string, 4)
		var t string
		err := rows.Scan(&t, &a[1], &a[2], &a[3])
		if err != nil {
			panic(err)
		}
		t = strings.Split(t, "T")[0]
		a[0] = t
		data = append(data, a)
	}

	return c.Render("efficiency/quality_quickinput", fiber.Map{"data": data}, "layout")
}

func (s *Server) qualityquickinputPostHandler(c *fiber.Ctx) error {
	inputdate := c.FormValue("inputdate")
	raw := c.FormValue("input")
	arr := strings.Fields(raw)
	if (len(arr) % 3) != 0 {
		return c.SendString("Thiếu dữ liệu")
	}
	sql := `insert into quatity_report(section_code, date_issue, qty_check, qty_fail) values `

	for i := 0; i < len(arr); i += 3 {
		sql += `('` + arr[i] + `', '` + inputdate + `', ` + arr[i+1] + `, ` + arr[i+2] + `),`
	}
	sql = sql[:len(sql)-1]

	_, err := s.db.Exec(sql)
	if err != nil {
		panic(err)
	}

	return c.Redirect("/efficiency", fiber.StatusSeeOther)
}
