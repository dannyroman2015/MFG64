package main

import (
	_ "github.com/lib/pq"
)

// func process(w http.ResponseWriter, r *http.Request) {
// 	r.FormFile("file")

// }

func main() {

	dbstore := NewPosgreDB()
	Server := NewServer(":3000", dbstore.db)
	Server.Run()
}
