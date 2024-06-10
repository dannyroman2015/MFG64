package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"slices"
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
	////
	rows, err := s.db.Query(`select date, target, workers, hours from targets 
	where workcenter = 'PACKING' and date >= '` + fromdate + `' order by date`)
	if err != nil {
		log.Println(err)
		return c.SendString("Loi lay du lieu targets")
	}
	var targets []float64
	var datesOfTarget []string
	var tmp_targets []float64
	for rows.Next() {
		var a string
		var b, c, d float64
		rows.Scan(&a, &b, &c, &d)
		datesOfTarget = append(datesOfTarget, a)
		tmp_targets = append(tmp_targets, b)
		targets = append(targets, b*c*d)
	}

	rows, err = s.db.Query(`SELECT date, work_center, sum(qty), sum(manhr) from 
		efficienct_reports group by date, work_center having work_center = 'PACKING' 
		and date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}

	var efficiency []float64
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		i := slices.Index(datesOfTarget, a)
		if d == 0 || i == -1 {
			efficiency = append(efficiency, 0)
		} else {
			efficiency = append(efficiency, math.Round((c/d)*100/tmp_targets[i]))
		}
	}
	////

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
	}

	numberOfTargets := len(targets)
	var rhlist1 = make([]float64, numberOfTargets)
	var rhlist2 = make([]float64, numberOfTargets)
	var brandlist1 = make([]float64, numberOfTargets)
	var brandlist2 = make([]float64, numberOfTargets)
	var outsourcelist1 = make([]float64, numberOfTargets)
	var outsourcelist2 = make([]float64, numberOfTargets)

	rows, err = s.db.Query(`SELECT date, factory_no, type, sum(qty) from 
		efficienct_reports where work_center = 'PACKING' group by date, factory_no, type having 
		date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}
	ld := ""
	i := -1
	for rows.Next() {
		var a, b, c string
		var d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		if ld != a {
			i++
			ld = a
		}
		if b == "1" && c == "RH" {
			rhlist1[i] = d
		}
		if b == "1" && c == "BRAND" {
			brandlist1[i] = d
		}
		if b == "2" && c == "BRAND" {
			brandlist2[i] = d
		}
		if b == "2" && c == "RH" {
			rhlist2[i] = d
		}
		if b == "1" && c == "Outsource" {
			outsourcelist1[i] = d
		}
		if b == "2" && c == "Outsource" {
			outsourcelist2[i] = d
		}
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

	return c.Render("efficiency/provalue_chart", fiber.Map{
		"workcenter":     "Production Value",
		"labels":         labels,
		"quanity":        quanity,
		"efficiency":     laborrate,
		"targets":        targets,
		"chartLabels":    []string{"Quanity", "labor rate($/manhr)", "Target"},
		"units":          units,
		"latestCreated":  latestCreated,
		"targetUnits":    targetUnits,
		"rhlist1":        rhlist1,
		"brandlist1":     brandlist1,
		"rhlist2":        rhlist2,
		"brandlist2":     brandlist2,
		"outsourcelist1": outsourcelist1,
		"outsourcelist2": outsourcelist2,
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

	sql := `select date, avg(target) from wood_recovery group by date having date >= '` + c.FormValue("fromdate") + `' order by date`
	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println("fail to get data from wood_recovery")
		c.SendString("loi truy xuat")
	}
	for rows.Next() {
		var a string
		var b float64
		rows.Scan(&a, &b)
		dates = append(dates, a)
		targets = append(targets, b)
	}
	nod := len(dates)

	sql = `select date, type, recovery from wood_recovery where date >= '` + c.FormValue("fromdate") + `' order by date`

	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println("fail to get data from wood_recovery")
		c.SendString("loi truy xuat")
	}

	var rhlist = make([]float64, nod)
	var brandlist = make([]float64, nod)
	i := -1
	ld := ""
	for rows.Next() {
		var a, b string
		var c float64
		rows.Scan(&a, &b, &c)
		if ld != a {
			i++
			ld = a
		}
		if b == "RH" {
			rhlist[i] = c
		}
		if b == "BRAND" {
			brandlist[i] = c
		}
	}

	for j := 0; j < len(dates); j++ {
		t := strings.Split(dates[j], "T")[0]
		dt, _ := time.Parse("2006-01-02", t)
		dates[j] = dt.Format("2 Jan")
	}

	return c.Render("efficiency/wood_recover", fiber.Map{
		"dates":      dates,
		"recoveries": recoveries,
		"targets":    targets,
		"rhlist":     rhlist,
		"brandlist":  brandlist,
	})
}

