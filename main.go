package main

import (
	"log"

	"github.com/ivanmakarychev/social-network/internal/authorization"
	"github.com/ivanmakarychev/social-network/internal/presentation"
	"github.com/ivanmakarychev/social-network/internal/repository"
)

func main() {
	db, err := repository.CreateMySQLConnectionAndInitDB()
	if err != nil {
		log.Fatal("failed to create MySQL connection: ", err)
	}
	defer db.Close()

	citiesRepo := repository.NewCitiesRepositoryImpl(db)
	interestRepo := repository.NewInterestsRepositoryImpl(db)
	friendsRepo := repository.NewFriendsRepoImpl(db)
	profileRepo := repository.NewProfileRepoImpl(db, friendsRepo)
	authManager := authorization.NewManagerImpl(db)

	app := presentation.NewApp(
		authManager,
		profileRepo,
		citiesRepo,
		interestRepo,
		friendsRepo,
	)

	log.Fatal(app.Run())
}
