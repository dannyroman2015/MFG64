package main

import (
	_ "github.com/lib/pq"
)

func main() {
	dbstore := NewPosgreDB()
	Server := NewServer(":3000", dbstore.db)
	Server.Run()
}
