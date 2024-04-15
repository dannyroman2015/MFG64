package main

import "database/sql"

type PosgreDB struct {
	Connection *sql.DB
}

func NewPosgreDB(uri string) PosgreDB {
	conn, err := sql.Open("postgres", uri)
	if err != nil {
		panic(err)
	}
	return PosgreDB{Connection: conn}
}
