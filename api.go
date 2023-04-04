package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

func getEMS(user *User, channel string) (OpenEMSAPIResponse, error) {
	fmt.Println("Calling OpenEMS for seller " + user.Name + " at channel " + channel)
	client := &http.Client{}
	var url = user.Url.String + channel
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", user.Password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		var responseObject OpenEMSAPIResponse
		return responseObject, err
	} else {
		if resp.StatusCode != 200 {
			err = errors.New("Error: HTTP status code " + strconv.Itoa(resp.StatusCode))
			var responseObject OpenEMSAPIResponse
			return responseObject, err
		}
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var responseObject OpenEMSAPIResponse
	json.Unmarshal(bodyBytes, &responseObject)
	fmt.Printf("API Response as struct %+v\n", responseObject)
	return responseObject, err
}

func setEMS(user *User, channel string, value int) (OpenEMSAPIResponse, error) {
	fmt.Println("Calling OpenEMS for seller " + user.Name + " at channel " + channel)
	client := &http.Client{}
	var url = user.Url.String + channel
	var jsonStr = `{"value": ` + strconv.Itoa(value) + `}`
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", user.Password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		var responseObject OpenEMSAPIResponse
		return responseObject, err
	} else {
		if resp.StatusCode != 200 {
			err = errors.New("Error: HTTP status code " + strconv.Itoa(resp.StatusCode))
			var responseObject OpenEMSAPIResponse
			return responseObject, err
		}
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var responseObject OpenEMSAPIResponse
	json.Unmarshal(bodyBytes, &responseObject)
	fmt.Printf("API Response as struct %+v\n", responseObject)
	return responseObject, err
}

func getSM(user *User) (vzloggerAPIResponse, error) {
	fmt.Println("Calling Smart Meter for buyer " + user.Name)
	client := &http.Client{}
	var url = user.Url.String
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", user.Password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		var responseObject vzloggerAPIResponse
		return responseObject, err
	} else {
		if resp.StatusCode != 200 {
			err = errors.New("Error: HTTP status code " + strconv.Itoa(resp.StatusCode))
			var responseObject vzloggerAPIResponse
			return responseObject, err
		}
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var responseObject vzloggerAPIResponse
	json.Unmarshal(bodyBytes, &responseObject)
	fmt.Printf("API Response as struct %+v\n", responseObject)
	return responseObject, err
}

func getSMLocal() (vzloggerAPIResponse, error) {
	fmt.Println("Calling Smart Meter")
	client := &http.Client{}
	//var url = "http://localhost:8081/" normal
	var url = "http://localhost:8000/" //testing
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		var responseObject vzloggerAPIResponse
		return responseObject, err
	} else {
		if resp.StatusCode != 200 {
			err = errors.New("Error: HTTP status code " + strconv.Itoa(resp.StatusCode))
			var responseObject vzloggerAPIResponse
			return responseObject, err
		}
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	var responseObject vzloggerAPIResponse
	err = json.Unmarshal(bodyBytes, &responseObject)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("API Response as struct %+v\n", responseObject)
	return responseObject, err
}
