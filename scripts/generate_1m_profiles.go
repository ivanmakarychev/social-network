package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/Masterminds/squirrel"
)

func generate1MProfiles() {
	db, err := createMySQLConnectionWithRetry(1)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	g := newNameGenerator()

	for k := 0; k < 1000; k++ {
		insert := squirrel.Insert("profile").Columns("name", "surname")

		for l := 0; l < 1000; l++ {
			ns := g.generate()
			insert = insert.Values(ns.name, ns.surname)
		}

		query, args, err := insert.ToSql()
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type nameGenerator struct {
	males    []string
	females  []string
	surnames []string
}

type nameSurname struct {
	name    string
	surname string
}

func newNameGenerator() *nameGenerator {
	return &nameGenerator{
		males:    getList("/Users/imakarychev/github.com/ivanmakarychev/social-network/scripts/males.txt"),
		females:  getList("/Users/imakarychev/github.com/ivanmakarychev/social-network/scripts/females.txt"),
		surnames: getList("/Users/imakarychev/github.com/ivanmakarychev/social-network/scripts/surnames.txt"),
	}
}

func (g *nameGenerator) generate() nameSurname {
	result := nameSurname{}
	result.surname = g.surnames[rand.Intn(len(g.surnames))]

	isFemale := rand.Intn(2)
	if isFemale == 1 {
		result.surname = result.surname + "а"
		result.name = g.females[rand.Intn(len(g.females))]
	} else {
		result.name = g.males[rand.Intn(len(g.males))]
	}
	return result
}

func getList(filename string) []string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var r []string

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		r = append(r, line)
	}

	return r
}
