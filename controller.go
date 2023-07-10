package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alics/go-json-config/jsonconfig"
	emitter "github.com/emitter-io/go"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gitlab.db.in.tum.de/c-chain/ccf-go/ccf/apimodels"
	"io/ioutil"
	"log"
	"math"
	"time"
)

var actions []Action
var action Action

func main() {
	c := cron.New()
	log.Println("Loading configuration")
	conf := getConfig()
	log.Println("Initializes services...")
	//-------------------User of this R-----------------------------------------
	//TODO: change to getUserConfig()
	//Url is for vzlogger_ip_adress: original port 8081, test port 8080
	userFelix := conf.User
	//------------------C-Chain inits------------------------------------------
	t := InitTransactionManager()
	log.Println("Transactionmanager successfully initialized.")
	//Create Chain consumption
	chainConsumption, _ := t.CreateChain(userFelix.Name+"ConsumptionChain", userFelix.Name, false, true, false)
	log.Println("Chain Consumption successfully initialized with uuid: " + chainConsumption.ChainID().String())
	//Create Chain feeding
	chainFeeding, _ := t.CreateChain(userFelix.Name+"FeedingChain", userFelix.Name, false, true, false)
	log.Println("Chain feeding successfully initialized with uuid: " + chainFeeding.ChainID().String())
	//Read Chain consumption
	//cfeedingi, _ := t.GetChain(*c_feeding_uuid)
	//fmt.Println("Chain 'consumption' created " + *cfeedingi.Name())

	//Create Chain changes
	chainChanges, _ := t.CreateChain(userFelix.Name+"ChangesChain", userFelix.Name, false, true, false)
	log.Println("Chain changes successfully initialized with uuid: " + chainChanges.ChainID().String())
	log.Println("C-Chain services successfully initialized.")

	//Update Config file with new chain Ids
	updateConfigFile(conf, chainConsumption.ChainID().String(), chainFeeding.ChainID().String(), chainChanges.ChainID().String())
	//Set current conf again
	conf = getConfig()
	//-----------------MQTT----------------------------------------------
	//var bconfig = getBrokerConfig()
	client := initEmitter(&conf.Broker)
	log.Println("Emitter broker successfully initialized.")
	log.Println("Initialization of services done.")

	//Start cron with scheduled jobs
	log.Println("Starting server with jobs")
	c.Start()
	printCronEntries(c.Entries())
	scheduleJobs(c, &userFelix, t, client, &conf.Broker)
	for { // endless loop
		time.Sleep(time.Duration(1<<63 - 1))
	}
}