func (s *Server) inputputwoodrecoveryHandler(c *fiber.Ctx) error {

	return c.Render("efficiency/input_wood_recovery", fiber.Map{}, "layout")
}

func (s *Server) inputwoodrecoveryPostHandler(c *fiber.Ctx) error {
	date := c.FormValue("inputdate")
	recovery := c.FormValue("recovery")
	target := c.FormValue("wrtarget")
	recoveryType := c.FormValue("recoveryType")

	sql := `insert into wood_recovery(date, recovery, target, type) values ($1, $2, $3, $4)`
	_, err := s.db.Exec(sql, date, recovery, target, recoveryType)
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
	curmon := time.Now().Format("01")
	nextmon := time.Now().AddDate(0, 1, 0).Format("01")

	sql := `select type, sum(qty), sum(pcs) from efficienct_reports where date >= '2024-` + curmon + `-01' and date < '2024-` + nextmon + `-01'
		 group by work_center, type having work_center = 'PACKING' order by type`

	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("Lỗi lấy dữ liệu")
	}
	var moneys = make([]float64, 2)
	var pcs = make([]int, 2)
	var totalm float64

	for rows.Next() {
		var a string
		var b float64
		var c int
		rows.Scan(&a, &b, &c)
		if a == "RH" {
			moneys[1] = b
			pcs[1] = c
		}
		if a == "BRAND" {
			moneys[0] = b
			pcs[0] = c
		}
		totalm += b
	}

	days := time.Now().Day()
	mtdavg := totalm / float64(days)
	rhmtdavgp := pcs[1] / days
	rhmtdavgm := moneys[1] / float64(days)
	brandavgp := pcs[0] / days
	brandavgm := moneys[0] / float64(days)
	rhmtdm := moneys[1]
	brandmtdm := moneys[0]

	nextdays := time.Since(time.Date(2024, time.Now().Month()+1, 1, 0, 0, 0, 0, time.Local))
	daystill := nextdays.Hours() / -24
	totales := math.Round(mtdavg*daystill + totalm)

	// var arr [][]string

	// sql := `select plan, actual, rh_act_pcs, rh_act_money, m64_act_pcs, m64_act_money
	// 		from packing_summary order by stt`
	// rows, err := s.db.Query(sql)
	// if err != nil {
	// 	panic(err)
	// }

	// for rows.Next() {
	// 	var as = []string{"", "", "", "", "", ""}
	// 	rows.Scan(&as[0], &as[1], &as[2], &as[3], &as[4], &as[5])
	// 	for i := 0; i < len(as); i++ {
	// 		if as[i] == "0" {
	// 			as[i] = ""
	// 		}
	// 	}
	// 	arr = append(arr, as)
	// }
	p := message.NewPrinter(message.MatchLanguage("en"))

	return c.Render("efficiency/summary_body", fiber.Map{
		// "arr": arr,
		"totalm":    p.Sprintf("%.f", totalm),
		"pcs":       pcs,
		"rhmtdm":    p.Sprintf("%.f", rhmtdm),
		"brandmtdm": p.Sprintf("%.f", brandmtdm),
		"days":      days,
		"mtdavg":    p.Sprintf("%.f", mtdavg),
		"rhmtdavgp": rhmtdavgp,
		"rhmtdavgm": p.Sprintf("%.f", rhmtdavgm),
		"brandavgp": brandavgp,
		"brandavgm": p.Sprintf("%.f", brandavgm),
		"totales":   p.Sprintf("%.f", totales),
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
	var checkqty = []float64{}
	var failqty = []float64{}
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
		checkqty = append(checkqty, c)
		failqty = append(failqty, d)
	}
	var colors = map[string]string{
		"M-FIN":     "#54bebe",
		"FIN-2":     "#A0DEFF",
		"FIN-1":     "#FF6500",
		"M-WELD":    "#badbdb",
		"UPH":       "#dedad2",
		"ASS-1":     "#e4bcad",
		"FIT-1":     "#df979e",
		"ASS-2":     "#d7658b",
		"WW-FM":     "#c80064",
		"PAC-2":     "#ffb400",
		"OBA-1":     "#527853",
		"WW-VN":     "#e1a692",
		"WW-RM":     "#9CAFAA",
		"PAC-1":     "#9080ff",
		"M-MACHINE": "#22a7f0",
	}
	return c.Render("efficiency/quality_chart", fiber.Map{
		"dates":    dates,
		"data":     data,
		"colors":   colors,
		"checkqty": checkqty,
		"failqty":  failqty,
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

func (s *Server) targetHandler(c *fiber.Ctx) error {
	units := map[string]string{
		"CUTTING":          "CBM",
		"LAMINATION":       "M2",
		"REEDEDLINE":       "M2",
		"VENEERLAMINATION": "M2",
		"PANELCNC":         "SHEET",
		"ASSEMBLY":         "$",
		"WOODFINISHING":    "$",
		"PACKING":          "$",
	}

	return c.Render("efficiency/inputtarget", fiber.Map{
		"msg":   "Form for setting Targets, choose dates to review updating history",
		"units": units,
	}, "layout")
}

func (s *Server) targetPostHandler(c *fiber.Ctx) error {
	wc := c.FormValue("workcenter")
	if wc == "" {
		return c.SendString("Please choose workcenter!")
	}

	rawDate := c.FormValue("dateRange")
	if rawDate == "" {
		return c.SendString("Please choose date range")
	}
	dateRange := strings.Split(rawDate, " - ")
	startDate, _ := time.Parse("2006-01-02", dateRange[0])
	endDate, _ := time.Parse("2006-01-02", dateRange[1])

	workers := c.FormValue("workers")
	hours := c.FormValue("hours")
	target := c.FormValue("target")
	if target == "" || workers == "" || hours == "" {
		return c.SendString("Please choose target, numbers of workers and working hours")
	}

	weekdays := []string{}
	if c.FormValue("Monday") != "" {
		weekdays = append(weekdays, c.FormValue("Monday"))
	}
	if c.FormValue("Tuesday") != "" {
		weekdays = append(weekdays, c.FormValue("Tuesday"))
	}
	if c.FormValue("Wednesday") != "" {
		weekdays = append(weekdays, c.FormValue("Wednesday"))
	}
	if c.FormValue("Thursday") != "" {
		weekdays = append(weekdays, c.FormValue("Thursday"))
	}
	if c.FormValue("Friday") != "" {
		weekdays = append(weekdays, c.FormValue("Friday"))
	}
	if c.FormValue("Saturday") != "" {
		weekdays = append(weekdays, c.FormValue("Saturday"))
	}
	if c.FormValue("Sunday") != "" {
		weekdays = append(weekdays, c.FormValue("Sunday"))
	}

	units := map[string]string{
		"CUTTING":          "m³/h",
		"LAMINATION":       "m²/h",
		"REEDEDLINE":       "m²/h",
		"VENEERLAMINATION": "m²/h",
		"PANELCNC":         "sheets/h",
		"ASSEMBLY":         "$/h",
		"WOODFINISHING":    "$/h",
		"PACKING":          "$/h",
	}
	unit := units[wc]
	var sql string
	demand := c.FormValue("demandofmonth")
	if demand == "" || demand == "0" {
		sql = `insert into targets(workcenter, date, target, unit, workers, hours) values `
		for i := startDate; endDate.Sub(i) >= 0; i = i.AddDate(0, 0, 1) {
			if slices.Contains(weekdays, i.Weekday().String()) {
				sql += `('` + wc + `', '` + i.Format("2006-01-02") + `', ` + target + `, '` + unit + `', ` + workers + `, ` + hours + `),`
			}
		}
		sql = sql[:len(sql)-1] + ` on conflict(workcenter, date) do update set target = EXCLUDED.target, workers = EXCLUDED.workers, hours = EXCLUDED.hours `
	} else {
		sql = `insert into targets(workcenter, date, target, unit, demandofmonth, workers, hours) values `
		for i := startDate; endDate.Sub(i) >= 0; i = i.AddDate(0, 0, 1) {
			if slices.Contains(weekdays, i.Weekday().String()) {
				sql += `('` + wc + `', '` + i.Format("2006-01-02") + `', ` + target + `, '` + unit + `',` + demand + `, ` + workers + `, ` + hours + `),`
			}
		}
		sql = sql[:len(sql)-1] + ` on conflict(workcenter, date) do update set target = EXCLUDED.target, demandofmonth = EXCLUDED.demandofmonth, workers = EXCLUDED.workers, hours = EXCLUDED.hours `
	}

	_, err := s.db.Exec(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("Fail to update!")
	}

	return c.SendString("Updates successful! ")
}

func (s *Server) getTargetsHistory(c *fiber.Ctx) error {
	rawDates := c.FormValue("dateRange")
	if rawDates == "" {
		return c.SendString("Choose dates to view history")
	}
	dateRange := strings.Split(rawDates, " - ")
	startDate, _ := time.Parse("2006-01-02", dateRange[0])
	endDate, _ := time.Parse("2006-01-02", dateRange[1])
	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")
	wc := c.FormValue("workcenter")
	var sql string
	if wc == "" {
		sql = `select date, workcenter, target, unit, workers, hours from targets where date >= '` + start + `' and date <= '` + end + `' order by date desc, workcenter`
	} else {

		sql = `select date, workcenter, target, unit, workers, hours from targets where date >= '` + start + `' and date <= '` + end + `' and workcenter = '` + wc + `' order by date desc`
	}
	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("")
	}
	var list = [][]string{}
	for rows.Next() {
		var a = make([]string, 6)
		rows.Scan(&a[0], &a[1], &a[2], &a[3], &a[4], &a[5])
		a[0] = strings.Split(a[0], "T")[0]
		list = append(list, a)
	}

	return c.Render("efficiency/targethistory", fiber.Map{
		"list": list,
	})
}

func (s *Server) safetyHandler(c *fiber.Ctx) error {
	rawDates := c.FormValue("safefromdate")
	start := "2024-01-01"
	end := time.Now().Format("2006-01-02")
	if rawDates != "" {
		dateRange := strings.Split(rawDates, " - ")
		start = dateRange[0]
		end = dateRange[1]
	}

	sql := `select count(id), max(accdate) from accidents where accdate >= '` + start + `' and accdate <= '` + end + `'`
	numberOfAccidents := 0
	var latestAccidentDate string
	if err := s.db.QueryRow(sql).Scan(&numberOfAccidents, &latestAccidentDate); err != nil {
		log.Println("fail to get numbers of accidents")
	}
	latestAccidentDate = strings.Split(latestAccidentDate, "T")[0]
	tmp, _ := time.Parse("2006-01-02", latestAccidentDate)
	latestAccidentDate = tmp.Format("02 Jan 2006")
	daysNoAcc := int(time.Since(tmp).Hours() / 24)
	return c.Render("efficiency/safety", fiber.Map{
		"numberOfAccidents":   numberOfAccidents,
		"lastestAccidentDate": latestAccidentDate,
		"daysNoAcc":           daysNoAcc,
	})
}

func (s *Server) viewreportHandler(c *fiber.Ctx) error {

	return c.Render("efficiency/viewreport", fiber.Map{}, "layout")
}

func (s *Server) viewreportPostHandler(c *fiber.Ctx) error {
	workcenter := c.FormValue("workcenter")
	fromdate := c.FormValue("fromdate")
	todate := c.FormValue("todate")

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	f.SetCellValue("Sheet1", "A1", "Work Center")
	f.SetCellValue("Sheet1", "B1", workcenter)
	f.SetCellValue("Sheet1", "D1", "From")
	f.SetCellValue("Sheet1", "E1", fromdate)
	f.SetCellValue("Sheet1", "F1", "To")
	f.SetCellValue("Sheet1", "G1", todate)
	f.SetCellValue("Sheet1", "A2", "Ngày giờ nhập")
	f.SetCellValue("Sheet1", "B2", "Sản Lượng")
	f.SetCellValue("Sheet1", "C2", "Manhr")
	f.SetCellValue("Sheet1", "D2", "Xưởng")
	f.SetCellValue("Sheet1", "E2", "Loại Hàng")
	f.SetCellValue("Sheet1", "F2", "Số lượng (pcs)")
	f.SetActiveSheet(1)
	sql := `select created_datetime, qty, manhr, factory_no, type, pcs from efficienct_reports 
		where work_center ='` + workcenter + `' and date >='` + fromdate + `' and date <='` + todate + `' order by date desc, created_datetime desc`

	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}

	var data [][]string
	i := 3
	for rows.Next() {
		var a = make([]string, 6)
		var t string
		rows.Scan(&t, &a[1], &a[2], &a[3], &a[4], &a[5])

		a[0] = t[0:19]
		a[0] = strings.Replace(a[0], "T", " ", 1)

		data = append(data, a)
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i), a[0])
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i), a[1])
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i), a[2])
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i), a[3])
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i), a[4])
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i), a[5])
		i++
	}
	if err := f.SaveAs("./static/uploads/Book1.xlsx"); err != nil {
		fmt.Println(err)
	}

	return c.Download("./static/uploads/Book1.xlsx")
}

