package main

import (
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"karting-grafana/database"
	"karting-grafana/parsers"
	"log"
	"os"
)

const AIM_CSV_FORMAT = "AiM CSV File"

func main() {
	filepaths := os.Args[1:]
	if len(filepaths) == 0 {
		fmt.Println("No file paths provided. Uploaded nothing. Exiting.")
		os.Exit(0)
	}

	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading sample.env file")
	}

	// Init mongodb connection
	db, disconnect := database.InitMongoDbClient()
	defer disconnect()

	// Parse csvs then upload lap times and lap frames to Mongodb
	for _, path := range filepaths {
		csv := parsers.ReadCsv(path)
		if csv[0][1] != AIM_CSV_FORMAT {
			fmt.Println(path, "is not Aim CSV File format")
			continue
		}
		meta := parsers.ParseMeta(csv)

		// Load csv data into db
		parsers.CreateLapTimes(csv, meta, db)
		parsers.CreateLapFrames(csv[parsers.LAP_START_INDEX:], meta, db)
	}
}