func scheduleJobs(scheduler *cron.Cron, user *User, tam *TAM, c emitter.Emitter, bconf *Broker) {
	//tracks current value, if changes greater > 1 kW -> store data into corresponding table
	//when value >=0 save into consuming
	//when value <= 0 save into feeding
	scheduler.AddFunc("@every 2s", func() {
		var lastConsumerDbValue = 0
		var lastFeedingDbValue = 0
		var currentValue = 0
		var err error
		var realdif = 0
		//request current value
		if user.PowerStorage { // user can feed or consum
			result, err := getEMS(user, "meter1/ActivePower")
			if err != nil {
				fmt.Println("Error reading Active Power")
				panic(err)
			}
			currentValue = result.Value
		} else { // user just consumes
			re, err := getSMLocal()
			if err != nil {
				fmt.Println("Error while getting local SM data. " + err.Error())
				return
			}
			currentValue, err = getWattOutOfVzLoggerData(re)
			if err != nil {
				fmt.Println("Error with format of local SM response. " + err.Error())
				return
			}
		}

		// get last entries from db
		conTableEmpty := checkEmptinessOfTable("consumption")
		fedTableEmpty := checkEmptinessOfTable("feeding")
		if !conTableEmpty { // table != nil
			lastConsumerDbValue = dbReadLastAmountOfPower("consumption") // 0.0 or > 0
		}
		if !fedTableEmpty {
			lastFeedingDbValue = dbReadLastAmountOfPower("feeding") // 0.0 or < 0
		}

		var lastEntryMadeValue = 0
		//Find out which is the last value in the DB
		if lastConsumerDbValue == 0 {
			lastEntryMadeValue = lastFeedingDbValue
		} else {
			lastEntryMadeValue = lastConsumerDbValue
		}

		//var difference = math.Abs(float64(lastEntryMadeValue - currentValue))
		// if difference >= 1000...

		//calculate real diff
		realdif = calRealDiff(lastEntryMadeValue, currentValue)

		// store changes only when they have a real impact, here > 1 kW
		// evaluates and writes data to tables +
		//  writes transactions from tables to chains
		if math.Abs(float64(realdif)) >= 1000 || (conTableEmpty && fedTableEmpty) {
			switch currentValue >= 0 {
			case false:
				if lastEntryMadeValue < 0 {

					err = dbWriteDataToFeedingTable(currentValue, user)
					if err != nil {
						fmt.Println("Error writing data to Feeding Table " + err.Error())
						return
					}
					//write last Transaction out of feeding table into feeding chain
					writeDataToChain(dbReadLastEntry("feeding"), user, tam)
				} else {
					err = dbWriteDataToFeedingTable(currentValue, user)
					if err != nil {
						fmt.Println("Error writing data to Feeding Table " + err.Error())
						return
					}
					err = dbWriteDataToConsumingTable(0, user)
					if err != nil {
						fmt.Println("Error writing data to Consuming Table " + err.Error())
						return
					}
					//write data to chains
					writeDataToChain(dbReadLastEntry("feeding"), user, tam)
					writeDataToChain(dbReadLastEntry("consumption"), user, tam)
				}
				break
			case true:
				if lastEntryMadeValue < 0 {
					err = dbWriteDataToConsumingTable(currentValue, user)
					if err != nil {
						fmt.Println("Error writing data to Consuming Table " + err.Error())
						return
					}
					err = dbWriteDataToFeedingTable(0, user)
					if err != nil {
						fmt.Println("Error writing data to Feeding Table " + err.Error())
						return
					}
					//write data to chains
					writeDataToChain(dbReadLastEntry("feeding"), user, tam)
					writeDataToChain(dbReadLastEntry("consumption"), user, tam)
				} else {
					err = dbWriteDataToConsumingTable(currentValue, user)
					if err != nil {
						fmt.Println("Error writing data to Consuming Table " + err.Error())
						return
					}
					//write data to chains
					writeDataToChain(dbReadLastEntry("consumption"), user, tam)
				}
				break
			}
		}
		//write changes to Changes table
		err = dbWriteDataToChangesTable(currentValue, realdif, user)
		if err != nil {
			fmt.Println("Failed to write Data to changes table" + err.Error())
		}
		//write Change from db to chain
		tc := writeDataToChangesChain(dbReadLastChangesEntry(), user, tam)

		//get Changes Chain element and create ChangeStatus
		changeCId, err := tam.GetTransaction(uuid.MustParse(user.ChainChangesId), *tc.DataToHash.Request.DataToHash.TransactionID)
		if err != nil {
			log.Println("Failed to load ChangesTransaction from changes Chain " + err.Error())
		}
		cs := ChangeStatus{Userid: user.Uuid, ChangeTransaction: changeCId, OldVal: lastEntryMadeValue, CurrentVal: currentValue, WDif: realdif}
		//Publish Changes to channel
		err = sendChanges(c, bconf, cs)
		if err != nil {
			log.Println("Error sending changes via broker. ")
			panic(err)
		}
		if user.PowerStorage { //create BatteryStatus if user can sell energy and publish to channel
			err = sendFeederStatus(c, bconf, getBatteryStatus(user))
			if err != nil {
				panic(err)
			}
		}
	})
	//Check actions array
	scheduler.AddFunc("@every 1s", func() {
		filterActions(user)
		var x Action
		//apply actions from Action request

		if action != x { //no action?
			//check if turn on is requested:
			if action.DischargePower > 0 { //Discharging requested?
				setEMS(user, "ess0/ActivePower", action.DischargePower)
			}
			// Turn on/off of relay
			//updatePinLocal(action.TurnOnRelay)
			updatePinLocalDummy(action.TurnOnRelay)

			//Load battery?
			if action.LoadBattery { // No discharge
				setEMS(user, "ess0/ActivePower", 0)
			}
		}
	})
}

