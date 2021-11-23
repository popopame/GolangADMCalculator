package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/viper"
)

func main() {

	//Basic config file is set to be config.yaml
	//Reading Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	log.Printf("Used Config File %s", viper.ConfigFileUsed())

	//Mail Variables
	receivers := viper.GetStringSlice("receivers")
	senderMail := viper.GetString("sender")
	smtpAddr := viper.GetString("smtpAddr")
	smtpPort := viper.GetString("smtpPort")

	//Retrieving the Rapid APi Key  and SMTP Password first via flag, if nil, via Environement variables, and finally via the config.yaml file.
	var ApiKey string
	var smtpPasswd string
	var Mail string
	flag.StringVar(&ApiKey, "key", "", "Rapid Api Key, used for authentification")
	flag.StringVar(&smtpPasswd, "smtppasswd", "", "SMTP Password, Used to auth against smtp server")
	flag.StringVar(&Mail, "mail", "true", "Set if the script send dma result via mail, set to true by default")
	flag.Parse()

	//If the variables were not passed via flag, trying to retrieve them via environement variables
	if len(ApiKey) == 0 {
		ApiKey, set := os.LookupEnv("YAHOO_API_KEY")
		if !set {
			ApiKey := viper.GetString("apikey")
			if len(ApiKey) == 0 {
				log.Fatal("The Rapid API Key was not provided: exiting \n The Rapid API Key can be passed via flag of environement variable")
			}
		}
		fmt.Println(ApiKey)
	}

	//If the variables were not passed via flag, trying to retrieve them via environement variables
	if len(smtpPasswd) == 0 {
		smtpPasswd, set := os.LookupEnv("SMTP_PASSWD")
		if !set {
			smtpPasswd := viper.GetString("smtpPasswd")
			if len(smtpPasswd) == 0 {
				log.Fatal("The SMTP Passwd was not provided: exiting \n The SMTP Password can be passed via flag (smtppasswd) of environement variable (SMTP_PASSWD)")
			}
		}
		fmt.Println(smtpPasswd)
	}

	//Configuring slice  for the Momentum Computation
	SymbolsMomentumMap := make(map[string]float64)
	SymbolsList := []string{"SPY", "QQQ", "TLT"} //SPY QQQ and TLT are the assets used in the Dual Momentum Accelerated Approach
	//Populating the Maps of Symbols/Momentum, by calling the GetHistory func to retrieve a Symbols History and Computing Momemtum from it
	for _, symbol := range SymbolsList {
		symbolClosingHistory, _ := GetHistory(symbol, "1y", "1mo", ApiKey)

		computedMomentum, _ := ComputeMomentum(symbolClosingHistory)
		SymbolsMomentumMap[symbol] = computedMomentum
	}

	HighestMomentum := float64(0)

	//DMA Calculation
	if SymbolsMomentumMap["SPY"] > SymbolsMomentumMap["QQQ"] {
		if SymbolsMomentumMap["SPY"] > 0 {
			HighestMomentum = SymbolsMomentumMap["SPY"]
		} else if SymbolsMomentumMap["SPY"] < 0 {
			HighestMomentum = SymbolsMomentumMap["TLT"]
		}
	} else {
		if SymbolsMomentumMap["QQQ"] > 0 {
			HighestMomentum = SymbolsMomentumMap["QQQ"]
		} else if SymbolsMomentumMap["QQQ"] < 0 {
			HighestMomentum = SymbolsMomentumMap["TLT"]
		}
	}

	//Iterating over the Symbols/Momentum map and comparing every value to the Highest Momentum, when found setting the value and the key in a map
	//I don't really think that it is the fastest and easiest way to find and extract the value of the key linked to the highest momentum found in the previous loop, but it is the only one that cam to my mind, need help.
	//HighestSymbolMomentum := make(map[string]float64)
	var HighestSymbolMomentum string
	for key, _ := range SymbolsMomentumMap {
		if SymbolsMomentumMap[key] == HighestMomentum {
			HighestSymbolMomentum = key

			break
		}
	}

	//Switch to test if mail sending is activated
	if Mail == "true" {
		log.Printf("Mail Sending Activated (mail flag set to true)")
		//Generating Mail body and Title
		mailBody, mailTitle := GenerateMomentumMail(HighestSymbolMomentum, SymbolsMomentumMap)
		//Sending Mail
		SendMomentumResult(mailBody, mailTitle, receivers, senderMail, smtpPasswd, smtpAddr, smtpPort)
	} else if Mail == "false" {
		log.Printf("Mail Sending Desactivated -mail flag set to false")
		fmt.Printf("This month ADM winner is: %s\nThe momentum values are:%v", HighestSymbolMomentum, SymbolsMomentumMap)
	}

}

