package main

import (
	"fmt"
	"github.com/google/uuid"
	_ "github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"sort"
	"strconv"
	"time"
)

//Matcher

func match() {
	dbResetMatched()
	var matchsets = []Matchset{createMatchset(80333), createMatchset(81541)}
	for i := range matchsets {
		matchset := &matchsets[i]
		scoreAndSortBuyers(matchset)
		scoreAndSortSellers(matchset)
		fmt.Println("Begin matching")
		for j := range matchset.unmatchedSellers {
			seller := &matchset.unmatchedSellers[j]
			var buyer *Buyer
			// no more buyers
			if len(matchset.unmatchedBuyers) > 0 {
				buyer = &matchset.unmatchedBuyers[0]
			} else {
				fmt.Println("No buyers in match set.")
				break
			}
			// negative score
			if seller.Score <= 0 {
				continue
			}
			var k = 0
			for buyer.User.Maxprice < seller.User.Minprice || buyer.User.Matched == 1 {
				buyer = &matchset.unmatchedBuyers[k]
				k++
				if k == len(matchset.unmatchedBuyers) {
					break
				}
			}
			if buyer.User.Maxprice >= seller.User.Minprice && buyer.User.Matched == 0 {
				successfulMatch(matchset, buyer, seller)
			} else {
				fmt.Println("No buyer for seller " + seller.User.Name + " found.")
			}
		}
		fmt.Println("Matching done.")
	}
}

func successfulMatch(matchset *Matchset, buyer *Buyer, seller *Seller) {
	var amount = max(buyer.Energy, seller.Energy)
	if seller.User.Url.Valid {
		setEMS(seller.User, "ess0/SetActivePowerEquals", amount*60)
	}
	matchset.matches = append(matchset.matches, Match{buyer, seller, amount})
	dbSetMatched(buyer.User, true)
	dbSetMatched(seller.User, true)
	buyer.User.Matched = 1
	dbPutTransaction(Transaction{
		uuid.New().String(),
		seller.User.Uuid,
		buyer.User.Uuid,
		time.Now(),
		amount,
		seller.User.Minprice,
	})
}

func scoreAndSortSellers(matchset *Matchset) {
	for j := range matchset.unmatchedSellers {
		seller := &matchset.unmatchedSellers[j]
		seller.Score = seller.Energy * seller.User.Minprice
	}
	sort.Slice(matchset.unmatchedSellers, func(i, j int) bool {
		return matchset.unmatchedSellers[i].Score < matchset.unmatchedSellers[j].Score
	})
}

func scoreAndSortBuyers(matchset *Matchset) {
	for j := range matchset.unmatchedBuyers {
		buyer := &matchset.unmatchedBuyers[j]
		buyer.Score = buyer.Energy * buyer.User.Maxprice
	}
	sort.Slice(matchset.unmatchedBuyers, func(i, j int) bool {
		return matchset.unmatchedBuyers[i].Score < matchset.unmatchedBuyers[j].Score
	})
}

func createMatchset(plz int) Matchset {
	fmt.Println("Create Matchset for PLZ " + strconv.Itoa(plz))
	var matchset Matchset
	var users = dbGetUsers()
	var powers []Power
	for i := range users {
		user := &users[i]
		if user.Active != 1 || user.Plz != plz {
			continue
		}
		if user.Seller == 1 {
			powers = dbGetRecentDisposablePower()
			disposables := matchPowerUser(powers, user)
			predictedEnergy := predictEnergy(disposables)
			var seller Seller
			seller.User = user
			seller.Energy = predictedEnergy
			matchset.unmatchedSellers = append(matchset.unmatchedSellers, seller)
		} else {
			powers = dbGetRecentSMPower()
			activePower := matchPowerUser(powers, user)
			predictedEnergy := predictEnergy(activePower)
			var buyer Buyer
			buyer.User = user
			buyer.Energy = predictedEnergy
			matchset.unmatchedBuyers = append(matchset.unmatchedBuyers, buyer)
		}
	}
	return matchset
}

func matchPowerUser(powers []Power, user *User) []int {
	var disposables []int
	for j := range powers {
		power := &powers[j]
		if power.User == user.Uuid {
			disposables = append(disposables, power.w)
		}
	}
	return disposables
}

func predictEnergy(disposables []int) int {
	average := avg(disposables)
	var predictedEnergy = float64(average) * (1.0 / 60.0)
	var predictedEnergyRounded = int(predictedEnergy)
	return predictedEnergyRounded
}

func matchContinuous(c *cron.Cron) {
	c.AddFunc("@every 60s", func() {
		match()
	})
}

// helper
func avg(disposables []int) int {
	var sum, average int
	for j := range disposables {
		sum += disposables[j]
	}
	if len(disposables) != 0 {
		average = sum / len(disposables)
	}

	return average
}
func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
