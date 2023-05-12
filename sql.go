package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/jmoiron/sqlx"
	"os"
	"time"
)

func loadConfigSQL(file string) ConfigSQL {
	var config ConfigSQL
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	var result map[string]*json.RawMessage
	if err := json.NewDecoder(configFile).Decode(&result); err != nil {
		fmt.Println(err.Error())
	}

	//json.NewDecoder(configFile).Decode(&config)
	json.Unmarshal(*result["Db"], &config)
	return config
}

// Inits DB Connection
func dbConnection() *sql.DB {
	config := loadConfigSQL("./config.json")
	info := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, 3306, config.DB)
	db, err := sql.Open("mysql", info)
	if err != nil {
		panic(err.Error())
	}
	return db
}

//func dbPutExample() error {
//	db := dbConnection()
//	defer db.Close()
//
//	_, err := db.Exec(
//		"INSERT INTO power (user, channel, w) VALUES (UUID_TO_BIN('249d442a-d5fc-4c65-bb43-7f21d7eecfd4', true), 'disposablePower', 180), (UUID_TO_BIN('44146e78-0d7a-4c2f-b2e4-f3fcccd29175', true), 'disposablePower', 3540), (UUID_TO_BIN('7d61316c-1456-452c-ac41-a373756b38a3', true), 'disposablePower', 120), (UUID_TO_BIN('153f948a-fa4a-446f-aa13-fa70be3f15fb', true), 'disposablePower', 600), (UUID_TO_BIN('03aa9c1b-49db-4f19-9fda-c46928afe320', true), 'disposablePower', 480), (UUID_TO_BIN('8a970131-b6e1-4767-a446-225fdf89d3e1', true), 'disposablePower', 600), (UUID_TO_BIN('1865ce5f-f865-43ba-bd9c-5c18f7f4f17e', true), 'disposablePower', 480), (UUID_TO_BIN('f0676026-f7e0-40d8-9102-79beb81716c4', true), 'disposablePower', 3000), (UUID_TO_BIN('ac3c7dea-9093-4b0c-90b1-b3350322d673', true), 'disposablePower', 2400), (UUID_TO_BIN('3de6fe02-ca22-48f9-8735-b093797e4e53', true), 'disposablePower', 540), (UUID_TO_BIN('470236a9-0045-4798-8f0b-1a6a73cfdce4', true), 'disposablePower', 420), (UUID_TO_BIN('d23b4d4f-78ea-4d9b-af7b-18e25c020fcc', true), 'disposablePower', 480), (UUID_TO_BIN('0d0c672a-bc57-40cb-8cc7-d2596bfddab0', true), 'disposablePower', 180);")
//	return err
//}

// ------------------------------------C-Quartier Implementations--------------------------------------

type Pow struct {
	Time   time.Time `json:""`
	userId string    `json:""`
	w      float64   `json:"value"`
}

// This is part of the code, used for the HTTP endpoint (access from central server)
// alterntiver Name für die Funktion: dbGetStatusOfChangesTable
func dbReadCurrentPowerOfChangesTable() Pow {
	db := dbConnection()
	var wData Pow

	// Query
	results, err := db.Query("SELECT time as t, BIN_TO_UUID(userid, true) as uId, wchange as wPower FROM changes ORDER BY time DESC LIMIT 1;")
	if err != nil {
		return wData
	}
	defer results.Close()
	defer db.Close()

	//Results
	for results.Next() {
		err = results.Scan(&wData.Time, &wData.userId, &wData.w)
		if err != nil {
			fmt.Println(err)
			return wData
		}
	}
	return wData
}

func dbWriteDataToConsumingTable(val int, user *User) error {
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"INSERT INTO consumption(uuid, userid, postalcode, city, address, wamount) values ((UUID_TO_BIN(?, true)),(UUID_TO_BIN(?, true)),?,?,?,?);",
		uuid.New().String(),
		user.Uuid,
		user.PostalCode,
		user.City,
		user.Address,
		val,
	)
	if err != nil {
		fmt.Println(err)
	}
	return err

}

func dbWriteDataToFeedingTable(value int, usr *User) error {
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"INSERT INTO feeding(uuid, userid, postalcode, city, address, powerstorage, pscapacity, wamount) VALUES (UUID_TO_BIN(?, true),(UUID_TO_BIN(?, true)),?,?,?,?,?,?);",
		uuid.New().String(),
		usr.Uuid,
		usr.PostalCode,
		usr.City,
		usr.Address,
		usr.PowerStorage,
		usr.PsCapacity,
		value,
	)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func dbWriteDataToChangesTable(absolut int, change int, user *User) error {
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"INSERT INTO changes(uuid,userid, absVal, wChange) values ((UUID_TO_BIN(?, true)),(UUID_TO_BIN(?, true)),?,?);",
		uuid.New().String(),
		user.Uuid,
		absolut,
		change,
	)
	if err != nil {
		fmt.Println(err)
	}
	return err

}

func dbReadLastAmountOfPower(tableName string) int {
	db := dbConnection()
	var wVal int

	// Query
	results, err := db.Query("SELECT wamount as wPower FROM " + tableName + " ORDER BY time DESC LIMIT 1;")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer results.Close()
	defer db.Close()

	//Results
	for results.Next() {
		err = results.Scan(&wVal)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
	return wVal
}

func dbReadLastEntry(tableName string) Transaction {
	db := dbConnection()
	var transaction Transaction

	//Query
	results, err := db.Query("SELECT uuid as u, time as t, userid as uid, wamount as w FROM " + tableName + " ORDER BY time DESC LIMIT 1;")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	//Results
	for results.Next() {
		err = results.Scan(&transaction.Uuid, &transaction.Time, &transaction.Userid, &transaction.AbsVal)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
	return transaction
}

func dbReadLastChangesEntry() ChangesTransaction {
	db := dbConnection()
	var tc ChangesTransaction

	//Query
	results, err := db.Query("SELECT uuid as u, time as t, userid as uid, absVal as av, wChange as w FROM changes ORDER BY time DESC LIMIT 1;")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	//Results
	for results.Next() {
		err = results.Scan(&tc.Uuid, &tc.Time, &tc.Userid, &tc.AbsVal, &tc.WChange)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
	return tc
}

// checkEmptinessOfTable checks if a table is empty and returns a bool indicating this
func checkEmptinessOfTable(tableName string) bool {
	db := dbConnection()
	var count = 0
	// Query
	//var query = "SELECT COUNT(*) AS RowCnt From " + tableName + ";"
	results, err := db.Query("SELECT COUNT(*) AS RowCnt From " + tableName + ";")
	if err != nil {
		panic(err)
	}
	defer results.Close()
	defer db.Close()

	//Results
	//err = results.Scan(&count) // delete line
	for results.Next() {
		err = results.Scan(&count)
		if err != nil {
			panic(err)
		}
	}
	if count != 0 { //RowCount > 0 means there are already some entries
		return false
	}
	return true
}

func getWattOutOfVzLoggerData(ap vzloggerAPIResponse) (int, error) {
	err := validateStructure(ap)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return int(ap.Data[1].Tuples[0][1]), err
}

//--------------------------------------------------------------------------------------------