func (s *Server) score6sHandler(c *fiber.Ctx) error {
	safefromdate := c.FormValue("fromdate")
	month := safefromdate[5:7]
	areas := []string{"CUTTING", "WOOD LAMINATION", "CURVE VENEER LAMINATION", "FLAT VENEER LAMINATION",
		"COMPONENT CNC", "COMPONENT FINEMILL", "PANEL CNC", "ASSEMBLY", "OEM ASSEMBLY", "WOOD FINISHING",
		"OEM WOOD FINISHING", "MECHANIC CAST IRON", "MECHANIC METAL", "WELDING", "METAL FINISHING", "CONCRETE",
		"UPHOLSTERY", "PACKING", "OEM PACKING", "QUALITY", "WH", "TECH", "MAINT", "PROCESS", "HR", "EHS",
	}
	numberOfareas := len(areas)

	sql := `select distinct date from score6s where date >= '2024-` + month + `-01' and date <= '2024-` + month + `-31' order by date`

	rows, _ := s.db.Query(sql)
	var data = [][]int{}
	var dates = []string{}
	var n = 0
	var targets = make([]int, numberOfareas)
	for i := 0; i < numberOfareas; i++ {
		targets[i] = 7
	}
	for rows.Next() {
		var a string
		rows.Scan(&a)
		dates = append(dates, a)
		var b = make([]int, numberOfareas)
		data = append(data, b)
		n++
	}

	sql = `select area, date, score from score6s where date >= '2024-` + month + `-01' and date <= '2024-` + month + `-31' order by date`

	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var a, b string
		var c int
		rows.Scan(&a, &b, &c)
		i := slices.Index(areas, a)
		j := slices.Index(dates, b)
		data[j][i] = c
	}

	return c.Render("efficiency/score6s", fiber.Map{
		"areas":   areas,
		"dates":   dates,
		"data":    data,
		"n":       n,
		"targets": targets,
	})
}