//func scheduleSMOld(scheduler *cron.Cron, user *User) error {
//	//tracks current value, if changes greater > 1 kW -> store data into corresponding table
//	//when value >=0 save into consuming
//	//when value <= 0 save into feeding
//	scheduler.AddFunc("@every 2s", func() {
//		var lastConsumerDbValue = 0.0
//		var lastFeedingDbValue = 0.0
//		var response, err = getSMLocal()
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		currentValue, err := getWattOutOfVzLoggerData(response)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		// get last entries from db
//		var conTableEmpty = checkEmptinessOfTable("consumption")
//		var fedTableEmpty = checkEmptinessOfTable("feeding")
//		if !conTableEmpty { // table != nil
//			lastConsumerDbValue = dbReadLastAmountOfPower("consumption") // 0.0 or > 0
//		}
//		if !fedTableEmpty {
//			lastFeedingDbValue = dbReadLastAmountOfPower("feeding") // 0.0 or < 0
//		}
//		var lastEntryMadeValue = 0.0
//		//Find out which is the last value in the DB
//		if lastConsumerDbValue == 0.0 {
//			lastEntryMadeValue = lastFeedingDbValue
//		} else {
//			lastEntryMadeValue = lastConsumerDbValue
//		}
//		//var difference, tableIndicator = calculateDifferenceAndTargetTable(currentValue, lastEntryMadeValue)
//		var difference = math.Abs(lastEntryMadeValue - currentValue)
//
//		// store changes only when they have a real impact, here > 1 kW
//		// evaluates and writes data to tables
//		// writes transactions / data to chains
//		if difference >= 1000 || (conTableEmpty && fedTableEmpty) {
//			switch currentValue >= 0.0 {
//			case false:
//				if lastEntryMadeValue < 0.0 {
//
//					err = dbWriteDataToFeedingTable(currentValue, user)
//					if err != nil {
//						return
//					}
//
//				} else {
//					err = dbWriteDataToFeedingTable(currentValue, user)
//					if err != nil {
//						return
//					}
//					err = dbWriteDataToConsumingTable(0.0, user)
//					if err != nil {
//						return
//					}
//
//				}
//				break
//			case true:
//				if lastEntryMadeValue < 0.0 {
//					err = dbWriteDataToConsumingTable(currentValue, user)
//					if err != nil {
//						return
//					}
//					err = dbWriteDataToFeedingTable(0.0, user)
//					if err != nil {
//						return
//					}
//				} else {
//					err = dbWriteDataToConsumingTable(currentValue, user)
//					if err != nil {
//						return
//					}
//				}
//				break
//			}
//		}
//	})
//	return nil
//}

// calRealDiff returns the real difference of two integer values independent of their sign
func calRealDiff(last int, current int) int {
	var l = float64(last)
	var c = float64(current)
	var result float64

	if l <= c {
		if c < 0 {
			result = math.Abs(l - c)
		} else if c == 0 {
			result = l
		} else if l <= 0 && c >= 0 {
			result = math.Abs(l) + c
		} else {
			result = c + l
		}
	} else {
		if l <= 0 {
			result = c - l
		} else if l == 0 {
			result = c
		} else if c <= 0 && l >= 0 {
			result = (math.Abs(c) + l) * -1.0
		} else {
			result = (l - c) * -1.0
		}
	}
	return int(result)
}

