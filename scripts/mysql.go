package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func createMySQLConnectionWithRetry(retries int) (*sql.DB, error) {
	const dbSource = "social-network-user:sQ7mDXwwLcfq@(localhost:3306)/social-network?parseTime=true"
	db, err := sql.Open("mysql", dbSource)
	for i := 0; err != nil && i < retries; i++ {
		log.Println("retrying opening db")
		db, err = sql.Open("mysql", dbSource)
	}
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	err = db.Ping()
	for i := 0; err != nil && i < retries; i++ {
		time.Sleep(time.Second)
		log.Println("retrying pinging db")
		err = db.Ping()
	}
	return db, err
}