func (s *Server) input6sHandler(c *fiber.Ctx) error {

	return c.Render("efficiency/input6s", fiber.Map{"msg": "Form chấm điểm 6S"}, "layout")
}

func (s *Server) input6sPostHandler(c *fiber.Ctx) error {
	area := c.FormValue("area")
	if area == "" {
		return c.SendString("Vui lòng chọn khu vực.")
	}
	date := c.FormValue("date")
	score := c.FormValue("score")

	_, err := s.db.Exec(`insert into score6s(area, date, score) values ($1, $2, $3)`, area, date, score)
	if err != nil {
		log.Println(err)
		return c.SendString("Cập nhật thất bại, vui lòng kiểm tra lại dữ liệu nhập.")
	}

	return c.SendString("Cập nhật thành công")
}

func (s *Server) getlist6sPostHandler(c *fiber.Ctx) error {
	area := c.FormValue("area")
	date := c.FormValue("date")
	var sql string
	if area == "" {
		sql = `select * from score6s where date ='` + date + `'`
	} else {
		sql = `select * from score6s where area = '` + area + `' and date ='` + date + `'`
	}

	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("Fail to load data")
	}
	var list = [][]string{}
	for rows.Next() {
		var a = make([]string, 3)
		rows.Scan(&a[0], &a[1], &a[2])
		list = append(list, a)
	}

	return c.Render("efficiency/list6s", fiber.Map{
		"list": list,
	})
}

