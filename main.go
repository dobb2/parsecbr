package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	layoutISO        = "02/01/2006"
	customDateFormat = "02.01.2006"
	countLastDays    = 90
)

type customTime time.Time

type XMLValue float64

type ExchangeRate struct {
	XMLName    xml.Name   `xml:"ValCurs"`
	Date       customTime `xml:"Date,attr"`
	Name       string     `xml:"name,attr"`
	Сurrencies []Currency `xml:"Valute"`
}

type Currency struct {
	XMLName xml.Name `xml:"Valute"`
	Id      string   `xml:"ID,attr,omitempty"`
	//NumCode  int      `xml:"NumCode"`
	//CharCode string   `xml:"CharCode"`
	Nominal int      `xml:"Nominal"`
	Name    string   `xml:"Name"`
	Value   XMLValue `xml:"Value"`
}

type OtherCurr struct {
	Name  string
	Date  customTime
	Value float64
}

func (ud *XMLValue) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	ValueString := ""
	err := d.DecodeElement(&ValueString, &start)
	if err != nil {
		return err
	}
	ValueString = strings.Replace(ValueString, ",", ".", -1)
	res, err := strconv.ParseFloat(ValueString, 64)
	if err != nil {

		return err
	}
	*ud = XMLValue(res)
	return nil
}

func (ud *customTime) UnmarshalXMLAttr(attr xml.Attr) error {
	parse, err := time.Parse(customDateFormat, attr.Value)
	if err != nil {
		return err
	}
	*ud = customTime(parse)
	return nil
}

func (ud customTime) String() string {
	return time.Time(ud).Format(layoutISO)
}

func main() {
	now := time.Now()
	client := &http.Client{}
	expensiveCurrency := new(OtherCurr)
	cheapCurrency := new(OtherCurr)

	CurrencyValute := make(map[string]float64)

	for i := 0; i < countLastDays; i++ {
		now = now.Add(-24 * time.Hour)

		req, err := http.NewRequest(
			http.MethodGet,
			"https://www.cbr.ru/scripts/XML_daily_eng.asp?date_req="+now.Format(layoutISO),
			nil)
		req.Header.Add("User-Agent", "golang")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}
		//log.Println(resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		v := new(ExchangeRate)
		r := bytes.NewReader(body)
		d := xml.NewDecoder(r)
		d.CharsetReader = charset.NewReaderLabel
		err = d.Decode(&v)
		if err != nil {
			log.Println(err)
			continue
		}

		for i := range v.Сurrencies {
			Value := float64(v.Сurrencies[i].Value) / float64(v.Сurrencies[i].Nominal)
			if Value > expensiveCurrency.Value || expensiveCurrency.Name == "" {
				expensiveCurrency.Value = Value
				expensiveCurrency.Date = v.Date
				expensiveCurrency.Name = v.Сurrencies[i].Name
			}

			if Value < cheapCurrency.Value || cheapCurrency.Name == "" {
				cheapCurrency.Value = Value
				cheapCurrency.Date = v.Date
				cheapCurrency.Name = v.Сurrencies[i].Name
			}

			CurrencyValute[v.Сurrencies[i].Name] += (1 / Value)

		}

	}
	fmt.Println("Значение максимального курса валюты:")
	fmt.Println(
		"Название:", expensiveCurrency.Name,
		"Дата:", expensiveCurrency.Date,
		"Значение:", expensiveCurrency.Value)
	fmt.Println()
	fmt.Println("Значение минимального курса валюты:")
	fmt.Println(
		"Название:", cheapCurrency.Name,
		"Дата:", cheapCurrency.Date,
		"Значение:", cheapCurrency.Value)
	fmt.Println()
	fmt.Println("Среднее значение курса рубля по всем валютам за период в", countLastDays, "последних дней")
	for k, value := range CurrencyValute {
		fmt.Println("Валюта:", k, "Значение", value/float64(countLastDays))
	}

}
