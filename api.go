package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/template/html/v2"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type InputDate_data struct {
	Mo_id      string
	Product_id string
	Section_id string
	InputDate  string
	Qty        int
	Staff      string
}

type Section_data struct {
	Mo_id      string
	Product_id string
	Section_id string
	Needed_qty string
	Done_qty   int
}

type Subblueprint_data struct {
	Mo_id        string
	Product_id   string
	Needed_qty   int
	Done_qty     int
	Done_percent int
}

type Blueprint_data struct {
	Mo_id        string
	Blueprint_id string
	Needed_qty   int
	Done_qty     int
	Done_percent int
}

type Mo_data struct {
	Mo_id        string
	Done_percent int
}

type Server struct {
	addr string
	db   *sql.DB
}

func (s *Server) Run() {
	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	//test import and process excel file
	//end test

	//just for test
	app.Get("/test", func(c *fiber.Ctx) error {

		return c.Render("test", fiber.Map{}, "layout")
	})
	app.Post("/fortest", func(c *fiber.Ctx) error {
		log.Println("fortest")
		log.Println(c.FormValue("sss"))
		return c.Render("fragments/t", nil)
	})
	//end just for test

	app.Static("/static", "./static")

	app.Use("/efficiencyreport", basicauth.New(basicauth.Config{
		Users: map[string]string{
			"assembly":   "2511",
			"cutting":    "7526",
			"lamination": "6824",
			"reeded":     "2037",
			"veneer":     "2486",
			"panel":      "1357",
			"wood":       "1302",
			"packing":    "9342",
			"nguyen":     "4526",
			"admin":      "1011",
		},
	}))

	// app.Get("/", s.indexGetHandler)
	app.Get("/", s.efficiencyHandler)

	app.Get("/guide", s.guideHandler)

	app.Get("/evaluate", s.evaluateHandler)
	app.Post("/workerbypw", s.workerbypwPostHandler)
	app.Post("/searchstaff", s.searchstaffPostHandler)

	app.Get("/efficiency", s.efficiencyHandler)
	app.Get("/efficiencywithdate", s.efficiencyWithdateHandler)

	app.Get("/assembly", s.assemblyHandler)
	app.Get("/woodfinishing", s.woodfinishingHandler)
	app.Get("/packing", s.packingHandler)

	app.Get("/login", s.loginGetHandler)

	app.Get("/efficiencyreport", s.efficiencyReportHandler)
	app.Post("/efficiencyreport", s.efficiencyReportPostHandler)

	app.Get("/viewreport", s.viewreportHandler)
	app.Post("/viewreport", s.viewreportPostHandler)

	app.Get("/inputmanhr", s.inputmanhrHandler)
	app.Post("/inputmanhr", s.inputmanhrPostHandler)

	app.Get("/qualityinput", s.qualityInputHandler)
	app.Post("/qualityinput", s.qualityInputPostHandler)
	app.Get("/qualityhistory", s.qualityhistoryHandler)

	app.Get("/qualitychart", s.qulityChartHandler)
	app.Get("/qualityquickinput", s.qualityquickinputHandler)
	app.Post("qualityquickinput", s.qualityquickinputPostHandler)

	app.Get("/safety", s.safetyHandler)
	app.Get("/score6s", s.score6sHandler)
	app.Get("/input6s", s.input6sHandler)
	app.Post("/input6s", s.input6sPostHandler)
	app.Post("/getlist6s", s.getlist6sPostHandler)

	app.Get("/prodAdBlueprints/:mo_id", s.prodAdBlueprintsHandler)

	app.Get("/productionadmin", s.productionAdminGetHandler)
	app.Get("/prodadfilter/:status", s.prodAdFilterHandler)
	app.Get("/prods/:mo_id/:blueprint_id", s.prodsHandler)
	app.Get("/section/:mo_id/:product_id/:needed_qty", s.sectionHandler)
	app.Get("/inputdate/:mo_id/:product_id/:section_id", s.inputdateHandler)

	app.Get("/inputSection", s.inputSectionHandler)
	app.Post("/inputSection", s.inputSectionPostHandler)
	app.Get("/inputSection/:mo_id/:product_id/:section_id", s.inputSectionWithParamsHandler)
	app.Get("/sections/productIds", s.getProductIdsHandler)
	app.Post("/section/checkremains", s.checkremainsHandler)

	app.Get("/efficientChart/:workcenter", s.efficientChartHandler)
	app.Get("/importreededf", s.importreededfHangler)
	app.Post("/proccess_reeded_excelfile", s.proccess_reeded_excelfilePostHandler)
	app.Get("/reededchart", s.reededcahrtHandler)
	app.Get("/woodrecover", s.woodrecoveryHandler)
	app.Get("/inputwoodrecovery", s.inputputwoodrecoveryHandler)
	app.Post("/inputwoodrecovery", s.inputwoodrecoveryPostHandler)
	app.Get("/cutting_wh", s.cuttingwhHandler)
	app.Get("/inputwhissue", s.inputwhissueHandler)
	app.Post("/inputwhissue", s.inputwhissuePostHandler)
	app.Get("/summarytable", s.summarytableHandler)
	app.Get("/inputsummary", s.inputsummaryHandler)
	app.Post("/proccessforsummary", s.proccessforsummaryHandler)

	app.Get("/target", s.targetHandler)
	app.Post("target", s.targetPostHandler)
	app.Post("/targethistory", s.getTargetsHistory)

	app.Get("/prodvalueChart", s.prodvalueChartHandler)

	app.Get("/importexcelfile", s.importexcelfileHandler)
	app.Post("/proccesexcelfile", s.proccesexcelfileHandler)

	app.Get("/provalue", s.provalueGetHandler)
	app.Post("/provalue", s.provaluePostHandler)

	app.Get("/accident", s.accidentGetHandler)
	app.Post("/accident", s.accidentPostHandler)

	app.Get("/shipped", s.shippedGetHandler)
	app.Post("/shipped", s.shippedPostHandler)

	app.Get("/dashboard/:dtype", s.dashboardGetHandler)

	app.Get("/about", s.aboutGetHandler)
	app.Post("/about", s.aboutPostHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}
	app.Listen(port)
}

