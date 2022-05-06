package repository

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func CreateMySQLConnectionAndInitDB() (*sql.DB, error) {
	log.Println("creating mysql connection")
	return createMySQLConnectionWithRetry(60)
}

func createMySQLConnectionWithRetry(retries int) (*sql.DB, error) {
	const dbSource = "social-network-user:sQ7mDXwwLcfq@(db:3306)/social-network?parseTime=true"
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
	if err != nil {
		return nil, err
	}

	err = initDB(db)
	return db, err
}

func initDB(db *sql.DB) error {
	log.Println("init DB started")

	const path = "./internal/repository/init.sql"

	script, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open sql script file: %s", err)
	}
	defer script.Close()

	scanner := bufio.NewScanner(script)

	sb := strings.Builder{}

	counter := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			sb.WriteString(line)
			sb.WriteRune(' ')
		} else {
			continue
		}
		if strings.HasSuffix(line, ";") {
			query := sb.String()
			sb.Reset()
			log.Println("[query]", query)
			_, err = db.Exec(query)
			if err != nil {
				log.Println("[init db] bad query:", query, "[error]", err)
				return fmt.Errorf("failed to execute sql script file: %s", err)
			}
			counter++
		}
	}

	log.Println("init DB finished.", counter, "queries executed")

	return nil
}
