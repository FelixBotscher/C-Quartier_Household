package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"math"
	"time"
)

//Controller

func main() {
	c := cron.New()
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
	//----------------_Test_Variables_--------------------
	//nString := sql.NullString{"https://127.0.0.1:8081", true} normal
	nString := sql.NullString{"https://127.0.0.1:8080", true} //testing
	testUserFelix := Usr{Uuid: "249d442a-d5fc-4c65-bb43-7f21d7eecfd4", Name: "Felix", PublicKey: "key", Plz: 80807, Email: "fbotscher@web.de", Password: "pw", Iban: "De45....", Joindate: time.Now(), Chainid: "fb_chainId", Active: 1, Url: nString, PostalCode: "80807", City: "München", Address: "Max-Bill-Straße 67", SolarPowerCapacity: 10.0, PowerStorage: true, PsCapacity: 3.6}

	//Start cron with scheuled jobs
	fmt.Println("Start cron")
	c.Start()
	printCronEntries(c.Entries())
	err := scheduleSM(c, &testUserFelix)
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(time.Duration(1<<63 - 1))
	}
}

func scheduleSM(scheduler *cron.Cron, user *Usr) error {
	//tracks current value, if changes greater > 1 kW -> store data into corresponding table
	//when value >=0 save into consuming
	//when value <= 0 save into feeding
	scheduler.AddFunc("@every 2s", func() {
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
		var lastConsumerDbValue = 0.0
		var lastFeedingDbValue = 0.0
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

		// TODO: Write Changes to Central Server
		// store changes only when they have a real impact, here > 1 kW
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