func NewServer(addr string, db *sql.DB) *Server {
	return &Server{
		addr: addr,
		db:   db,
	}
}

func (s *Server) indexGetHandler(c *fiber.Ctx) error {
	log.Println("enter index")
	return c.Render("index", fiber.Map{
		"Title": "Hello, World!",
	}, "layout")
}

func (s *Server) provaluePostHandler(c *fiber.Ctx) error {
	fd := c.FormValue("finishdate")
	pt := c.FormValue("proType")
	fn := c.FormValue("fac_no")
	mn := c.FormValue("money")

	sqlStatement := `INSERT INTO moneyvalue (dateissue, type, money, factory_no)VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(sqlStatement, fd, pt, mn, fn)
	if err != nil {
		panic(err)
	}
	return c.Redirect("provalue", fiber.StatusFound)
}

func (s *Server) provalueGetHandler(c *fiber.Ctx) error {
	return c.Render("provalue", fiber.Map{}, "layout")
}

func (s *Server) accidentGetHandler(c *fiber.Ctx) error {
	return c.Render("efficiency/accident", fiber.Map{}, "layout")
}

func (s *Server) accidentPostHandler(c *fiber.Ctx) error {
	accdate := c.FormValue("accdate")
	person := c.FormValue("person")
	description := c.FormValue("description")

	sqlStatement := `INSERT INTO accidents (accdate, name, description) VALUES ($1, $2, $3)`
	_, err := s.db.Exec(sqlStatement, accdate, person, description)
	if err != nil {
		panic(err)
	}
	return c.Redirect("/accident", fiber.StatusSeeOther)
}

func (s *Server) shippedGetHandler(c *fiber.Ctx) error {
	log.Println("enter shipped")
	return c.Render("shipped", nil, "layout")
}

func (s *Server) shippedPostHandler(c *fiber.Ctx) error {
	shipdate := c.FormValue("shipdate")
	money := c.FormValue("money")
	sqlStatement := `INSERT INTO ship (shipdate, money) VALUES ($1, $2)`
	_, err := s.db.Exec(sqlStatement, shipdate, money)

	if err != nil {
		panic(err)
	}

	return c.Redirect("/shipped", fiber.StatusFound)
}

func (s *Server) dashboardGetHandler(c *fiber.Ctx) error {
	// data for efficient chart
	var efficient_cutting_dates []string
	var efficient_cutting_data []float64
	var cutting_efficiencies []float64
	var actual_target float64
	var cutting_targets []float64
	var target float64

	rows, err := s.db.Query(`select actual_target, target from efficienct_workcenter 
		where workcenter = 'CUTTING'`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&actual_target, &target)
	}

	rows, err = s.db.Query(`SELECT date, work_center, sum(qty), sum(manhr) 
		from efficienct_reports group by date, work_center having work_center = 'CUTTING' 
		order by date`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("Jan 2")

		cutting_efficiencies = append(cutting_efficiencies, math.Round((c/d)*100/actual_target))
		efficient_cutting_dates = append(efficient_cutting_dates, a)
		efficient_cutting_data = append(efficient_cutting_data, c)
		cutting_targets = append(cutting_targets, target)
	}
	// end data for efficient charts

	dtype := c.Params("dtype")
	var whichToDate string

	days := int(time.Since(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)).Hours() / 24)
	var accidents int

	rows, err = s.db.Query("SELECT shipdate, money FROM ship order by shipdate")
	if err != nil {
		panic(err)
	}
	var shipdate []string
	var money []string
	for rows.Next() {
		var s, m string
		rows.Scan(&s, &m)
		str := strings.Split(s, "T")[0]
		shipdate = append(shipdate, str)
		money = append(money, m)
	}

	rows, err = s.db.Query("SELECT count(accdate) FROM accidents where accdate >= '2024-01-01'")
	if err != nil {
		panic(err)
	}
	rows.Next()
	rows.Scan(&accidents)

	switch dtype {
	case "today":
		whichToDate = time.Now().Format("2006-01-02")
	case "yesterday":
		whichToDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	case "MTD":
		whichToDate = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	default:
		whichToDate = time.Date(time.Now().Year(), 01, 01, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	}

	rows, err = s.db.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND type = 'OEM'")
	if err != nil {
		panic(err)
	}
	rows.Next()
	var sumOEM float64
	rows.Scan(&sumOEM)

	rows, err = s.db.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND type = 'BRAND'")
	if err != nil {
		panic(err)
	}
	rows.Next()
	var sumBRAND float64
	rows.Scan(&sumBRAND)

	rows, err = s.db.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND factory_no = '1'")
	if err != nil {
		panic(err)
	}
	rows.Next()
	var factory_1 string
	rows.Scan(&factory_1)

	rows, err = s.db.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND factory_no = '2'")
	if err != nil {
		panic(err)
	}
	rows.Next()
	var factory_2 string
	rows.Scan(&factory_2)

	rows, err = s.db.Query("SELECT dateissue, sum(money) FROM moneyvalue group by dateissue order by dateissue")
	if err != nil {
		panic(err)
	}

	var dateissue, moneys []string
	for rows.Next() {
		var v, m string
		rows.Scan(&v, &m)
		dateissue = append(dateissue, strings.Split(v, "T")[0])
		moneys = append(moneys, m)
	}

	defer rows.Close()

	return c.Render("dashboard", fiber.Map{
		"shipdate":                shipdate,
		"money":                   money,
		"days":                    days,
		"accidents":               accidents,
		"sumOEM":                  sumOEM,
		"sumBRAND":                sumBRAND,
		"factory_1":               factory_1,
		"factory_2":               factory_2,
		"dateissue":               dateissue,
		"proValue":                moneys,
		"dtype":                   dtype,
		"efficient_cutting_dates": efficient_cutting_dates,
		"efficient_cutting_data":  efficient_cutting_data,
		"cutting_efficiencies":    cutting_efficiencies,
		"cutting_targets":         cutting_targets,
	}, "layout")
}

func (s *Server) aboutGetHandler(c *fiber.Ctx) error {
	return c.Render("about", fiber.Map{
		"Title": "About",
	}, "layout")
}

func (s *Server) aboutPostHandler(c *fiber.Ctx) error {
	log.Println(c.FormValue("message"))
	return c.Render("about", fiber.Map{
		"Title": "About",
	}, "layout")
}

func (s *Server) productionAdminGetHandler(c *fiber.Ctx) error {
	var mos_data []Mo_data

	status := "running"
	var sql string

	switch status {
	case "all":
		sql = "Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking group by mo_id order by mo_id"
	case "running":
		sql = `Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking 
					group by mo_id having sum(done_qty) > 0 and sum(done_qty) < sum(needed_qty) order by mo_id`
	case "ready":
		sql = `Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking 
					group by mo_id having sum(done_qty) = 0 order by mo_id`
	case "done":
		sql = `Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking 
								group by mo_id having sum(done_qty) = 100 order by mo_id`
	}

	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var mo string
		var needed, done int
		rows.Scan(&mo, &needed, &done)
		mos_data = append(mos_data, Mo_data{
			Mo_id:        mo,
			Done_percent: done * 100 / needed,
		})
	}

	return c.Render("production_admin/main", fiber.Map{
		"mos": mos_data,
	}, "layout")
}

func (s *Server) searchPostProdAdHandler(c *fiber.Ctx) error {
	log.Println("search")
	return c.Redirect("/productionadmin", fiber.StatusFound)
}

func (s *Server) loginGetHandler(c *fiber.Ctx) error {

	return c.Render("login", nil)
}

func (s *Server) extractFromExcelHandler(c *fiber.Ctx) error {
	return nil
}

func (s *Server) prodAdFilterHandler(c *fiber.Ctx) error {
	var mos_data []Mo_data
	status := strings.ToLower(c.Params("status"))
	var sql string

	switch status {
	case "all":
		sql = "Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking group by mo_id order by mo_id"
	case "running":
		sql = `Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking 
					group by mo_id having sum(done_qty) > 0 and sum(done_qty) < sum(needed_qty) order by mo_id`
	case "ready":
		sql = `Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking 
					group by mo_id having sum(done_qty) = 0 order by mo_id`
	case "done":
		sql = `Select mo_id, sum(needed_qty), sum(done_qty) from mo_tracking 
								group by mo_id having sum(done_qty) = sum(needed_qty) order by mo_id`
	}

	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var mo string
		var needed, done int
		rows.Scan(&mo, &needed, &done)
		mos_data = append(mos_data, Mo_data{
			Mo_id:        mo,
			Done_percent: done * 100 / needed,
		})
	}

	return c.Render("production_admin/listMos", fiber.Map{
		"mos": mos_data,
	})
}

func (s *Server) prodAdBlueprintsHandler(c *fiber.Ctx) error {
	var blueprints_data []Blueprint_data

	mo_id := c.Params("mo_id")
	sql := `SELECT mo_id, blueprint_id, sum(m.needed_qty), sum(m.done_qty) 
					FROM mo_tracking m join products p on m.product_id = p.product_id 
					GROUP BY mo_id, p.blueprint_id HAVING mo_id = '` + mo_id + `'`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var data Blueprint_data
		rows.Scan(&data.Mo_id, &data.Blueprint_id, &data.Needed_qty, &data.Done_qty)
		data.Done_percent = data.Done_qty * 100 / data.Needed_qty
		blueprints_data = append(blueprints_data, data)
	}

	return c.Render("production_admin/listBlueprints", fiber.Map{
		"blueprints": blueprints_data,
	})
}

func (s *Server) prodsHandler(c *fiber.Ctx) error {
	var subBp []Subblueprint_data

	rows, err := s.db.Query("SELECT m.product_id, m.needed_qty, m.done_qty FROM mo_tracking m join products p on m.product_id = p.product_id where m.mo_id ='" + c.Params("mo_id") + "' and p.blueprint_id ='" + c.Params("blueprint_id") + "'")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var data Subblueprint_data
		rows.Scan(&data.Product_id, &data.Needed_qty, &data.Done_qty)
		data.Done_percent = data.Done_qty * 100 / data.Needed_qty
		data.Mo_id = c.Params("mo_id")
		subBp = append(subBp, data)
	}

	return c.Render("production_admin/listSubblueprint", fiber.Map{
		"subBluePrint": subBp,
	})
}

func (s *Server) sectionHandler(c *fiber.Ctx) error {
	var sections_data []Section_data

	sql := `SELECT mo_id, product_id, section, sum(qty) FROM prod_reports GROUP BY mo_id, product_id, section 
					HAVING mo_id = '` + c.Params("Mo_id") + `' and product_id = '` + c.Params("Product_id") + `'`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var data Section_data
		rows.Scan(&data.Mo_id, &data.Product_id, &data.Section_id, &data.Done_qty)
		data.Needed_qty = c.Params("Needed_qty")
		sections_data = append(sections_data, data)
	}

	return c.Render("production_admin/listSection", fiber.Map{
		"sections": sections_data,
	})
}

func (s *Server) inputdateHandler(c *fiber.Ctx) error {
	var inputdates []InputDate_data
	mo_id := c.Params("Mo_id")
	product_id := c.Params("Product_id")
	section_id := c.Params("section_id")

	sql := `SELECT mo_id, product_id, section, input_date, qty, staff FROM prod_reports 
				WHERE mo_id = '` + c.Params("Mo_id") + `' and 
				product_id = '` + c.Params("Product_id") + `' and 
				section = '` + c.Params("section_id") + `'`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var data InputDate_data
		rows.Scan(&data.Mo_id, &data.Product_id, &data.Section_id, &data.InputDate, &data.Qty, &data.Staff)
		inputdates = append(inputdates, data)
	}

	return c.Render("production_admin/listInputdates", fiber.Map{
		"inputdates": inputdates,
		"mo_id":      mo_id,
		"product_id": product_id,
		"section_id": section_id,
	})
}

func (s *Server) inputSectionHandler(c *fiber.Ctx) error {
	// mo_id := c.Params("Mo_id")
	// product_id := c.Params("Product_id")
	// section_id := c.Params("Section_id")
	var Mo_ids []string

	sql := `select mo_id from mo_tracking group by mo_id 
					having sum(done_qty) < sum(needed_qty)`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var mo string
		rows.Scan(&mo)
		Mo_ids = append(Mo_ids, mo)
	}

	return c.Render("section/inputSection", fiber.Map{
		"data": map[string]interface{}{
			// "Mo_id":      mo_id,
			// "Product_id": product_id,
			// "Section_id": section_id,
			"Mo_ids": Mo_ids,
		},
	}, "layout")
}

func (s *Server) getProductIdsHandler(c *fiber.Ctx) error {
	mo_id := c.Query("mo")
	var product_ids []string

	sql := `select product_id from mo_tracking where 
					mo_id = '` + mo_id + `' and done_qty < needed_qty`

	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var product_id string
		rows.Scan(&product_id)
		product_ids = append(product_ids, product_id)
	}

	return c.Render("section/listProducts", fiber.Map{
		"product_ids": product_ids,
	})
}

func (s *Server) checkremainsHandler(c *fiber.Ctx) error {
	input_qty, _ := strconv.Atoi(c.FormValue("qty"))
	mo_id := c.FormValue("mo")
	product_id := c.FormValue("productId")
	section := c.FormValue("section_id")
	var sectionDoneQty int
	var needed_qty int

	sql := `select sum(qty) from prod_reports group by mo_id, product_id, section 
		having mo_id = '` + mo_id + `' and product_id = '` + product_id + `' and 
		section = '` + section + `'`
	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&sectionDoneQty)
	}

	sql = `select needed_qty from mo_tracking where 
		mo_id = '` + mo_id + `' and product_id = '` + product_id + `'`
	rows, err = s.db.Query(sql)
	if err != nil {
		panic(err)
	}

	rows.Next()
	rows.Scan(&needed_qty)

	remains := needed_qty - sectionDoneQty

	var message string
	var msgColor string
	if input_qty > remains {
		message = fmt.Sprintf("Invalid. Your input quanity is greater than needs. Remains %d", remains)
		msgColor = "is-danger"
	} else {
		message = fmt.Sprintf("Valid. The remains are %d", remains-input_qty)
		msgColor = "is-link"
	}

	return c.Render("section/inputQtyError", fiber.Map{
		"message":   message,
		"input_qty": input_qty,
		"remains":   remains,
		"msgColor":  msgColor,
	})
}

func (s *Server) inputSectionWithParamsHandler(c *fiber.Ctx) error {
	return c.SendString("chua lam")
}

func (s *Server) inputSectionPostHandler(c *fiber.Ctx) error {
	mo_id := c.FormValue("mo")
	product_id := c.FormValue("productId")
	section := c.FormValue("section_id")
	input_date := c.FormValue("inputdate")
	qty := c.FormValue("qty")
	staff := c.FormValue("staff")

	sql := `insert into prod_reports(mo_id, product_id, section, input_date, qty, staff)
					values($1, $2, $3, $4, $5, $6)`

	_, err := s.db.Exec(sql, mo_id, product_id, section, input_date, qty, staff)
	if err != nil {
		panic(err)
	}

	return c.Redirect("/inputSection", fiber.StatusFound)
}

func (s *Server) importexcelfileHandler(c *fiber.Ctx) error {
	return c.Render("production_admin/importexcelfile", fiber.Map{}, "layout")
}

func (s *Server) proccesexcelfileHandler(c *fiber.Ctx) error {
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
	cell, err := f.GetCellValue("Sheet2", "A2")
	if err != nil {
		panic(err)

	}

	// c.SaveFile(file, "static/public/"+file.Filename)

	// f, err := excelize.OpenFile("static/public/" + file.Filename)
	// if err != nil {
	// 	panic(err)
	// }
	// cell, err := f.GetCellValue("Sheet1", "A2")
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println(cell)

	return c.Render("test", fiber.Map{
		"excel": cell,
	})
}

func (s *Server) efficiencyReportHandler(c *fiber.Ctx) error {
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
	user := c.Locals("username").(string)
	workcenters := map[string][]string{
		"cutting":    {"CUTTING"},
		"lamination": {"LAMINATION"},
		"reeded":     {"REEDEDLINE"},
		"veneer":     {"VENEERLAMINATION"},
		"panel":      {"PANELCNC"},
		"assembly":   {"ASSEMBLY"},
		"wood":       {"WOODFINISHING"},
		"packing":    {"PACKING"},
		"nguyen":     {"CUTTING", "LAMINATION", "REEDEDLINE", "VENEERLAMINATION"},
		"admin":      {"CUTTING", "LAMINATION", "REEDEDLINE", "VENEERLAMINATION", "PANELCNC", "ASSEMBLY", "WOODFINISHING", "PACKING"},
	}
	wc := workcenters[user]
	// sql := `select date, qty, manhr, type, factory_no, cnc_machine from efficienct_reports where work_center in'` + wc + `'
	// 			order by date desc limit 10`
	sql := `select date, qty, manhr, type, factory_no, cnc_machine from efficienct_reports where work_center in (`
	for _, w := range wc {
		sql += `'` + w + `'`
	}
	sql += `) order by date desc limit 10`

	rows, err := s.db.Query(sql)
	if err != nil {
		panic(err)
	}
	var list [][]string
	for rows.Next() {
		var a, b, c, d, e, f string
		rows.Scan(&a, &b, &c, &d, &e, &f)
		a = strings.Split(a, "T")[0]
		ar := []string{a, b, c, d, e, f}
		list = append(list, ar)
	}

	return c.Render("efficiency/report", fiber.Map{
		"units": units,
		// "unit":       units[wc],
		"workcenters": workcenters[user],
		"list":        list,
	}, "layout")
}

