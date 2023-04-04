package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"math"
	"time"
)

//Controller

func main() {
	c := cron.New()
	//Check Key file and format first
	/*var pwd = "chain"
	var pwdB = []byte(pwd)
	var pemPublicKey = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoAKCF6zyrD3cXSy0ogVj\nuUwjwMGNyzmxc5LDRSRFKpFfsynyBamktOPnhhpisOXPF5G1237GUDd21bcZ6iVq\nvFV0wANsRx8OWQgdZfFiJhiuWMA5kUIri3ERKscmsh+xyf0/N7iEZwSY16XqBW0Y\nTzNMIe/aqV4KnrUCGE3KWj3DAAusOkWNCOyIIKAVdiufcv+G7vZFILNDcUo4bq5p\n9t4jf3K/EXBWm6YRSOber9lrzjAvHWf0iiFmS18mmMAg3yG1pSMBERVccz/M4GVp\nH/KqA33MA72QxU8PeP35SHO6Ys+qLD3tTiraAtntoehmnIkMAp5FkU5DsVe2N9hy\nBQIDAQAB\n-----END PUBLIC KEY-----\n"
	fmt.Println(pemPublicKey)
	block, _ := pem.Decode([]byte(pemPublicKey))
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}
	fmt.Println(block)
	pkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	rsaKey, ok := pkey.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("got unexpected key type: %T", pkey)
	}

	key, err := LoadPrivate("./key")
	fmt.Println()
	*/
	t := InitTransactionManager()
	fmt.Println("Transactionsmanager successfully initialized")

	//nString := sql.NullString{"https://127.0.0.1:8081", true} normal
	nString := sql.NullString{"https://127.0.0.1:8080", true} //testing with python script
	testUserFelix := Usr{Uuid: "249d442a-d5fc-4c65-bb43-7f21d7eecfd4", Name: "Felix", PublicKey: "key", Plz: 80807, Email: "fbotscher@web.de", Password: "pw", Iban: "De45....", Joindate: time.Now(), ChainConsumptionId: "fb_chain_consumption_5849", ChainFeedingId: "fb_chain_feeding_4389", Active: 1, Url: nString, PostalCode: "80807", City: "München", Address: "Max-Bill-Straße 67", SolarPowerCapacity: 10.0, PowerStorage: true, PsCapacity: 3.6}

	//Create Chain consumption
	chain_consumption, _ := t.CreateChain(testUserFelix.ChainConsumptionId, testUserFelix.Name, true, true, false)
	c_consumption_uuid := chain_consumption.ChainID()
	testUserFelix.ChainConsumptionId = c_consumption_uuid.String()
	//Read Chain consumption
	cconsumptioni, _ := t.GetChain(*c_consumption_uuid)
	fmt.Println("Chain 'consumption' created " + *cconsumptioni.Name())

	//Create Chain feeding
	chain_feeding, _ := t.CreateChain(testUserFelix.ChainFeedingId, testUserFelix.Name, true, true, false)
	c_feeding_uuid := chain_feeding.ChainID()
	testUserFelix.ChainFeedingId = c_feeding_uuid.String()
	//Read Chain consumption
	cfeedingi, _ := t.GetChain(*c_feeding_uuid)
	fmt.Println("Chain 'consumption' created " + *cfeedingi.Name())

	//var users = dbGetUsers()
	//argsWithProg := os.Args
	//if len(argsWithProg) == 2 {
	//	if argsWithProg[1] == "--no-query" {
	//		runWithoutQuery()
	//		match()
	//	}
	//} else {
	//	runWithQuery(users, c)
	//	matchContinuous(c)
	//
	//	// Start cron with scheduled jobs
	//	fmt.Println("Start cron")
	//	c.Start()
	//	printCronEntries(c.Entries())
	//
	//	//Main thread sleeps forever
	//	for {
	//		time.Sleep(time.Duration(1<<63 - 1))
	//	}
	//}

	/*
		//Start API Server
		initPin()
		router := gin.Default()
		router.POST("/pin", updatePin)
		router.Run("localhost:9191")

		//Test Endpoint by sending following input from a terminal line:

			//curl http://localhost:9191/pin \
			//    --include \
			//    --header "Content-Type: application/json" \
			//    --request "POST" \
			//    --data '1'
	*/

	//Start cron with scheduled jobs
	fmt.Println("Start cron")
	c.Start()
	printCronEntries(c.Entries())
	err := scheduleSM(c, &testUserFelix, t)
	//err := scheduleSM(c, &testUserFelix, t)
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(time.Duration(1<<63 - 1))
	}
}

