package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"
	"os"
	"time"
)

type ConfigSQL struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DB       string `json:"dbname"`
	UserId   string `json:"userId"`
}

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
	json.Unmarshal(*result["db"], &config)
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

func dbGetAllPower() []Power {
	db := dbConnection()
	var power []Power

	//Query
	results, err := db.Query("SELECT time, BIN_TO_UUID(user, true) as user, w as watts, channel FROM power")
	if err != nil {
		return power
	}
	defer results.Close()
	defer db.Close()

	//Results
	for results.Next() {
		var p Power
		err = results.Scan(&p.Time, &p.User, &p.w, &p.Channel)
		if err != nil {
			return power
		}
		power = append(power, p)
	}
	return power
}

func dbGetAllEnergy() []Energy {
	db := dbConnection()
	var energy []Energy

	//Query
	results, err := db.Query("SELECT time, BIN_TO_UUID(user, true) as user, wh as watthours, channel FROM energy")
	if err != nil {
		return energy
	}
	defer results.Close()
	defer db.Close()

	//Results
	for results.Next() {
		var e Energy
		err = results.Scan(&e.Time, &e.User, &e.wh, &e.Channel)
		if err != nil {
			return energy
		}
		energy = append(energy, e)
	}
	return energy
}

func dbGetRecentDisposablePower() []Power {
	db := dbConnection()
	var power []Power

	//Query
	results, err := db.Query("SELECT time, BIN_TO_UUID(user, true) as user, w AS watts, channel FROM power WHERE channel = 'disposablePower' AND time >= NOW() - INTERVAL 10 SECOND")
	if err != nil {
		return power
	}
	defer results.Close()
	defer db.Close()

	//Results
	for results.Next() {
		var p Power
		err = results.Scan(&p.Time, &p.User, &p.w, &p.Channel)
		if err != nil {
			return power
		}
		power = append(power, p)
	}
	return power
}

func dbGetRecentSMPower() []Power {
	db := dbConnection()
	var power []Power

	//Query
	results, err := db.Query("SELECT time, BIN_TO_UUID(user, true) as user, w AS watts, channel FROM power WHERE channel = 'vzlogger' AND time >= NOW() - INTERVAL 10 SECOND")
	if err != nil {
		return power
	}
	defer results.Close()
	defer db.Close()

	//Results
	for results.Next() {
		var p Power
		err = results.Scan(&p.Time, &p.User, &p.w, &p.Channel)
		if err != nil {
			return power
		}
		power = append(power, p)
	}
	return power
}

func dbGetUsers() []User {
	config := loadConfigSQL("./config.json")
	info := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, 3306, config.DB)
	db, err := sqlx.Connect("mysql", info)

	var users []User
	err = db.Select(&users, "SELECT BIN_TO_UUID(uuid, true) as uuid, name, seller, public_key, plz, email, password, iban, joindate, chainid, active, minprice, maxprice, matched, url FROM users;")
	if err != nil {
		return users
	}
	return users
}

func dbPutEMSPower(ap OpenEMSAPIResponse, user *User) error {
	err := validateStructure(ap)
	if err != nil {
		return err
	}
	db := dbConnection()
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO power(user, channel, w) values (UUID_TO_BIN(?, true),?,?)",
		user.Uuid,
		ap.Address,
		ap.Value,
	)
	return err
}

func dbPutSMPower(ap vzloggerAPIResponse, user *User) error {
	err := validateStructure(ap)
	if err != nil {
		return err
	}
	db := dbConnection()
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO power(user, channel, w) values (UUID_TO_BIN(?, true),?,?)",
		user.Uuid,
		ap.Generator,
		int(ap.Data[1].Tuples[0][1]),
	)
	return err
}

func dbPutEMSEnergy(ap OpenEMSAPIResponse, user *User) error {
	err := validateStructure(ap)
	if err != nil {
		return err
	}
	db := dbConnection()
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO energy(user, channel, wh) values (UUID_TO_BIN(?, true),?,?)",
		user.Uuid,
		ap.Address,
		ap.Value,
	)
	return err
}

func dbPutSMEnergy(ap vzloggerAPIResponse, user *User) error {
	err := validateStructure(ap)
	if err != nil {
		return err
	}
	db := dbConnection()
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO energy(user, channel, wh) values (UUID_TO_BIN(?, true),?,?)",
		user.Uuid,
		ap.Generator,
		int(ap.Data[0].Tuples[0][1]),
	)
	return err
}

func dbPutTransaction(transaction TransactionDebiricated) error {
	err := validateStructure(transaction)
	if err != nil {
		return err
	}
	db := dbConnection()
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO transactions(uuid, seller, buyer, amount, price) values (UUID_TO_BIN(?, true), UUID_TO_BIN(?, true), UUID_TO_BIN(?, true), ?, ?)",
		transaction.Uuid,
		transaction.Seller,
		transaction.Buyer,
		transaction.Amount,
		transaction.Price,
	)
	return err
}

func dbSetActive(user *User, active bool) error {
	var integer = 0
	if active {
		integer = 1
	}
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"UPDATE users SET active = ? WHERE uuid = (UUID_TO_BIN(?, true));",
		integer,
		user.Uuid,
	)
	return err
}

