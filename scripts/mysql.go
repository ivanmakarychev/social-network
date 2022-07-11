package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/ivanmakarychev/social-network/internal/config"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

func createMySQLCluster() (*repository.MySQLCluster, error) {
	db := repository.NewMySQLCluster(config.Database{
		User:     "social-network-user",
		Password: "sQ7mDXwwLcfq",
		Master:   "db1:3306",
		Replicas: []string{
			"db2:3307",
			"db3:3308",
		},
	},
		func(host string) string {
			return "localhost"
		},
	)
	err := db.Connect()
	return db, err
}
