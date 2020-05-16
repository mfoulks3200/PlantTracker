package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type PlantList struct {
	Plants []Plant
}

type VarietyList struct {
	Varieties []Variety
}

type Plant struct {
	PlantID      int
	PlantName    string
	VarietyID    int
	Variety      Variety
	BelongsTo    int
	LocationID   int
	PlantDate    string
	LocationName string
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

type Location struct {
	LocationID   int
	LocationName string
	BelongsTo    int
}

var db, sqlerr = sql.Open("sqlite3", "./planttracker.db")

func initDB() {
	logMessage("Core", "Opening DB")

	if sqlerr != nil {
		log.Fatal("Fatal SQL ERR")
		log.Fatal(sqlerr)
	}

}

func getAllVarieties() VarietyList {
	var stmt, err = db.Query("select * from `varieties`")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var plnts VarietyList
	for stmt.Next() {
		var plnt Variety
		err = stmt.Scan(&plnt.VarietyID, &plnt.VarietyName, &plnt.VarietyFamily, &plnt.VarietyCatagory, &plnt.SunlightCatagory, &plnt.WaterCatagory, &plnt.AvgSprout, &plnt.AvgHarvest)
		if err != nil {
			logMessage("Core", "user lookup error")
			log.Fatal(err)
		}
		plnts.Varieties = append(plnts.Varieties, plnt)
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	return plnts
}

func getAllPlants(userID int) PlantList {
	var query string
	if userID == -1 {
		query = "select * from `plants`"
	} else {
		query = "select * from `plants` where belongsTo = ?"
	}
	var stmt, err = db.Query(query, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var plnts PlantList
	for stmt.Next() {
		var plnt Plant
		err = stmt.Scan(&plnt.PlantID, &plnt.PlantName, &plnt.VarietyID, &plnt.BelongsTo, &plnt.LocationID, &plnt.PlantDate)
		if err != nil {
			logMessage("Core", "user lookup error")
			log.Fatal(err)
		}
		plnt = getPlantData(plnt)
		plnts.Plants = append(plnts.Plants, plnt)
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	return plnts
}

func getLocation(lID int) (l Location) {
	var stmt, err = db.Query("select * from `location` where locationID = ? ", lID)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for stmt.Next() {
		err = stmt.Scan(&l.LocationID, &l.LocationName, &l.BelongsTo)
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
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
		err = stmt.Scan(&p.PlantID, &p.PlantName, &p.VarietyID, &p.BelongsTo, &p.LocationID, &p.PlantDate)
	}
	err = stmt.Err()
	if err != nil {
		log.Fatal(err)
	}
	p = getPlantData(p)
	return
}

func getPlantData(plant Plant) (p Plant) {
	p = plant
	if p.PlantName == "" {
		p = getPlant(p.PlantID)
	}
	if p.Variety.VarietyName == "" {
		var v Variety = getVariety(plant.VarietyID)
		p.Variety = v
	}

	var l Location = getLocation(plant.LocationID)
	p.LocationName = l.LocationName
	return
}

func createVariety(v Variety) {
	db.Exec("INSERT INTO varieties (varietyName, varietyFamily, varietyCatagory, sunlightCatagory, waterCatagory, avgSproutTime, avgHarvestTime) VALUES (?,?,?,?,?,?,?);", v.VarietyName, v.VarietyFamily, v.VarietyCatagory, v.SunlightCatagory, v.WaterCatagory, v.AvgSprout, v.AvgHarvest)
}