func dbSetMatched(user *User, matched bool) error {
	var integer = 0
	if matched {
		integer = 1
	}
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"UPDATE users SET matched = ? WHERE uuid = (UUID_TO_BIN(?, true));",
		integer,
		user.Uuid,
	)
	return err
}

func dbResetMatched() error {
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"UPDATE users SET matched = 0 WHERE matched = 1;")
	return err
}

func dbPutExample() error {
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"INSERT INTO power (user, channel, w) VALUES (UUID_TO_BIN('249d442a-d5fc-4c65-bb43-7f21d7eecfd4', true), 'disposablePower', 180), (UUID_TO_BIN('44146e78-0d7a-4c2f-b2e4-f3fcccd29175', true), 'disposablePower', 3540), (UUID_TO_BIN('7d61316c-1456-452c-ac41-a373756b38a3', true), 'disposablePower', 120), (UUID_TO_BIN('153f948a-fa4a-446f-aa13-fa70be3f15fb', true), 'disposablePower', 600), (UUID_TO_BIN('03aa9c1b-49db-4f19-9fda-c46928afe320', true), 'disposablePower', 480), (UUID_TO_BIN('8a970131-b6e1-4767-a446-225fdf89d3e1', true), 'disposablePower', 600), (UUID_TO_BIN('1865ce5f-f865-43ba-bd9c-5c18f7f4f17e', true), 'disposablePower', 480), (UUID_TO_BIN('f0676026-f7e0-40d8-9102-79beb81716c4', true), 'disposablePower', 3000), (UUID_TO_BIN('ac3c7dea-9093-4b0c-90b1-b3350322d673', true), 'disposablePower', 2400), (UUID_TO_BIN('3de6fe02-ca22-48f9-8735-b093797e4e53', true), 'disposablePower', 540), (UUID_TO_BIN('470236a9-0045-4798-8f0b-1a6a73cfdce4', true), 'disposablePower', 420), (UUID_TO_BIN('d23b4d4f-78ea-4d9b-af7b-18e25c020fcc', true), 'disposablePower', 480), (UUID_TO_BIN('0d0c672a-bc57-40cb-8cc7-d2596bfddab0', true), 'disposablePower', 180);")
	return err
}

// ------------------------------------New Implementations--------------------------------------

type Pow struct {
	Time   time.Time `json:""`
	userId string    `json:""`
	w      float64   `json:"value"`
}

type Usr struct {
	Uuid               string         `db:"uuid" json:"uuid"`
	Name               string         `db:"name" json:"name"`
	PublicKey          string         `db:"public_key" json:"public_key"`
	Plz                int            `db:"plz" json:"plz"`
	Email              string         `db:"email" json:"email"`
	Password           string         `db:"password" json:"password"`
	Iban               string         `db:"iban" json:"iban"`
	Joindate           time.Time      `db:"joindate" json:"joindate"`
	ChainConsumptionId string         `db:"chainconsumpitonid" json:"chainconsumpitonid"`
	ChainFeedingId     string         `db:"chainfeedingid" json:"chainfeedingid"`
	Active             int            `db:"active" json:"active"`
	Url                sql.NullString `db:"url" json:"url"`
	//added from UserSpecInfo
	PostalCode         string  `db:"postalcode" json:"postalcode"`
	City               string  `db:"city" json:"city"`
	Address            string  `db:"address" json:"address"`
	SolarPowerCapacity float64 `db:"solarpowercapacity" json:"solarpowercapacity"`
	PowerStorage       bool    `db:"powerstorage" json:"powerstorage"`
	PsCapacity         float64 `db:"pscapcity" json:"pscapcity"`
	//Timestamp          time.Time `db:"timestamp" json:"timestamp"`
	//WAmount float64 `db:"wamount" json:"wamount"`
}

// TODO Test implementation
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

func dbWriteDataToConsumingTable(val float64, usr *Usr) error {
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"INSERT INTO consumption(userid, postalcode,city, address, wamount) values (UUID_TO_BIN(?, true),?,?,?,?)",
		usr.Uuid,
		usr.Plz,
		usr.City,
		usr.Address,
		val,
	)
	if err != nil {
		fmt.Println(err)
	}
	return err

}

func dbWriteDataToFeedingTable(value float64, usr *Usr) error {
	db := dbConnection()
	defer db.Close()

	_, err := db.Exec(
		"INSERT INTO feeding(userid, postalcode, city, address, solarpowercapacity, powerstorage, pscapacity, wamount) values (UUID_TO_BIN(?, true),?,?,?,?,?,?,?)",
		usr.Uuid,
		usr.Plz,
		usr.City,
		usr.Address,
		usr.SolarPowerCapacity,
		usr.PowerStorage,
		usr.PsCapacity,
		value,
	)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func dbReadLastAmountOfPower(tableName string) float64 {
	db := dbConnection()
	var wVal float64

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
	results, err := db.Query("SELECT time as t, userid as uid, wamount as w FROM " + tableName + " ORDER BY time DESC LIMIT 1;")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	//Results
	for results.Next() {
		err = results.Scan(&transaction.Time, &transaction.Uuid, &transaction.Amount)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
	return transaction
}

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

func getWattOutOfVzLoggerData(ap vzloggerAPIResponse) (float64, error) {
	err := validateStructure(ap)
	if err != nil {
		fmt.Println(err)
		return 0.0, err
	}
	return ap.Data[1].Tuples[0][1], err
}

//--------------------------------------------------------------------------------------------
