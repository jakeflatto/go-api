package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func MakeWebRequest(url string) []byte {
	res, err := http.Get(url)
if err != nil {
	fmt.Println("Bad Req")
    panic(err.Error())
}

body, err := ioutil.ReadAll(res.Body)
if err != nil {
	fmt.Println("Bad Res")
    panic(err.Error())
}
return body
}

func GetTickerList() []string {
	var data []map[string]interface{}
	var tickers []string
	tickersURL := "http://www.sharadar.com/meta/tickers.json"
	bytes := MakeWebRequest(tickersURL)
	if err := json.Unmarshal(bytes, &data); err != nil {
        panic(err)
    }
	for _,v := range data {
		s := v["Ticker"].(string) 
		tickers = append(tickers,s)		
	}
	return tickers
}

func TickerValuesURL(ticker string) string {
	return fmt.Sprintf("https://www.quandl.com/api/v3/datasets/EOD/%v.json?api_key=-vmeVHiyChZMrx4zPdWE&start_date=2018-01-01&end_date=2018-01-05",ticker)
}

func CreateMapFromSlice(keys []string, values []float64) map[string]float64 {
	result := make(map[string]float64)
	for index,value := range values {
		result[keys[index + 1]] = value
	}
	return result
}

func GetValuesForTicker(ticker string) (string, map[string]map[string]float64) {
	var resData map[string]interface{}
	valuesMap := make(map[string]map[string]float64)
	bytes := MakeWebRequest(TickerValuesURL(ticker))
	if err := json.Unmarshal(bytes, &resData); err != nil {
		panic(err)
	}
	if resData["quandl_error"] != nil {
		return "", valuesMap 
	} else {
		var dataset map[string]interface{}
		var columnNames []string
		var values []float64
		var dailyValues [][]float64
		var dates []string
		dataset = resData["dataset"].(map[string]interface{})
		for k, v := range dataset { 
			switch t:= v.(type) {
				case []interface{}:
					for _,i:= range t {
						if k == "data" {
							switch s:= i.(type) {
								case []interface{}:
									values = values[len(values):]
									for index,j:= range s {
										if index > 0 {
											values = append(values, j.(float64))
										} else {
											dates = append(dates, j.(string))
										}
									}
									dailyValues = append(dailyValues, values)
							} 
						}
						if k == "column_names" {
							if column, ok := i.(string); ok {
								columnNames = append(columnNames, column)
							}
						}
					}
			}
		}
		for dateIndex, date := range dates {
			valuesMap[date] = CreateMapFromSlice(columnNames,dailyValues[dateIndex])
		}
		name := dataset["name"].(string)
		fmt.Println("Found one!",name)
		return name, valuesMap
	}
	
}

func main() {
	var successList []string
	tickerList := GetTickerList()
	numTickers := len(tickerList)
	fmt.Println(numTickers, "tickers found. Starting search on Quandl.")
	successCount := 0
	for i:=0;(i<numTickers && successCount <5) ;i++ {
		if tickerName, tickerValues:=GetValuesForTicker(tickerList[i]) ; tickerName!="" {
			successList = append(successList, tickerName)
			successCount++
			for i := range(tickerValues) {
				fmt.Printf("(%s) %v\n", i, tickerValues[i])
			}
		}
		if(i!=0 && i%100==0){
			fmt.Println(i," tickers checked.", (len(tickerList) - i), " tickers remaining.")
		}
	}
	fmt.Println(successList)
}