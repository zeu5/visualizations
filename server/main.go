package server

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kamva/mgm/v3"
	"github.com/zeu5/visualizations/log"
	"github.com/zeu5/visualizations/server/config"
	"github.com/zeu5/visualizations/server/middleware"
	"github.com/zeu5/visualizations/server/routes"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Run(configPath string) {
	fmt.Println("Initializing...")
	gin.SetMode(gin.ReleaseMode)

	config, err := config.ParseConfig(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Init(&config.Log)
	err = mgm.SetDefaultConfig(nil, "vis", options.Client().ApplyURI(config.DBURI))
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to initialize db: %s", err))
	}

	router := gin.New()
	router.Use(middleware.Logger)

	routes.Initialize(router)
	fmt.Println("Starting server...")
	router.Run(config.ServerAddr)
}
