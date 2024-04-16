package main

import (
	_ "github.com/lib/pq"
	"github.com/xuri/excelize/v2"
)

// func process(w http.ResponseWriter, r *http.Request) {
// 	r.FormFile("file")

// }

func main() {
	f, _ := excelize.OpenFile("firsttest.xlsx")
	f.SetCellFormula("Sheet2", "A1", "=SUM(A2:A3)")

	f.Save()

	// dbstore := NewPosgreDB()
	// Server := NewServer(":3000", dbstore.db)
	// Server.Run()
}
