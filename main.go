package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/lib/pq"
)

func main() {
	// init db connection
	// conn, err := sql.Open("postgres", "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway")
	// if err != nil {
	// 	panic(err)
	// }
	// defer conn.Close()

	// init server app
	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// routes here
	app.Static("/static", "./static")

	app.Get("/", func(c *fiber.Ctx) error {
		log.Println("get")
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layout")
	}).Name("index")

	app.Get("/accident", func(c *fiber.Ctx) error {
		log.Println("enter accident")

		return c.Render("accident", fiber.Map{}, "layout")
	})

	app.Post("/accident", func(c *fiber.Ctx) error {
		log.Println("post accident")
		accdate := c.FormValue("accdate")

		conn, err := sql.Open("postgres", "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		sqlStatement := `
    INSERT INTO accidents (accdate)
    VALUES ($1)
		`
		_, err = conn.Exec(sqlStatement, accdate)

		if err != nil {
			panic(err)
		}
		return c.Redirect("/dashboard", fiber.StatusFound)
	})

	app.Get("/shipped", func(c *fiber.Ctx) error {
		log.Println("enter shipped")
		return c.Render("shipped", nil, "layout")
	})

	app.Post("/shipped", func(c *fiber.Ctx) error {
		shipdate := c.FormValue("shipdate")
		money := c.FormValue("money")

		conn, err := sql.Open("postgres", "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		sqlStatement := `
    INSERT INTO ship (shipdate, money)
    VALUES ($1, $2)
		`
		_, err = conn.Exec(sqlStatement, shipdate, money)

		if err != nil {
			panic(err)
		}

		return c.Redirect("/shipped", fiber.StatusFound)
	})

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		log.Println("enter dashboard")
		days := int(time.Since(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)).Hours() / 24)
		var accidents int

		conn, err := sql.Open("postgres", "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		rows, err := conn.Query("SELECT shipdate FROM ship")
		if err != nil {
			panic(err)
		}
		var shipdate []string
		var money []float64
		for rows.Next() {
			var version string
			rows.Scan(&version)
			shipdate = append(shipdate, version)
		}

		rows, err = conn.Query("SELECT money FROM ship")
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			var version float64
			rows.Scan(&version)
			money = append(money, version)
		}

		rows, err = conn.Query("SELECT count(accdate) FROM accidents where accdate >= '2024-01-01'")
		if err != nil {
			panic(err)
		}
		rows.Next()
		rows.Scan(&accidents)

		defer rows.Close()

		return c.Render("dashboard", fiber.Map{
			"shipdate":  shipdate,
			"money":     money,
			"days":      days,
			"accidents": accidents,
		}, "layout")
	}).Name("dashboard")

	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{
			"Title": "About",
		}, "layout")
	})
	app.Post("/about", func(c *fiber.Ctx) error {
		log.Println(c.FormValue("message"))
		return c.Render("about", fiber.Map{
			"Title": "About",
		}, "layout")
	})
	app.Delete("/contact/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		log.Println(id)
		return c.Redirect("/", fiber.StatusSeeOther)
	})
	app.Get("/change", func(c *fiber.Ctx) error {
		log.Println(c.FormValue("message"))
		return c.SendString(c.FormValue("message"))
	})

	// run server
	app.Listen(getPort())
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

// ************* database *************
func connectDb() {
	connectionStr := "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway"

	conn, err := sql.Open("postgres", connectionStr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT version();")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var version string
		rows.Scan(&version)
		fmt.Println(version)
	}
	defer rows.Close()
}

//*************/database *************