func scheduleSM(scheduler *cron.Cron, user *Usr, tam *TAM) error {
	//tracks current value, if changes greater > 1 kW -> store data into corresponding table
	//when value >=0 save into consuming
	//when value <= 0 save into feeding
	scheduler.AddFunc("@every 2s", func() {
		var lastConsumerDbValue = 0.0
		var lastFeedingDbValue = 0.0
		var response, err = getSMLocal()
		if err != nil {
			fmt.Println(err)
			return
		}
		currentValue, err := getWattOutOfVzLoggerData(response)
		if err != nil {
			fmt.Println(err)
			return
		}
		// get last entries from db
		var conTableEmpty = checkEmptinessOfTable("consumption")
		var fedTableEmpty = checkEmptinessOfTable("feeding")
		if !conTableEmpty { // table != nil
			lastConsumerDbValue = dbReadLastAmountOfPower("consumption") // 0.0 or > 0
		}
		if !fedTableEmpty {
			lastFeedingDbValue = dbReadLastAmountOfPower("feeding") // 0.0 or < 0
		}
		var lastEntryMadeValue = 0.0
		//Find out which is the last value in the DB
		if lastConsumerDbValue == 0.0 {
			lastEntryMadeValue = lastFeedingDbValue
		} else {
			lastEntryMadeValue = lastConsumerDbValue
		}
		//var difference, tableIndicator = calculateDifferenceAndTargetTable(currentValue, lastEntryMadeValue)
		var difference = math.Abs(lastEntryMadeValue - currentValue)

		// store changes only when they have a real impact, here > 1 kW
		// evaluates and writes data to tables
		// writes transactions / data to chains
		if difference >= 1000 || (conTableEmpty && fedTableEmpty) {
			switch currentValue >= 0.0 {
			case false:
				if lastEntryMadeValue < 0.0 {

					err = dbWriteDataToFeedingTable(currentValue, user)
					if err != nil {
						return
					}
					//write last Transaction out of feeding table into feeding chain
					writeDataToChain(dbReadLastEntry("feeding"), user, tam)
				} else {
					err = dbWriteDataToFeedingTable(currentValue, user)
					if err != nil {
						return
					}
					err = dbWriteDataToConsumingTable(0.0, user)
					if err != nil {
						return
					}
					//write data to chains
					writeDataToChain(dbReadLastEntry("feeding"), user, tam)
					writeDataToChain(dbReadLastEntry("consumption"), user, tam)
				}
				break
			case true:
				if lastEntryMadeValue < 0.0 {
					err = dbWriteDataToConsumingTable(currentValue, user)
					if err != nil {
						return
					}
					err = dbWriteDataToFeedingTable(0.0, user)
					if err != nil {
						return
					}
					//write data to chains
					writeDataToChain(dbReadLastEntry("feeding"), user, tam)
					writeDataToChain(dbReadLastEntry("consumption"), user, tam)
				} else {
					err = dbWriteDataToConsumingTable(currentValue, user)
					if err != nil {
						return
					}
					//write data to chains
					writeDataToChain(dbReadLastEntry("consumption"), user, tam)
				}
				break
			}
		}
	})
	return nil
}

