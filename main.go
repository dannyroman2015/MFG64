package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {

	engine := html.New("./templates", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Hooks().OnName(func(r fiber.Route) error {
		log.Println("hook onname", r.Name)
		return nil
	})

	app.Static("/static", "./static")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layout")
	}).Name("index")

	app.Get("/about", func(c *fiber.Ctx) error {
		log.Println("kshfkh")
		return c.Render("about", fiber.Map{
			"Title": "About",
			"Person": Person{
				Name: "Trung",
				Age:  18,
			},
		})
	})

	app.Listen(getPort())
}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (c Person) Show(s string) string {
	return c.Name + " yeu " + s
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
