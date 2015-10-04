package main
 
import (
  "github.com/gorilla/rpc"
  "github.com/gorilla/rpc/json"
  "net/http"
  "fmt"
  "io/ioutil"
  "os"
  "strconv"
  "strings"
  "math/rand"
)
 
type RPCArguments struct {	
	Budget float64
	StockMap
}

var holdingStockPriceMap map[string]float64
var holdingStockCountMap map[string]int
type StockMap map[string]int

var inputStockMap map[string]int

type StockService struct{}

type StockDetail struct{
        name string
        shareCount int
        value float64
}

func (s StockDetail) getString() string{
        ss := strconv.FormatFloat(s.value, 'f', 2, 64) 
        return s.name +":" +strconv.Itoa(s.shareCount) +":$" + ss
}

type ResponseVal struct {
	TradeId string
        Stocks string
        UninvestedAmount float64
	holdingStockPriceMap map[string]float64
	holdingStockCountMap map[string]int
}

func (r ResponseVal) getCurrentFolioStr(tradeID string, currentValMap map[string]float64) (string, float64) {
		var responsemap = ResponseMap[tradeID]
		var purchasedMapVal = responsemap.holdingStockPriceMap
		var purchasedCountMapVal = responsemap.holdingStockCountMap
		var returnString string
		var currentMarketValue float64
		for key, value := range currentValMap {
			var sign = ""
			if currentValMap[key] > holdingStockPriceMap[key]{
				sign = "+"	
			} else if currentValMap[key] < holdingStockPriceMap[key] {
				sign = "-"
			}
			if returnString == "" {
 			returnString = key +":"+ strconv.Itoa(purchasedCountMapVal[key])+":"+sign+"$"+strconv.FormatFloat(purchasedMapVal[key],'f',2,64)
			} else {		
			returnString = returnString + "," + key +":"+ strconv.Itoa(purchasedCountMapVal[key])+":"+sign+"$"+strconv.FormatFloat(purchasedMapVal[key],'f',2,64)
			}
			currentMarketValue = currentMarketValue + (value * float64(purchasedCountMapVal[key]))
		} 
	return returnString,currentMarketValue 	
}
type ResponsePortfolio struct {
	Stocks string
	CurrentMarketValue float64
	UninvestedAmount float64
}

type RequestArguments struct {
	TradeId string
}

var ResponseMap map[string]ResponseVal

func (service *StockService) BuyStocks(r *http.Request, arguments *RPCArguments, resp *ResponseVal) error {
	var stockString string
	var responseData string	
	inputStockMap = make(map[string]int)
	m := arguments.StockMap

	for key,value := range m {
		if stockString != ""{
			stockString = stockString +"+"+key
			} else {
			stockString = stockString + key
		}
		inputStockMap[key] = value
	}	

	responseData = fetchData(stockString)
	parseAndStructData(responseData, arguments.Budget, resp)
	return nil
}

func (service *StockService) CheckPortfolio(r *http.Request, arguments *RequestArguments, resp *ResponsePortfolio) error {
	tradeID := arguments.TradeId
	var responseMap = ResponseMap[tradeID]
	var priceMap = responseMap.holdingStockPriceMap
	var stockString string
	var fetchResponse string
	currentValMap := make(map[string]float64)
	for key,_ := range priceMap {
		 if stockString != ""{
                        stockString = stockString +"+"+key
                        } else {
                        stockString = stockString + key
                }
		
	}	
	fetchResponse = fetchData(stockString)
        data := strings.Split(strings.TrimSpace(fetchResponse),"\n")
	
	fmt.Println(" data : ",data)
 	for _,commaSepVal := range data {
                indiVal := strings.Split(commaSepVal, ",")

                //calculate how many shares to buy
                shareName := strings.Trim(indiVal[0],"\"")
		sharePrice, err := strconv.ParseFloat(strings.Trim(indiVal[1]," "),64)
                if err != nil{
                        fmt.Println(err)
                }
		fmt.Println("shareprice",sharePrice)
		currentValMap[shareName] = sharePrice
	}
	resp.Stocks, resp.CurrentMarketValue = responseMap.getCurrentFolioStr(tradeID, currentValMap)	
	resp.UninvestedAmount = responseMap.UninvestedAmount
	
	return nil
}
func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}

func main() {
	fmt.Println("Started service")
	
  	ResponseMap = make(map[string]ResponseVal)
	holdingStockPriceMap = make(map[string]float64)
	holdingStockCountMap = make(map[string]int)
  	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
  	s.RegisterService(new(StockService), "")
  	http.Handle("/stocks", s)
  	http.ListenAndServe(":9999", nil)
}

func parseAndStructData(responseData string, budget float64, resp *ResponseVal) ResponseVal{
      
	var responseStockPurchased string
        var remainingAmount float64
        data := strings.Split(strings.TrimSpace(responseData),"\n")
	remainingAmount = 0.00
	responseStockPurchased = ""
        for _,commaSepVal := range data {
                indiVal := strings.Split(commaSepVal, ",")
		var noOfShares int

		//calculate how many shares to buy
		shareName := strings.Trim(indiVal[0],"\"")
	
		percentageShare := inputStockMap[shareName]	                       
		budgetShare := budget*float64(percentageShare)/100

		sharePrice, err := strconv.ParseFloat(strings.Trim(indiVal[1]," "),64)
		if err != nil{
			fmt.Println(err)
		}
		
		//calculate no. of shares
		noOfShares = int(budgetShare/sharePrice)
		fmt.Println(noOfShares)
				
		remainingAmount = remainingAmount + (budgetShare - (float64(noOfShares)*sharePrice))			
	
		var stock = StockDetail{shareName,noOfShares,sharePrice}
	
		if responseStockPurchased == "" {
			responseStockPurchased = stock.getString()
		} else {
			responseStockPurchased = responseStockPurchased +", " + stock.getString()
		}
		holdingStockPriceMap[shareName] = sharePrice				
		holdingStockCountMap[shareName] = noOfShares
	 }
		//forming final response
		//Fetch random number as the tradeID
                var tradeId = rand.Int()
               
		resp.TradeId = strconv.Itoa(tradeId)
		resp.Stocks = responseStockPurchased
		resp.UninvestedAmount = remainingAmount
		//NOTE: value of folioString is same as stock string because as soon the request comes to check folio this needs to be updated at runtime
		var response = ResponseVal{strconv.Itoa(tradeId), responseStockPurchased, remainingAmount, holdingStockPriceMap,holdingStockCountMap}
		ResponseMap[resp.TradeId] = response
		
	return response
}

func fetchData(stockString string) string{
        url := "http://finance.yahoo.com/d/quotes.csv?s="+stockString+"&f=sa"
        fmt.Println("URL:>", url)

        req, err := http.NewRequest("GET", url, nil)

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
                panic(err)
        }
        defer resp.Body.Close()
        body, _ := ioutil.ReadAll(resp.Body)
        return string(body)
}