func updateConfigFile(conf Config, consumption string, feeding string, changes string) {
	copyDb := Db{
		Host:     conf.Db.Host,
		User:     conf.Db.User,
		Password: conf.Db.Password,
		Dbname:   conf.Db.Dbname,
	}
	copyChain := Chain{
		Host:     conf.Chain.Host,
		Encrypt:  conf.Chain.Encrypt,
		Key_file: conf.Chain.Key_file,
	}
	updatedUser := User{
		Uuid:               conf.User.Uuid,
		Name:               conf.User.Name,
		Seller:             conf.User.Seller,
		PublicKey:          conf.User.PublicKey,
		Email:              conf.User.Email,
		Password:           conf.User.Password,
		Iban:               conf.User.Iban,
		Joindate:           conf.User.Joindate,
		ChainConsumptionId: consumption,
		ChainFeedingId:     feeding,
		ChainChangesId:     changes,
		Active:             conf.User.Active,
		Url:                conf.User.Url,
		PostalCode:         conf.User.PostalCode,
		City:               conf.User.City,
		Address:            conf.User.Address,
		PowerStorage:       conf.User.PowerStorage,
		PsCapacity:         conf.User.PsCapacity,
	}
	copyBroker := Broker{
		BaseAddress:    conf.Broker.BaseAddress,
		TargetChannel1: conf.Broker.TargetChannel1,
		KeyChannel1:    conf.Broker.KeyChannel1,
		TargetChannel2: conf.Broker.TargetChannel2,
		KeyChannel2:    conf.Broker.KeyChannel2,
		TargetChannel3: conf.Broker.TargetChannel3,
		KeyChannel3:    conf.Broker.KeyChannel3,
	}
	newConf := Config{
		Db:     copyDb,
		Chain:  copyChain,
		User:   updatedUser,
		Broker: copyBroker,
	}
	newConfigfile, _ := json.MarshalIndent(newConf, "", " ")
	_ = ioutil.WriteFile("config.json", newConfigfile, 0644)
	log.Println("config.json updated with chain ids")
}

func writeDataToChain(transaction Transaction, user *User, tam *TAM) {
	tablename := "consumption"
	varUuid := user.ChainConsumptionId
	if transaction.AbsVal < 0 {
		tablename = "feeding"
		varUuid = user.ChainFeedingId
	}
	b, _ := json.Marshal(transaction)
	taCreated, err := tam.CreateTransaction(uuid.Must(uuid.Parse(varUuid)), b)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	log.Println("Transaction of user " + user.Name + " with ID " + taCreated.DataToHash.Request.DataToHash.TransactionID.String() + " written to" + tablename + "chain.")
}

func writeDataToChangesChain(ct ChangesTransaction, user *User, tam *TAM) *apimodels.Transaction {
	b, _ := json.Marshal(ct)
	taCreated, err := tam.CreateTransaction(uuid.Must(uuid.Parse(user.ChainChangesId)), b)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	log.Println("ChangesTransaction of user " + user.Name + " with ID " + taCreated.DataToHash.Request.DataToHash.TransactionID.String() + " written to changes chain.")
	return taCreated
}

func filterActions(user *User) {
	for _, val := range actions {
		if val.User == user.Uuid {
			action = val
			return
		}
	}
	actions = nil
}

func getUserConfig() *User {
	//TODO:implement
	return nil
}

func getConfig() Config {
	f, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("Failed to load confi.json file")
		panic(err)
	}
	config := Config{}
	err = json.Unmarshal([]byte(f), &config)
	if err != nil {
		fmt.Println("Failed to load config.json bytestream into struct of type Config")
		panic(err)
	}
	return config
}

