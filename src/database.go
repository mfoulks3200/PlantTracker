package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Plant struct {
	PlantID      int
	PlantName    string
	PlantVariety string
	BelongsTo    int
	LocationID   int
	PlantDate    string
}

type Variety struct {
	VarietyID        int
	VarietyName      string
	VarietyFamily    string
	VarietyCatagory  string
	SunlightCatagory string
	WaterCatagory    string
	AvgSprout        int
	AvgHarvest       int
}

var db, sqlerr = sql.Open("sqlite3", "./planttracker.db")

func initDB() {
	logMessage("Core", "Opening DB")

	if sqlerr != nil {
		log.Fatal("Fatal SQL ERR")
		log.Fatal(sqlerr)
	}

}

func getVariety(vID int) (v Variety) {
	var stmt, err = db.Query("select * from `varieties` where varietyID = ? ", vID)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for stmt.Next() {
		err = stmt.Scan(&v.VarietyID, &v.VarietyName, &v.VarietyFamily, &v.VarietyCatagory, &v.SunlightCatagory, &v.WaterCatagory, &v.AvgSprout, &v.AvgHarvest)
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func getPlant(pID int) (p Plant) {
	var stmt, err = db.Query("select * from `plants` where plantID = ? ", pID)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for stmt.Next() {
		err = stmt.Scan(&p.PlantID, &p.PlantName, &p.PlantVariety, &p.BelongsTo, &p.LocationID, &p.PlantDate)
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}