//GetHistory take a symbol(Ticker a a string), the range of the history (string), the interval(string) and ApiKey as arguments an will return the closing history for the symbol
func GetHistory(symbol string, daterange string, interval string, ApiKey string) ([]float64, error) {
	//url is the yahoo rapid api, where the symbol, daterange and interval is hard-coded in it
	url := fmt.Sprintf(
		"https://yfapi.net/v8/finance/spark?symbols=%s&range=%s&interval=%s",
		symbol,
		daterange,
		interval,
	)

	//Creation the request positioning the Api-Key and and executing the request
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("x-api-key", ApiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	//Initialisation of the responseStruct
	var response ResponseStruct

	//Unmarshaling the responseBody and outputing the result in the previously initializated struct
	err = json.Unmarshal(body, &response)

	//Initialisating a Variable use to store the slice that will be returned
	closingHistory := []float64{}

	//The order of the slice that is return in reversed, as the closing's price slice returned by the API is backward order: the most recent price first
	for i := len(response[symbol].Close) - 1; i != 0; i-- {
		closingHistory = append(closingHistory, response[symbol].Close[i])
	}
	//Returning the Symbols history
	return closingHistory, nil
}

//ComputeMomentum will compute the momentum of a given symbol, history on twelve month is passed via a float64 slice
func ComputeMomentum(history []float64) (float64, error) {

	//Calculating the moementum of 6 month
	SixMonthMomentum := ((history[1] - history[7]) / history[7]) * 100
	//Calculatin the momentum of 3 month
	ThreeMonthMomentum := ((history[1] - history[4]) / history[4]) * 100
	//Calculate the momentum of 1 month
	OneMonthMomentum := ((history[1] - history[2]) / history[2]) * 100
	//Calculation of the Compiled Momentum
	CompilatedMomentum := (SixMonthMomentum + ThreeMonthMomentum + OneMonthMomentum) / 3
	//Rounding the result
	CompilatedMomentum = math.Round(CompilatedMomentum*10000) / 10000
	return CompilatedMomentum, nil
}

//GenerateMomentumMail Take as imput the "winning" asset ticker, the full momentum map and return two string: one is the html encoded mail and the other is the MailTitle
func GenerateMomentumMail(ticker string, momentumMap map[string]float64) (string, string) {
	//Generating Month and year for the email
	time := time.Now()
	month := time.Month()
	year := time.Year()
	dmaMonthYear := fmt.Sprint(month) + "-" + fmt.Sprint(year)

	//Making the momentum a string an reworking the writing
	var mapString string
	for key, value := range momentumMap {
		mapString = mapString + fmt.Sprintf("%s:%v, ", key, value)
	}

	//Creating the template and populating it
	template, _ := template.ParseFiles("mail-template.html")

	var body bytes.Buffer

	template.Execute(&body, struct {
		DMAMonthYear string
		Ticker       string
		MomentumMap  string
	}{
		DMAMonthYear: dmaMonthYear,
		Ticker:       ticker,
		MomentumMap:  mapString,
	})

	mailbody := body.String()
	//Creating the title
	mailTitle := fmt.Sprintf("Subject: DMA Results %s", dmaMonthYear)

	return mailbody, mailTitle
}

//SendingMomentumResult take as input the mail body (in html format), mail title (string), a list of receivers(string), sender mail(string), senderpassword (string), smtp address (string) and smtp port (string)
func SendMomentumResult(mailBody string, mailTitle string, receivers []string, sender string, senderpassword string, smtpAddr string, smtpPort string) error {

	receiversheader := strings.Join(receivers, ",")
	log.Printf("Sending Mail via %s:%s", smtpAddr, smtpPort)
	header := make(map[string]string)
	header["From"] = sender
	header["To"] = receiversheader
	header["Content-Type"] = "text/html;"
	header["Subject"] = mailTitle
	header["MIME-Version"] = "1.0"

	var mailMessage string

	for key, value := range header {
		mailMessage += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	mailMessage += "\r\n" + mailBody
	//Creating the Auth Object for smtp
	auth := smtp.PlainAuth("", sender, senderpassword, smtpAddr)
	//sending Mail
	err := smtp.SendMail(smtpAddr+":"+smtpPort, auth, sender, receivers, []byte(mailMessage))
	if err != nil {
		fmt.Println(err)
		return err
	}
	//log.Println(mailMessage) Used to print Mail to debug
	log.Println("Mail sent")

	return nil
}

//ResponseStruct is compsoe of a map[string]SymbolStruct, as the name of the symbol change on each call, an "hard-coded" json struct is not possible
type ResponseStruct map[string]SymbolStruct

//SymbolStruct is defined on the RapidApi JSOn response
type SymbolStruct struct {
	ChartPreviousClose float64     `json:"chartPreviousClose"`
	Close              []float64   `json:"close"`
	DataGranularity    int64       `json:"dataGranularity"`
	End                interface{} `json:"end"`
	PreviousClose      interface{} `json:"previousClose"`
	Start              interface{} `json:"start"`
	Symbol             string      `json:"symbol"`
	Timestamp          []int64     `json:"timestamp"`
}
