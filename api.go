package main

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/xuri/excelize/v2"
)

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
	app.Post("/testprocessimportedexcelfile", s.testprocessimportedexcelfile)
	app.Get("/testimportexcel", s.testimportexcel)
	//end test

	//just for test
	app.Get("/test", func(c *fiber.Ctx) error {

		return c.Render("test", fiber.Map{
			"status": "waiting",
		})
	})
	app.Get("/fortest", func(c *fiber.Ctx) error {
		log.Println("fortest")
		return c.Render("fragments/t", nil)
	})
	//end just for test

	app.Static("/static", "./static")

	app.Get("/", s.indexGetHandler)

	app.Get("/login", s.loginGetHandler)

	app.Get("/prodAdBlueprints/:mo_id", s.prodAdBlueprintsHandler)

	app.Get("/productionadmin", s.productionAdminGetHandler)
	app.Get("/prodadfilter/:status", s.prodAdFilterHandler)
	app.Get("/prods/:mo_id/:blueprint_id", s.prodsHandler)
	app.Get("/section/:mo_id/:product_id/:needed_qty", s.sectionHandler)
	app.Get("/inputdate", s.inputdateHandler)

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
	log.Println("enter accident")

	return c.Render("accident", fiber.Map{}, "layout")
}

func (s *Server) accidentPostHandler(c *fiber.Ctx) error {
	log.Println("post accident")
	accdate := c.FormValue("accdate")

	sqlStatement := `INSERT INTO accidents (accdate) VALUES ($1)`
	_, err := s.db.Exec(sqlStatement, accdate)
	if err != nil {
		panic(err)
	}
	return c.Redirect("/accident", fiber.StatusFound)
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
	dtype := c.Params("dtype")
	var whichToDate string

	days := int(time.Since(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)).Hours() / 24)
	var accidents int

	rows, err := s.db.Query("SELECT shipdate, money FROM ship order by shipdate")
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
		"shipdate":  shipdate,
		"money":     money,
		"days":      days,
		"accidents": accidents,
		"sumOEM":    sumOEM,
		"sumBRAND":  sumBRAND,
		"factory_1": factory_1,
		"factory_2": factory_2,
		"dateissue": dateissue,
		"proValue":  moneys,
		"dtype":     dtype,
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

func (s *Server) testimportexcel(c *fiber.Ctx) error {
	return c.Render("testexcel", nil)
}

func (s *Server) testprocessimportedexcelfile(c *fiber.Ctx) error {
	log.Println("exceltest")
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
	log.Println(cell)
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
	log.Println("here")
	return c.Render("production_admin/listInputdates", fiber.Map{})
}