func (s *Server) qualityhistoryHandler(c *fiber.Ctx) error {
	date_issue := c.Query("date_issue")
	section_code := c.Query("section_code")

	var sql string
	if section_code == "" {
		sql = `select date_issue, section_code, qty_check, qty_fail from quatity_report
		 where date_issue = '` + date_issue + `'`
	} else {
		sql = `select date_issue, section_code, qty_check, qty_fail from quatity_report
			 where date_issue = '` + date_issue + `' and section_code = '` + section_code + `'`
	}

	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println("fail to access quality report")
		return c.SendString("Loi truy xuat")
	}
	var data = [][]string{}
	for rows.Next() {
		var a = make([]string, 4)
		rows.Scan(&a[0], &a[1], &a[2], &a[3])
		a[0] = a[0][:10]
		data = append(data, a)
	}

	return c.Render("efficiency/qualityhistory", fiber.Map{
		"data": data,
	})
}

func (s *Server) guideHandler(c *fiber.Ctx) error {

	return c.Render("staffquality/guide", fiber.Map{}, "layout")
}

func (s *Server) assemblyHandler(c *fiber.Ctx) error {
	fromdate := c.Query("fromdate")

	sql := `select distinct date from efficienct_reports where work_center = 'ASSEMBLY' and date >= '` + fromdate + `' order by date`
	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("Loi truy van")
	}
	var dates []string
	for rows.Next() {
		var a string
		rows.Scan(&a)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")
		dates = append(dates, a)
	}

	nod := len(dates)
	var rhlist1 = make([]float64, nod)
	var rhlist2 = make([]float64, nod)
	var brandlist1 = make([]float64, nod)
	var brandlist2 = make([]float64, nod)
	rows, err = s.db.Query(`SELECT date, factory_no, type, sum(qty) from 
 		efficienct_reports where work_center = 'ASSEMBLY' group by date, factory_no, type having 
 		date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}
	ld := ""
	i := -1
	for rows.Next() {
		var a, b, c string
		var d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		if ld != a {
			i++
			ld = a
		}
		if b == "1" && c == "RH" {
			rhlist1[i] = d
		}
		if b == "1" && c == "BRAND" {
			brandlist1[i] = d
		}
		if b == "2" && c == "BRAND" {
			brandlist2[i] = d
		}
		if b == "2" && c == "RH" {
			rhlist2[i] = d
		}
	}

	// targets
	rows, err = s.db.Query(`select date, target, workers, hours from targets 
	where workcenter = 'ASSEMBLY' and date >= '` + fromdate + `' order by date`)
	if err != nil {
		log.Println(err)
		return c.SendString("Loi lay du lieu targets")
	}
	var workers []int
	var hours []float64
	var targets []float64
	var datesOfTarget []string
	var tmp_targets []float64
	for rows.Next() {
		var a string
		var b, c, d float64
		rows.Scan(&a, &b, &c, &d)
		datesOfTarget = append(datesOfTarget, a)
		tmp_targets = append(tmp_targets, b)
		targets = append(targets, b*c*d)
		workers = append(workers, int(c))
		hours = append(hours, d)
	}

	rows, err = s.db.Query(`SELECT date, work_center, sum(qty), sum(manhr) from 
		efficienct_reports group by date, work_center having work_center = 'ASSEMBLY' 
		and date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}

	var efficiency []float64
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		i := slices.Index(datesOfTarget, a)
		if d == 0 || i == -1 {
			efficiency = append(efficiency, 0)
		} else {
			efficiency = append(efficiency, math.Round((c/d)*100/tmp_targets[i]))
		}
	}

	var latestCreated string
	rows, err = s.db.Query(`select created_datetime from efficienct_reports where work_center 
		= 'ASSEMBLY' order by id desc limit 1`)
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

	var month string
	var nextmonth string
	if slices.Contains([]string{"28", "29", "30", "31"}, fromdate[8:10]) {
		tmpt, _ := time.Parse("2006-01-02", fromdate)
		month = tmpt.Format("01")
		nextmonth = tmpt.AddDate(0, 1, 0).Format("01")
	} else {
		month = time.Now().Format("01")
		nextmonth = time.Now().AddDate(0, 1, 0).Format("01")
	}
	var demand float64

	sql = `select demandofmonth from targets where 
		workcenter = 'ASSEMBLY' and date >= '2024-` + month + `-01' 
		and date <= '2024-` + nextmonth + `-01' and demandofmonth <> 0 order by demandofmonth desc limit 1`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var a float64
		rows.Scan(&a)
		demand = a
	}

	var mtd float64
	sql = `select sum(qty) from efficienct_reports where work_center = 'ASSEMBLY' 
	 and date >='2024-` + month + `-01' and date <= '2024-` + nextmonth + `-01'`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		rows.Scan(&mtd)
	}
	mtdstr := message.NewPrinter(language.English).Sprintf("%.f", mtd)

	sql = `select distinct on (date) onconveyor from efficienct_reports 
		where date >= '` + fromdate + `' and work_center = 'ASSEMBLY' order by date nulls last`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("loi lay du lieu tren chuyen")
	}
	var onconveyors []float64
	for rows.Next() {
		var a float64
		rows.Scan(&a)
		onconveyors = append(onconveyors, a)
	}

	p := message.NewPrinter(language.English)

	return c.Render("efficiency/assemblychart", fiber.Map{
		"dates":         dates,
		"rhlist1":       rhlist1,
		"rhlist2":       rhlist2,
		"brandlist1":    brandlist1,
		"brandlist2":    brandlist2,
		"targets":       targets,
		"efficiency":    efficiency,
		"latestCreated": latestCreated,
		"demand":        p.Sprintf("%.f", demand),
		"mtd":           mtdstr,
		"onconveyors":   onconveyors,
		"workers":       workers,
		"hours":         hours,
	})
}

