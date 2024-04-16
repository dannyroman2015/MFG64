package main

import (
	_ "github.com/lib/pq"
)

func main() {
	// f, err := excelize.OpenFile("firsttest.xlsx")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer func() {
	// 	if err := f.Close(); err != nil {
	// 		fmt.Println(err)
	// 	}
	// }()

	// cell, err := f.GetCellValue("Sheet2", "A2")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(cell)

	// rows, _ := f.GetRows("Sheet2")
	// for _, v := range rows {
	// 	for _, c := range v {
	// 		fmt.Println(c)
	// 	}
	// }

	dbstore := NewPosgreDB()
	Server := NewServer(":3000", dbstore.db)
	Server.Run()
}
