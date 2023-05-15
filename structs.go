package main

import (
	"github.com/go-playground/validator/v10"
	"gitlab.db.in.tum.de/c-chain/ccf-go/ccf"
	"gitlab.db.in.tum.de/c-chain/ccf-go/ccf/apimodels"
	"time"
)

//---------------- Structs for Chain handling + Transactions ---------------------

type TAM struct {
	ccf       *ccf.Service
	publicKey []byte
	signKey   []byte
	expires   time.Time
}

// for storing data into feeding and consuming chain
type Transaction struct {
	Uuid   string    `db:"uuid" json:"uuid"`
	Time   time.Time `db:"time" json:"time"`
	Userid string    `db:"userid" json:"userid"`
	AbsVal int       `db:"wchange" json:"wchange"`
}

// for storing data into the changes chain
type ChangesTransaction struct { //vormals Change
	Uuid   string    `db:"uuid" json:"uuid"`
	Time   time.Time `db:"time" json:"time"`
	Userid string    `db:"userid" json:"userid"`
	//OldVal  int       `db:"oldval" json:"oldval"`
	AbsVal  int `db:"absVal" json:"absVal"`
	WChange int `db:"wChange" json:"wChange"`
}

// structs for config.json handling-------------------
type ConfigSQL struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DB       string `json:"dbname"`
	UserId   string `json:"userId"`
}

type Db struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}
type Chain struct {
	Host     string `json:"host"`
	Encrypt  string `json:"encrypt"`
	Key_file string `json:"key_file"`
}

// type User structs excluded

type Broker struct {
	BaseAddress    string
	TargetChannel1 string
	KeyChannel1    string
	TargetChannel2 string
	KeyChannel2    string
	TargetChannel3 string
	KeyChannel3    string
}

type Config struct {
	Db     Db
	Chain  Chain //deprecated
	User   User
	Broker Broker
}

// end structs for conif.json handling------------------
type Action struct {
	Time           time.Time `json:"time"`
	Uuid           string    `json:"uuid"`
	User           string    `json:"user"`
	LoadBattery    bool      `json:"loadbattery"`
	DischargePower int       `json:"dischargepower"`
	TurnOnRelay    bool      `json:"turnonrelay"`
}

// for sending changes via emitter broker
type ChangeStatus struct {
	Userid            string //uuid not necessary due to ChangeTransaction
	ChangeTransaction *apimodels.Transaction
	OldVal            int //TODO:delete and get info out of ChangeTransaction
	CurrentVal        int //TODO:delete and get info out of ChangeTransaction
	WDif              int //TODO:delete and get info out of ChangeTransaction
}

type FeederStatus struct {
	uuid               string
	ConsumptionHH      int
	ProductionPV       int
	RestCapacityB      int
	DischargePowerMinB int
	DischargePowerMaxB int
	StateHH            bool
}

//type Power struct {
//	Time    time.Time `json:""`
//	User    string    `json:""`
//	Channel string    `json:"address"`
//	w       int       `json:"value"`
//	Status  int       `json:"status"`
//}
//
//type Energy struct {
//	Time    time.Time `json:""`
//	User    string    `json:""`
//	Channel string    `json:"address"`
//	wh      int       `json:"value"`
//	Status  int       `json:"status"`
//}

type User struct {
	Uuid               string    `db:"uuid" json:"uuid"`
	Name               string    `db:"name" json:"name"`
	Seller             string    `db:"seller" json:"seller"`
	PublicKey          string    `db:"public_key" json:"public_key"`
	Email              string    `db:"email" json:"email"`
	Password           string    `db:"password" json:"password"`
	Iban               string    `db:"iban" json:"iban"`
	Joindate           time.Time `db:"joindate" json:"joindate"`
	ChainConsumptionId string    `db:"chainconsumpitonid" json:"chainconsumpitonid"`
	ChainFeedingId     string    `db:"chainfeedingid" json:"chainfeedingid"`
	ChainChangesId     string    `db:"chainchangesid" json:"chainchangesid"`
	Active             int       `db:"active" json:"active"`
	Url                string    `db:"url" json:"url"`
	//added from UserSpecInfo
	PostalCode   string  `db:"postalcode" json:"postalcode"`
	City         string  `db:"city" json:"city"`
	Address      string  `db:"address" json:"address"`
	PowerStorage bool    `db:"powerstorage" json:"powerstorage"`
	PsCapacity   float64 `db:"pscapcity" json:"pscapcity"`
}

type OpenEMSAPIResponse struct {
	Address    string `json:"address"`
	Type       string `json:"type"`
	AccessMode string `json:"accessMode"`
	Text       string `json:"text"`
	Unit       string `json:"unit"`
	Value      int    `json:"value"`
}

type vzloggerAPIResponse struct {
	Version   string `json:"version"`
	Generator string `json:"generator"`
	Data      []struct {
		UUID     string      `json:"uuid"`
		Last     int64       `json:"last"`
		Interval int         `json:"interval"`
		Protocol string      `json:"protocol"`
		Tuples   [][]float64 `json:"tuples"`
	} `json:"data"`
}

func validateStructure(v interface{}) error {
	validate := validator.New()
	return validate.Struct(v)
}
