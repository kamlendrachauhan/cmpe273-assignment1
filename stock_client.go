package main

import (
	"fmt"
	"strings"
	"strconv"
	"net/http"
        "io/ioutil"
        "bytes"
	"os"
)

type StockMap struct {
	stockVals map[string]string
	budget float64
}
func main() {
	var input string
	var budget string
	var tradeid string
	var optionFlag bool
	var stockString string
	var flag bool
	stocks := make(map[string]string)

	argumentCount := len(os.Args[1:])
		
	if argumentCount == 2{
		input = os.Args[1]
		budget = os.Args[2]
		optionFlag = true
		tempStrings := strings.Split(input,",")


        	for _, sValue := range tempStrings{
                	//subdividing the string
	                vals := strings.Split(sValue,":")

        	        stocks[vals[0]] = vals[1]
        	}

        	flag, stockString = validateinput(stocks)
		if !flag {
		  return
		}
	} else if argumentCount == 1 {
		tradeid = os.Args[1]
		optionFlag = false
	}

	
		
		url := "http://localhost:9999/stocks"
	    	
		var jsonStr []byte
		if optionFlag {
    			jsonStr = []byte(`{"method":"StockService.BuyStocks","params":[{"Budget":`+budget+`,"StockMap":{`+stockString+`}}],"id":"1"}`)
		} else {
			fmt.Println(tradeid)
			jsonStr = []byte(`{"method":"StockService.CheckPortfolio","params":[{"TradeId":"`+tradeid+`"}],"id":"1"}`)
		}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
 
    		req.Header.Set("Content-Type", "application/json")

    		client := &http.Client{}
	    	resp, err := client.Do(req)
    		if err != nil {
 		       panic(err)
    		}
    		defer resp.Body.Close()

    		fmt.Println("response Status:", resp.Status)
    		//fmt.Println("response Headers:", resp.Header)
    		body, _ := ioutil.ReadAll(resp.Body)
    		fmt.Println("response Body:", string(body))		
	
}

func validateinput(input map[string]string) (bool,string ){
	var sum int
	var StockString string
	for key,value := range input{
		intVal, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println(err)
		}
		sum += intVal
		if StockString == "" {
			StockString = "\""+key +"\":"+ value
		} else {
			StockString = StockString + ", \"" + key + "\":" + value
		}	

	}
	
	if sum != 100{
		return false,""
	}
	return true, StockString;
}
