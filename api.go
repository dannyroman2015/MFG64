package main

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type Server struct {
	App *fiber.App
}

func (s *Server) Run() {
	conn, err := sql.Open("postgres", "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	s.App.Static("/static", "./static")

	s.App.Get("/", func(c *fiber.Ctx) error {
		log.Println("enter index")
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layout")
	})

	s.App.Get("/test", func(c *fiber.Ctx) error {
		q := "trung"
		return c.Render("test", fiber.Map{
			"q": q,
		})
	})

	s.App.Delete("/fortest", func(c *fiber.Ctx) error {
		time.Sleep(1 * time.Second)
		str := c.FormValue("s")
		log.Println(str)
		return c.SendString("")
	})

	s.App.Get("/provalue", func(c *fiber.Ctx) error {
		return c.Render("provalue", fiber.Map{}, "layout")
	})

	s.App.Post("/provalue", func(c *fiber.Ctx) error {
		fd := c.FormValue("finishdate")
		pt := c.FormValue("proType")
		fn := c.FormValue("fac_no")
		mn := c.FormValue("money")

		sqlStatement := `INSERT INTO moneyvalue (dateissue, type, money, factory_no)VALUES ($1, $2, $3, $4)`
		_, err := conn.Exec(sqlStatement, fd, pt, mn, fn)
		if err != nil {
			panic(err)
		}
		return c.Redirect("provalue", fiber.StatusFound)
	})

	s.App.Get("/accident", func(c *fiber.Ctx) error {
		log.Println("enter accident")

		return c.Render("accident", fiber.Map{}, "layout")
	}).Name("accident")

	s.App.Post("/accident", func(c *fiber.Ctx) error {
		log.Println("post accident")
		accdate := c.FormValue("accdate")

		sqlStatement := `INSERT INTO accidents (accdate) VALUES ($1)`
		_, err := conn.Exec(sqlStatement, accdate)
		if err != nil {
			panic(err)
		}
		return c.Redirect("/dashboard", fiber.StatusFound)
	})

	s.App.Get("/shipped", func(c *fiber.Ctx) error {
		log.Println("enter shipped")
		return c.Render("shipped", nil, "layout")
	})

	s.App.Post("/shipped", func(c *fiber.Ctx) error {
		shipdate := c.FormValue("shipdate")
		money := c.FormValue("money")
		sqlStatement := `INSERT INTO ship (shipdate, money) VALUES ($1, $2)`
		_, err := conn.Exec(sqlStatement, shipdate, money)

		if err != nil {
			panic(err)
		}

		return c.Redirect("/shipped", fiber.StatusFound)
	})

	s.App.Get("/dashboard/:dtype", func(c *fiber.Ctx) error {
		dtype := c.Params("dtype")
		var whichToDate string

		days := int(time.Since(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)).Hours() / 24)
		var accidents int

		rows, err := conn.Query("SELECT shipdate, money FROM ship order by shipdate")
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

		rows, err = conn.Query("SELECT count(accdate) FROM accidents where accdate >= '2024-01-01'")
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

		rows, err = conn.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND type = 'OEM'")
		if err != nil {
			panic(err)
		}
		rows.Next()
		var sumOEM float64
		rows.Scan(&sumOEM)

		rows, err = conn.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND type = 'BRAND'")
		if err != nil {
			panic(err)
		}
		rows.Next()
		var sumBRAND float64
		rows.Scan(&sumBRAND)

		rows, err = conn.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND factory_no = '1'")
		if err != nil {
			panic(err)
		}
		rows.Next()
		var factory_1 string
		rows.Scan(&factory_1)

		rows, err = conn.Query("SELECT sum(money) FROM moneyvalue where dateissue >= '" + whichToDate + "' AND factory_no = '2'")
		if err != nil {
			panic(err)
		}
		rows.Next()
		var factory_2 string
		rows.Scan(&factory_2)

		rows, err = conn.Query("SELECT dateissue, sum(money) FROM moneyvalue group by dateissue order by dateissue")
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
	}).Name("dashboard")

	s.App.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{
			"Title": "About",
		}, "layout")
	})

	s.App.Post("/about", func(c *fiber.Ctx) error {
		log.Println(c.FormValue("message"))
		return c.Render("about", fiber.Map{
			"Title": "About",
		}, "layout")
	})

	s.App.Delete("/contact/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		log.Println(id)
		return c.Redirect("/", fiber.StatusSeeOther)
	})

	s.App.Get("/change", func(c *fiber.Ctx) error {
		log.Println(c.FormValue("message"))
		return c.SendString(c.FormValue("message"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}
	s.App.Listen(port)
}

func NewServer() *Server {
	return &Server{
		App: fiber.New(fiber.Config{Views: html.New("./templates", ".html")}),
	}
}
