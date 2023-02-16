package main

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"time"
)

//Controller

func main() {
	c := cron.New()
	var users = dbGetUsers()
	argsWithProg := os.Args
	if len(argsWithProg) == 2 {
		if argsWithProg[1] == "--no-query" {
			runWithoutQuery()
			match()
		}
	} else {
		runWithQuery(users, c)
		matchContinuous(c)

		// Start cron with scheduled jobs
		fmt.Println("Start cron")
		c.Start()
		printCronEntries(c.Entries())

		//Main thread sleeps forever
		for {
			time.Sleep(time.Duration(1<<63 - 1))
		}
	}
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