func getBrokerConfig() *Broker {
	var bconfig *Broker
	err := jsonconfig.Bind(&bconfig, "broker") //adapted from package https://github.com/alics/go-json-config
	if err != nil {
		fmt.Println("Couldn't read json config file section \"broker\"")
		panic(err)
	}
	return bconfig
}

func getBatteryStatus(user *User) FeederStatus {
	var statusOk bool = false
	var dummyMinDischarge = 200  //TODO: specify in config.json file
	var dummyMaxDischarge = 5000 //TODO: specify in config.json file
	emsRestCapacityB, err := getEMS(user, "ess0/AllowedDischargePower")
	if err != nil {
		fmt.Println(err)
	}
	emsConsumptionHH, err := getEMS(user, "meter2/ActivePower")
	if err != nil {
		fmt.Println(err)
	}
	emsProductionPV, err := getEMS(user, "meter1/ActivePower")
	if err != nil {
		fmt.Println(err)
	}
	if initCheckSeller(user) == nil {
		statusOk = true
	}
	return FeederStatus{user.Uuid, emsConsumptionHH.Value, emsProductionPV.Value, emsRestCapacityB.Value, dummyMinDischarge, dummyMaxDischarge, statusOk}
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

//func scheduleSeller(scheduler *cron.Cron, seller User) error {
//	err := initCheckSeller(seller)
//	if err != nil {
//		fmt.Println(err)
//	}
//	scheduler.AddFunc("@every 2s", func() {
//		var soc, err = getEMS(seller, "ess0/Soc")
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		if soc.Value == 0 {
//			dbSetActive(&seller, false)
//		} else {
//			dbSetActive(&seller, true)
//		}
//	})
//	scheduler.AddFunc("@every 2s", func() {
//		var activePower, err = getEMS(seller, "ess0/ActivePower")
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		dbPutEMSPower(activePower, &seller)
//	})
//	scheduler.AddFunc("@every 2s", func() {
//		var allowedDischargePower = -1
//		var activePowerMet1 = -1
//		var activePowerMet2 = -1
//
//		var openEMSAPIREsponse, err = getEMS(seller, "ess0/AllowedDischargePower")
//		if err != nil {
//			fmt.Println(err)
//		} else {
//			dbPutEMSPower(openEMSAPIREsponse, &seller)
//			allowedDischargePower = openEMSAPIREsponse.Value
//		}
//
//		openEMSAPIREsponse, err = getEMS(seller, "meter1/ActivePower")
//		if err != nil {
//			fmt.Println(err)
//		} else {
//			dbPutEMSPower(openEMSAPIREsponse, &seller)
//			activePowerMet1 = openEMSAPIREsponse.Value
//		}
//
//		openEMSAPIREsponse, err = getEMS(seller, "meter2/ActivePower")
//		if err != nil {
//			fmt.Println(err)
//		} else {
//			dbPutEMSPower(openEMSAPIREsponse, &seller)
//			activePowerMet2 = openEMSAPIREsponse.Value
//		}
//
//		if allowedDischargePower == -1 || activePowerMet1 == -1 || activePowerMet2 == -1 {
//			fmt.Println("Not all requests for calculating disposable power succeeded. Skipping this turn.")
//			return
//		}
//		openEMSAPIREsponse.Address = "disposablePower"
//		openEMSAPIREsponse.Value = allowedDischargePower + activePowerMet1 - activePowerMet2
//		dbPutEMSPower(openEMSAPIREsponse, &seller)
//	})
//	scheduler.AddFunc("@every 2s", func() {
//		var activeProductionEnergy, err = getEMS(seller, "meter0/ActiveProductionEnergy")
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		dbPutEMSEnergy(activeProductionEnergy, &seller)
//	})
//	scheduler.AddFunc("@every 2s", func() {
//		var activeConsumptionEnergy, err = getEMS(seller, "meter0/ActiveConsumptionEnergy")
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		dbPutEMSEnergy(activeConsumptionEnergy, &seller)
//	})
//	return err
//}

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