func scheduleSMOld(scheduler *cron.Cron, user *Usr) error {
	//tracks current value, if changes greater > 1 kW -> store data into corresponding table
	//when value >=0 save into consuming
	//when value <= 0 save into feeding
	scheduler.AddFunc("@every 2s", func() {
		var lastConsumerDbValue = 0.0
		var lastFeedingDbValue = 0.0
		var response, err = getSMLocal()
		if err != nil {
			fmt.Println(err)
			return
		}
		currentValue, err := getWattOutOfVzLoggerData(response)
		if err != nil {
			fmt.Println(err)
			return
		}
		// get last entries from db
		var conTableEmpty = checkEmptinessOfTable("consumption")
		var fedTableEmpty = checkEmptinessOfTable("feeding")
		if !conTableEmpty { // table != nil
			lastConsumerDbValue = dbReadLastAmountOfPower("consumption") // 0.0 or > 0
		}
		if !fedTableEmpty {
			lastFeedingDbValue = dbReadLastAmountOfPower("feeding") // 0.0 or < 0
		}
		var lastEntryMadeValue = 0.0
		//Find out which is the last value in the DB
		if lastConsumerDbValue == 0.0 {
			lastEntryMadeValue = lastFeedingDbValue
		} else {
			lastEntryMadeValue = lastConsumerDbValue
		}
		//var difference, tableIndicator = calculateDifferenceAndTargetTable(currentValue, lastEntryMadeValue)
		var difference = math.Abs(lastEntryMadeValue - currentValue)

		// store changes only when they have a real impact, here > 1 kW
		// evaluates and writes data to tables
		// writes transactions / data to chains
		if difference >= 1000 || (conTableEmpty && fedTableEmpty) {
			switch currentValue >= 0.0 {
			case false:
				if lastEntryMadeValue < 0.0 {

					err = dbWriteDataToFeedingTable(currentValue, user)
					if err != nil {
						return
					}

				} else {
					err = dbWriteDataToFeedingTable(currentValue, user)
					if err != nil {
						return
					}
					err = dbWriteDataToConsumingTable(0.0, user)
					if err != nil {
						return
					}

				}
				break
			case true:
				if lastEntryMadeValue < 0.0 {
					err = dbWriteDataToConsumingTable(currentValue, user)
					if err != nil {
						return
					}
					err = dbWriteDataToFeedingTable(0.0, user)
					if err != nil {
						return
					}
				} else {
					err = dbWriteDataToConsumingTable(currentValue, user)
					if err != nil {
						return
					}
				}
				break
			}
		}
	})
	return nil
}

func writeDataToChain(transaction Transaction, user *Usr, tam *TAM) {
	b, _ := json.Marshal(transaction)
	taCreated, err := tam.CreateTransaction(uuid.Must(uuid.Parse(user.ChainFeedingId)), b)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var table string = "consumption"
	if transaction.Amount < 0.0 {
		table = "feeding"
	}
	fmt.Println("Transaction of user " + user.Name + " with ID " + taCreated.DataToHash.Request.DataToHash.TransactionID.String() + " written to" + table + "chain.")
}

func evaluateAndStoreChangesToCentralServer(currentValue float64, lastValue float64, diffAbsolute float64) float64 {
	/*
		Sign table
		currentValue | last Value | result		-> How to calculate the difference?
		------------------------------------------------------------------------
		pos--------- | pos------- | pos----   	-> abs(lastValue - currentValue)
		pos--------- | neg------- | pos/neg  	-> abs(lastValue + currentValue)
		neg--------- | pos------- | pos/neg  	-> abs(lastValue + currentValue)
		neg--------- | neg------- | neg----		-> abs(lastValue - currentValue)
	*/
	switch lastValue < 0.0 {
	case false: // last = pos
		if currentValue >= 0.0 { // cur = pos
			return math.Abs(lastValue - currentValue) //definitive a positive value, true = write new value to consuming table
		} else { // cur = neg
			return math.Abs(lastValue) + currentValue
			//if lastValue+currentValue >= 0.0 {
			//	return math.Abs(lastValue) + currentValue, false // -> sum is positive, val neg write to feeding table
			//} else {
			//	return lastValue + currentValue, false // -> negative value
			//}
		}
	case true: //last = neg
		if currentValue >= 0.0 { // cur = pos
			if lastValue+currentValue >= 0.0 {
				return lastValue - currentValue // -> positive difference
			} else {
				return lastValue - currentValue // -> negative value
			}
		} else { // cur = neg
			return math.Abs(lastValue - currentValue) // definitive a negative value, false = write new value to feeding table
		}
	}
	return 0.0
}

func runWithoutQuery() (int, error) {
	dbPutExample()
	return fmt.Println("C-Power server running WITHOUT querying REST-APIs.")
}

func runWithQuery(users []User, c *cron.Cron) {
	fmt.Println("C-Power server running WITH querying REST-APIs.")
	for i := range users {
		user := &users[i]
		if user.Url.Valid == false {
			fmt.Println("No URL specified for user " + user.Name)
		} else {
			if user.Seller == 1 {
				scheduleSeller(c, user)
			} else {
				scheduleBuyer(c, user)
			}
		}
	}
}