func (s *Server) efficiencyReportPostHandler(c *fiber.Ctx) error {
	var factory string
	var typeofproduct string
	cncmachine := c.FormValue("cncmachine")
	workcenter := c.FormValue("workcenter")
	inputdate := c.FormValue("inputdate")
	factory = c.FormValue("factory")
	typeofproduct = c.FormValue("typeofproduct")
	color := c.FormValue("color")
	var qty, manhr float64
	qty, _ = strconv.ParseFloat(c.FormValue("qty"), 64)
	manhr, _ = strconv.ParseFloat(c.FormValue("manhr"), 64)
	pcs := c.FormValue("pcs")
	onconveyor := c.FormValue("qtywait")
	veneertype := c.FormValue("veneertype")

	// if pcs == "" || pcs == "0" {
	// sql = `insert into efficienct_reports(work_center, date, qty, manhr, type, factory_no, cnc_machine, created_datetime) values ($1, $2, $3, $4, $5, $6, $7, current_timestamp)`
	// } else {
	sql := `insert into efficienct_reports(work_center, date, qty, manhr, type, factory_no, cnc_machine, created_datetime, pcs, onconveyor, veneertype, reededline_color) values ($1, $2, $3, $4, $5, $6, $7, current_timestamp, $8, $9, $10, $11)`
	// }

	_, err := s.db.Exec(sql, workcenter, inputdate, qty, manhr, typeofproduct, factory, cncmachine, pcs, onconveyor, veneertype, color)
	if err != nil {
		panic(err)
	}

	return c.Redirect("/efficiencyreport", fiber.StatusFound)
}

