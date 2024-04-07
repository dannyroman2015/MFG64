package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/lib/pq"
)

func main() {
	// init db connection
	conn, err := sql.Open("postgres", "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

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
		log.Println("change")
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
