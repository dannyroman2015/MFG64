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

	app.Get("/productionadmin", s.productionAdminGetHandler)

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

	return c.Render("productionadmin", fiber.Map{}, "layout")
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
