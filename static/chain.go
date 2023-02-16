package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.db.in.tum.de/c-chain/ccf-go/ccf"
	"gitlab.db.in.tum.de/c-chain/ccf-go/ccf/apimodels"
	"log"
	"os"
	"time"
)

func main() {
	t := InitTransactionManager()
	var users = dbGetUsers()
	for i := range users {
		user := &users[i]
		//Create Chain
		chain, _ := t.CreateChain(user.Chainid, user.Name, false, true, false)
		cuuid := chain.ChainID()
		user.Chainid = cuuid.String()

		//Read Chain
		ci, _ := t.GetChain(*cuuid)
		name := ci.Name()
		fmt.Println("Chain for user " + *name + "created.")
	}

	var transactions = dbGetTransactions()
	for i := range transactions {
		transaction := &transactions[i]
		for j := range users {
			user := &users[j]
			if transaction.Buyer == user.Uuid {
				b, _ := json.Marshal(transaction)
				ta_created, err := t.CreateTransaction(uuid.Must(uuid.Parse(user.Chainid)), b)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println("Transaction of buyer " + user.Name + " with ID " + ta_created.DataToHash.Request.DataToHash.TransactionID.String() + " written to buyers chain.")
			}
			if transaction.Seller == user.Uuid {
				b, _ := json.Marshal(transaction)
				ta_created, err := t.CreateTransaction(uuid.Must(uuid.Parse(user.Chainid)), b)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println("Transaction of seller " + user.Name + " with ID " + ta_created.DataToHash.Request.DataToHash.TransactionID.String() + " written to sellers chain.")
			}
		}
	}
	t.Close()
}

func InitTransactionManager() *TAM {
	session, err := ccf.StartSession("https://localhost", "RSA_WITH_AES_128_CBC", "keys")
	if err != nil {
		log.Fatal("Error Init CCF Service", err.Error())
	}
	return &TAM{
		ccf:       session,
		publicKey: session.PublicKeyB,
		signKey:   session.SignedKey,
		expires:   time.Now().Add(time.Minute * 10),
	}
}

func (T *TAM) CreateChain(name string, desc string, closed bool, publicRead bool, publicWrite bool) (*apimodels.ChainInfo, error) {
	return T.ccf.CreateChain(name, desc, closed, publicRead, publicWrite)
}

func (T *TAM) GetChain(chainID uuid.UUID) (*apimodels.ChainInfo, error) {
	return T.ccf.GetChain(chainID)
}

func (T *TAM) CreateTransaction(chainID uuid.UUID, payload []byte) (*apimodels.Transaction, error) {
	return T.ccf.CreateTransaction(chainID, T.publicKey, T.signKey, payload, 30, false, true)
}

func (T *TAM) GetTransaction(chainID uuid.UUID, transactionID uuid.UUID) (*apimodels.Transaction, error) {
	return T.ccf.GetTransaction(chainID, transactionID)
}

func (T *TAM) Close() error {
	return T.ccf.TerminateSession()
}

type TAM struct {
	ccf       *ccf.Service
	publicKey []byte
	signKey   []byte
	expires   time.Time
}

type Transaction struct {
	Uuid   string    `db:"uuid" json:"uuid"`
	Seller string    `db:"seller" json:"seller"`
	Buyer  string    `db:"buyer" json:"buyer"`
	Time   time.Time `db:"time" json:"time"`
	Amount int       `db:"amount" json:"amount"`
	Price  int       `db:"price" json:"price"`
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

type ConfigSQL struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DB       string `json:"dbname"`
}

func dbGetTransactions() []Transaction {
	config := loadConfigSQL("../config.json")
	info := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, 3306, config.DB)
	db, err := sqlx.Connect("mysql", info)

	var transactions []Transaction
	err = db.Select(&transactions, "SELECT BIN_TO_UUID(uuid, true) as uuid, BIN_TO_UUID(uuid, true) as seller, BIN_TO_UUID(uuid, true) as buyer, time, amount, price FROM transactions;")
	if err != nil {
		return transactions
	}
	return transactions
}

func dbGetUsers() []User {
	config := loadConfigSQL("../config.json")
	info := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, 3306, config.DB)
	db, err := sqlx.Connect("mysql", info)

	var users []User
	err = db.Select(&users, "SELECT BIN_TO_UUID(uuid, true) as uuid, name, seller, public_key, plz, email, password, iban, joindate, chainid, active, minprice, maxprice, matched, url FROM users;")
	if err != nil {
		return users
	}
	return users
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
