package main

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"time"
)

type Power struct {
	Time    time.Time `json:""`
	User    string    `json:""`
	Channel string    `json:"address"`
	w       int       `json:"value"`
	Status  int       `json:"status"`
}

type Energy struct {
	Time    time.Time `json:""`
	User    string    `json:""`
	Channel string    `json:"address"`
	wh      int       `json:"value"`
	Status  int       `json:"status"`
}

type User struct {
	Uuid      string         `db:"uuid" json:"uuid"`
	Name      string         `db:"name" json:"name"`
	Seller    int            `db:"seller" json:"seller"`
	PublicKey string         `db:"public_key" json:"public_key"`
	Plz       int            `db:"plz" json:"plz"`
	Email     string         `db:"email" json:"email"`
	Password  string         `db:"password" json:"password"`
	Iban      string         `db:"iban" json:"iban"`
	Joindate  time.Time      `db:"joindate" json:"joindate"`
	Chainid   string         `db:"chainid" json:"chainid"`
	Active    int            `db:"active" json:"active"`
	Minprice  int            `db:"minprice" json:"minprice"`
	Maxprice  int            `db:"maxprice" json:"maxprice"`
	Matched   int            `db:"matched" json:"matched"`
	Url       sql.NullString `db:"url" json:"url"`
}

type Transaction struct {
	Uuid   string    `json:"uuid"`
	Seller string    `json:"seller"`
	Buyer  string    `json:"buyer"`
	Time   time.Time `json:"time"`
	Amount int       `json:"amount"`
	Price  int       `json:"price"`
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

type Buyer struct {
	User   *User
	Energy int
	Score  int
}

type Seller struct {
	User   *User
	Energy int
	Score  int
}

type Matchset struct {
	unmatchedBuyers  []Buyer
	unmatchedSellers []Seller
	matches          []Match
}

type Match struct {
	buyer  *Buyer
	seller *Seller
	Energy int
}

func validateStructure(v interface{}) error {
	validate := validator.New()
	return validate.Struct(v)
}
