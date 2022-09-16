package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ivanmakarychev/social-network/internal/repository"
)

func makeWriteLoad() {
	db, err := createMySQLCluster()
	if err != nil {
		log.Fatal(err)
	}
	l := writeLoader{
		db:            db,
		nameGenerator: newNameGenerator(),
	}
	l.start()
}

type writeLoader struct {
	db                         *repository.MySQLCluster
	nameGenerator              *nameGenerator
	sigChan                    chan os.Signal
	transactionCounter         int64
	successTransactionsCounter int64
}

func (l *writeLoader) start() {
	log.Println("write loader started")
	l.sigChan = make(chan os.Signal, 1)
	signal.Notify(l.sigChan, os.Interrupt)
insertLoop:
	for {
		select {
		case <-l.sigChan:
			l.printStatistics()
			break insertLoop
		default:
			l.insert()
		}
	}
	log.Println("write loader finished")
}

func (l *writeLoader) insert() {
	nameSurname := l.nameGenerator.generate()
	l.transactionCounter++
	rs, err := l.db.Master().Exec(
		"insert into profile (first_name, surname) values (?, ?)",
		nameSurname.name,
		nameSurname.surname,
	)
	if err != nil {
		log.Println("exec failed:", err)
		return
	}
	rowsInserted, err := rs.RowsAffected()
	if err != nil {
		log.Println("rowsAffected failed:", err)
		return
	}
	l.successTransactionsCounter += rowsInserted
}

func (l *writeLoader) printStatistics() {
	fmt.Printf("Total transactions: %d\n", l.transactionCounter)
	fmt.Printf("Successful transactions: %d\n", l.successTransactionsCounter)
	fmt.Printf("Lost transactions: %d\n", l.transactionCounter-l.successTransactionsCounter)
}