func (s *Server) woodfinishingHandler(c *fiber.Ctx) error {
	fromdate := c.Query("fromdate")

	sql := `select distinct date from efficienct_reports where work_center = 'WOODFINISHING' and date >= '` + fromdate + `' order by date`
	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("Loi truy van")
	}
	var dates []string
	for rows.Next() {
		var a string
		rows.Scan(&a)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")
		dates = append(dates, a)
	}

	nod := len(dates)
	var rhlist1 = make([]float64, nod)
	var rhlist2 = make([]float64, nod)
	var brandlist1 = make([]float64, nod)
	var brandlist2 = make([]float64, nod)
	var outsourcelist1 = make([]float64, nod)
	var outsourcelist2 = make([]float64, nod)
	rows, err = s.db.Query(`SELECT date, factory_no, type, sum(qty) from 
 		efficienct_reports where work_center = 'WOODFINISHING' group by date, factory_no, type having 
 		date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}
	ld := ""
	i := -1
	for rows.Next() {
		var a, b, c string
		var d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		if ld != a {
			i++
			ld = a
		}
		if b == "1" && c == "RH" {
			rhlist1[i] = d
		}
		if b == "1" && c == "BRAND" {
			brandlist1[i] = d
		}
		if b == "2" && c == "BRAND" {
			brandlist2[i] = d
		}
		if b == "2" && c == "RH" {
			rhlist2[i] = d
		}
		if b == "1" && c == "Outsource" {
			outsourcelist1[i] = d
		}
		if b == "2" && c == "Outsource" {
			outsourcelist2[i] = d
		}
	}

	// targets
	rows, err = s.db.Query(`select date, target, workers, hours from targets 
	where workcenter = 'WOODFINISHING' and date >= '` + fromdate + `' order by date`)
	if err != nil {
		log.Println(err)
		return c.SendString("Loi lay du lieu targets")
	}
	var targets []float64
	var datesOfTarget []string
	var tmp_targets []float64
	for rows.Next() {
		var a string
		var b, c, d float64
		rows.Scan(&a, &b, &c, &d)
		datesOfTarget = append(datesOfTarget, a)
		tmp_targets = append(tmp_targets, b)
		targets = append(targets, b*c*d)
	}

	rows, err = s.db.Query(`SELECT date, work_center, sum(qty), sum(manhr) from 
		efficienct_reports group by date, work_center having work_center = 'WOODFINISHING' 
		and date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}

	var efficiency []float64
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		i := slices.Index(datesOfTarget, a)
		if d == 0 || i == -1 {
			efficiency = append(efficiency, 0)
		} else {
			efficiency = append(efficiency, math.Round((c/d)*100/tmp_targets[i]))
		}
	}

	var latestCreated string
	rows, err = s.db.Query(`select created_datetime from efficienct_reports where work_center 
		= 'WOODFINISHING' order by id desc limit 1`)
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

	var month string
	var nextmonth string
	if slices.Contains([]string{"28", "29", "30", "31"}, fromdate[8:10]) {
		tmpt, _ := time.Parse("2006-01-02", fromdate)
		month = tmpt.Format("01")
		nextmonth = tmpt.AddDate(0, 1, 0).Format("01")
	} else {
		month = time.Now().Format("01")
		nextmonth = time.Now().AddDate(0, 1, 0).Format("01")
	}
	var demand float64

	sql = `select demandofmonth from targets where 
		workcenter = 'WOODFINISHING' and date >= '2024-` + month + `-01' 
		and date <= '2024-` + nextmonth + `-01' and demandofmonth <> 0 order by demandofmonth desc limit 1`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var a float64
		rows.Scan(&a)
		demand = a
	}

	var mtd float64
	sql = `select sum(qty) from efficienct_reports where work_center = 'WOODFINISHING' 
	 and date >='2024-` + month + `-01' and date <= '2024-` + nextmonth + `-01'`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		rows.Scan(&mtd)
	}
	mtdstr := message.NewPrinter(language.English).Sprintf("%.f", mtd)

	sql = `select distinct on (date) onconveyor from efficienct_reports 
		where date >= '` + fromdate + `' and work_center = 'WOODFINISHING' order by date nulls last`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("loi lay du lieu tren chuyen")
	}
	var onconveyors []float64
	for rows.Next() {
		var a float64
		rows.Scan(&a)
		onconveyors = append(onconveyors, a)
	}

	p := message.NewPrinter(language.English)

	return c.Render("efficiency/woodfinishingchart", fiber.Map{
		"dates":          dates,
		"rhlist1":        rhlist1,
		"rhlist2":        rhlist2,
		"brandlist1":     brandlist1,
		"brandlist2":     brandlist2,
		"targets":        targets,
		"efficiency":     efficiency,
		"latestCreated":  latestCreated,
		"demand":         p.Sprintf("%.f", demand),
		"mtd":            mtdstr,
		"onconveyors":    onconveyors,
		"outsourcelist1": outsourcelist1,
		"outsourcelist2": outsourcelist2,
	})
}