func scheduleBuyer(scheduler *cron.Cron, buyer *User) error {
	scheduler.AddFunc("@every 2s", func() {
		var response, err = getSM(buyer)
		if err != nil {
			fmt.Println(err)
			return
		}
		dbPutSMPower(response, buyer)
		dbPutSMEnergy(response, buyer)
	})
	return nil
}

func scheduleSeller(scheduler *cron.Cron, seller *User) error {
	err := initCheckSeller(seller)
	if err != nil {
		fmt.Println(err)
	}
	scheduler.AddFunc("@every 2s", func() {
		var soc, err = getEMS(seller, "ess0/Soc")
		if err != nil {
			fmt.Println(err)
			return
		}
		if soc.Value == 0 {
			dbSetActive(seller, false)
		} else {
			dbSetActive(seller, true)
		}
	})
	scheduler.AddFunc("@every 2s", func() {
		var activePower, err = getEMS(seller, "ess0/ActivePower")
		if err != nil {
			fmt.Println(err)
			return
		}
		dbPutEMSPower(activePower, seller)
	})
	scheduler.AddFunc("@every 2s", func() {
		var allowedDischargePower = -1
		var activePowerMet1 = -1
		var activePowerMet2 = -1

		var openEMSAPIREsponse, err = getEMS(seller, "ess0/AllowedDischargePower")
		if err != nil {
			fmt.Println(err)
		} else {
			dbPutEMSPower(openEMSAPIREsponse, seller)
			allowedDischargePower = openEMSAPIREsponse.Value
		}

		openEMSAPIREsponse, err = getEMS(seller, "meter1/ActivePower")
		if err != nil {
			fmt.Println(err)
		} else {
			dbPutEMSPower(openEMSAPIREsponse, seller)
			activePowerMet1 = openEMSAPIREsponse.Value
		}

		openEMSAPIREsponse, err = getEMS(seller, "meter2/ActivePower")
		if err != nil {
			fmt.Println(err)
		} else {
			dbPutEMSPower(openEMSAPIREsponse, seller)
			activePowerMet2 = openEMSAPIREsponse.Value
		}

		if allowedDischargePower == -1 || activePowerMet1 == -1 || activePowerMet2 == -1 {
			fmt.Println("Not all requests for calculating disposable power succeeded. Skipping this turn.")
			return
		}
		openEMSAPIREsponse.Address = "disposablePower"
		openEMSAPIREsponse.Value = allowedDischargePower + activePowerMet1 - activePowerMet2
		dbPutEMSPower(openEMSAPIREsponse, seller)
	})
	scheduler.AddFunc("@every 2s", func() {
		var activeProductionEnergy, err = getEMS(seller, "meter0/ActiveProductionEnergy")
		if err != nil {
			fmt.Println(err)
			return
		}
		dbPutEMSEnergy(activeProductionEnergy, seller)
	})
	scheduler.AddFunc("@every 2s", func() {
		var activeConsumptionEnergy, err = getEMS(seller, "meter0/ActiveConsumptionEnergy")
		if err != nil {
			fmt.Println(err)
			return
		}
		dbPutEMSEnergy(activeConsumptionEnergy, seller)
	})
	return err
}

func initCheckSeller(seller *User) error {
	var response, err = getEMS(seller, "ess0/State")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if response.Value != 0 {
		return errors.New("ESS State not OK.")
	}

	response, err = getEMS(seller, "ess0/GridMode")
	if err != nil {
		fmt.Println(err)
		return err
	}
	if response.Value == 2 {
		return errors.New("ESS is off grid.")
	}

	response, err = getEMS(seller, "meter0/State")
	if err != nil {
		fmt.Println(err)
		return err
	}
	if response.Value != 0 {
		return errors.New("Grid Meter State not OK.")
	}

	response, err = getEMS(seller, "meter1/State")
	if err != nil {
		fmt.Println(err)
		return err
	}
	if response.Value != 0 {
		return errors.New("PV Meter State not OK.")
	}

	response, err = getEMS(seller, "meter2/State")
	if err != nil {
		fmt.Println(err)
		return err
	}
	if response.Value != 0 {
		return errors.New("Consumption Meter State not OK.")
	}
	return err
}

// helper
func printCronEntries(cronEntries []cron.Entry) {
	fmt.Println("Cron Info: %+v\n", cronEntries)
}