func (s *Server) efficientChartHandler(c *fiber.Ctx) error {
	workcenter := strings.ToUpper(c.Params("workcenter"))
	fromdate := c.Query("fromdate")

	var labels []string
	var quanity []float64
	var efficiency []float64
	// var actual_target float64
	// var target float64

	var units = map[string]string{
		"Production Value": "Amount($)", "CUTTING": "Quantity(cbm)", "LAMINATION": "Quantity(m2)", "REEDEDLINE": "Quantity(m2)", "VENEERLAMINATION": "Quantity(m2)", "PANELCNC": "Quantity(sheet)", "ASSEMBLY": "Amount($)", "WOODFINISHING": "Amount($)", "PACKING": "Amount($)",
	}
	var targetUnits = map[string]string{
		"Production Value": "$", "CUTTING": "cbm", "LAMINATION": "m2", "REEDEDLINE": "m2", "VENEERLAMINATION": "m2", "PANELCNC": "sheet", "ASSEMBLY": "$", "WOODFINISHING": "$", "PACKING": "$",
	}

	rows, err := s.db.Query(`select date, target, workers, hours from targets 
	where workcenter = '` + workcenter + `' and date >= '` + fromdate + `' order by date`)
	// rows, err := s.db.Query(`select actual_target, target from efficienct_workcenter
	// 	where workcenter = '` + workcenter + `'`)
	if err != nil {
		panic(err)
	}
	// for rows.Next() {
	// 	rows.Scan(&actual_target, &target)
	// }

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
		efficienct_reports group by date, work_center having work_center = '` + workcenter + `' 
		and date >= '` + fromdate + `' order by date`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var a, b string
		var c, d float64
		rows.Scan(&a, &b, &c, &d)
		i := slices.Index(datesOfTarget, a)
		a = strings.Split(a, "T")[0]
		t, _ := time.Parse("2006-01-02", a)
		a = t.Format("2 Jan")
		if d == 0 || i == -1 {
			efficiency = append(efficiency, 0)
		} else {
			efficiency = append(efficiency, math.Round((c/d)*100/tmp_targets[i]))
		}

		labels = append(labels, a)
		c = math.Round(c)
		quanity = append(quanity, c)
		// if i == -1 {
		// 	targets = append(targets, 0)
		// } else {
		// 	targets = append(targets, tmp_targets[i]*d)
		// }
	}

	var latestCreated string
	rows, err = s.db.Query(`select created_datetime from efficienct_reports where work_center 
		= '` + workcenter + `' order by id desc limit 1`)
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

	sql := `select demandofmonth from targets where 
		workcenter = '` + workcenter + `' and date >= '2024-` + month + `-01' 
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
	sql = `select sum(qty) from efficienct_reports where work_center = '` + workcenter + `' 
	 and date >='2024-` + month + `-01' and date <= '2024-` + nextmonth + `-01'`
	rows, err = s.db.Query(sql)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		rows.Scan(&mtd)
	}

	log.Println(targets)

	mtdstr := message.NewPrinter(language.English).Sprintf("%.f", mtd)
	randColor := fmt.Sprintf("rgba(%d, %d, %d, 0.4)", rand.Intn(255), rand.Intn(255), rand.Intn(255))

	return c.Render("efficiency/chart", fiber.Map{
		"workcenter":    workcenter,
		"labels":        labels,
		"quanity":       quanity,
		"efficiency":    efficiency,
		"targets":       targets,
		"chartLabels":   []string{"Quantity", "Efficiency(%)", "Target"},
		"bg_color":      randColor,
		"units":         units,
		"latestCreated": latestCreated,
		"targetUnits":   targetUnits,
		"demand":        message.NewPrinter(language.English).Sprintf("%.f", demand),
		"mtd":           mtdstr,
	})
}