func (s *Server) packingHandler(c *fiber.Ctx) error {
	fromdate := c.Query("fromdate")

	sql := `select distinct date from efficienct_reports where work_center = 'PACKING' and date >= '` + fromdate + `' order by date`
	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("Loi truy van")
	}
	var dates []string
	for rows.Next() {
		var a string
		rows.Scan(&a)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")
		dates = append(dates, a)
	}

	nod := len(dates)
	var rhlist1 = make([]float64, nod)
	var rhlist2 = make([]float64, nod)
	var brandlist1 = make([]float64, nod)
	var brandlist2 = make([]float64, nod)
	var outsourcelist1 = make([]float64, nod)
	var outsourcelist2 = make([]float64, nod)
	rows, err = s.db.Query(`SELECT date, factory_no, type, sum(qty) from 
 		efficienct_reports where work_center = 'PACKING' group by date, factory_no, type having 
 		date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}
	ld := ""
	i := -1
	for rows.Next() {
		var a, b, c string
		var d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		if ld != a {
			i++
			ld = a
		}
		if b == "1" && c == "RH" {
			rhlist1[i] = d
		}
		if b == "1" && c == "BRAND" {
			brandlist1[i] = d
		}
		if b == "2" && c == "BRAND" {
			brandlist2[i] = d
		}
		if b == "2" && c == "RH" {
			rhlist2[i] = d
		}
		if b == "1" && c == "Outsource" {
			outsourcelist1[i] = d
		}
		if b == "2" && c == "Outsource" {
			outsourcelist2[i] = d
		}
	}

	// targets
	rows, err = s.db.Query(`select date, target, workers, hours from targets 
	where workcenter = 'PACKING' and date >= '` + fromdate + `' order by date`)
	if err != nil {
		log.Println(err)
		return c.SendString("Loi lay du lieu targets")
	}
	var targets []float64
	var datesOfTarget []string
	var tmp_targets []float64
	for rows.Next() {
		var a string
		var b, c, d float64
		rows.Scan(&a, &b, &c, &d)
		datesOfTarget = append(datesOfTarget, a)
		tmp_targets = append(tmp_targets, b)
		targets = append(targets, b*c*d)
	}

	rows, err = s.db.Query(`SELECT date, work_center, sum(qty), sum(manhr) from 
		efficienct_reports group by date, work_center having work_center = 'PACKING' 
		and date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}

	var efficiency []float64
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		i := slices.Index(datesOfTarget, a)
		if d == 0 || i == -1 {
			efficiency = append(efficiency, 0)
		} else {
			efficiency = append(efficiency, math.Round((c/d)*100/tmp_targets[i]))
		}
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

	var month string
	var nextmonth string
	if slices.Contains([]string{"28", "29", "30", "31"}, fromdate[8:10]) {
		tmpt, _ := time.Parse("2006-01-02", fromdate)
		month = tmpt.Format("01")
		nextmonth = tmpt.AddDate(0, 1, 0).Format("01")
	} else {
		month = time.Now().Format("01")
		nextmonth = time.Now().AddDate(0, 1, 0).Format("01")
	}
	var demand float64

	sql = `select demandofmonth from targets where 
		workcenter = 'PACKING' and date >= '2024-` + month + `-01' 
		and date <= '2024-` + nextmonth + `-01' and demandofmonth <> 0 order by demandofmonth desc limit 1`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var a float64
		rows.Scan(&a)
		demand = a
	}

	var mtd float64
	sql = `select sum(qty) from efficienct_reports where work_center = 'PACKING' 
	 and date >='2024-` + month + `-01' and date <= '2024-` + nextmonth + `-01'`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		rows.Scan(&mtd)
	}
	mtdstr := message.NewPrinter(language.English).Sprintf("%.f", mtd)

	sql = `select distinct on (date) onconveyor from efficienct_reports 
		where date >= '` + fromdate + `' and work_center = 'PACKING' order by date nulls last`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
		return c.SendString("loi lay du lieu tren chuyen")
	}
	var onconveyors []float64
	for rows.Next() {
		var a float64
		rows.Scan(&a)
		onconveyors = append(onconveyors, a)
	}

	p := message.NewPrinter(language.English)

	return c.Render("efficiency/packingchart", fiber.Map{
		"dates":          dates,
		"rhlist1":        rhlist1,
		"rhlist2":        rhlist2,
		"brandlist1":     brandlist1,
		"brandlist2":     brandlist2,
		"targets":        targets,
		"efficiency":     efficiency,
		"latestCreated":  latestCreated,
		"demand":         p.Sprintf("%.f", demand),
		"mtd":            mtdstr,
		"onconveyors":    onconveyors,
		"outsourcelist1": outsourcelist1,
		"outsourcelist2": outsourcelist2,
	})
}
