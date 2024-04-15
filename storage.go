package main

import "database/sql"

type PosgreDB struct {
	db *sql.DB
}

func NewPosgreDB() *PosgreDB {
	db, err := sql.Open("postgres", "postgresql://postgres:kbEviyUjJecPLMxXRNweNyvIobFzCZAQ@monorail.proxy.rlwy.net:27572/railway")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}

	return &PosgreDB{
		db: db,
	}
}
