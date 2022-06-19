package repository

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/ivanmakarychev/social-network/internal/config"
)

func CreateMySQLConnectionAndInitDB(cfg config.Database) (*sql.DB, error) {
	log.Println("creating mysql connection")
	return createMySQLConnectionWithRetry(cfg, 60)
}

func createMySQLConnectionWithRetry(cfg config.Database, retries int) (*sql.DB, error) {
	const dbSourceFmt = "%s:%s@(%s)/social-network?parseTime=true"
	dbSource := fmt.Sprintf(dbSourceFmt, cfg.User, cfg.Password, cfg.Master)

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
			err = checkError(err)
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

func checkError(err error) error {
	me, ok := err.(*mysql.MySQLError)
	if !ok {
		return err
	}
	const duplicateKey = uint16(1061)
	if me.Number == duplicateKey {
		return nil
	}
	return err
}
