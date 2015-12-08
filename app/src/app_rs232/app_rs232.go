package main

import (
	"fmt"
	"github.com/mikepb/go-serial"
	"log"
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
)

type jconfig struct {
	Command	string `json:"command"`
	Data	string `json:"data"`
}

var RS232options = serial.Options{

	BitRate	 	: 115200,
	DataBits 	: 8,
	Parity	 	: serial.PARITY_NONE,
	StopBits 	: 1,
	FlowControl : serial.FLOWCONTROL_NONE,
	Mode		: serial.MODE_READ_WRITE,
}

var RS232Path string
var serialPort *serial.Port
var url string

func RS232Open() {

	options := RS232options

	var err error
	if serialPort != nil {
		serialPort.Close()
	}

	serialPort, err = options.Open(RS232Path)
	if err != nil {
		log.Fatal("serial open ",err)
	} else {
		fmt.Println("serial open...")
	}
}

func RS232Close() {
	if serialPort != nil {
		serialPort.Close()
		fmt.Println("serial close...")
	}
}

func RS232Rx(ch chan<- []byte) {
	for {
		remain, err := serialPort.InputWaiting()
		if err != nil {
			fmt.Println(err)

		} else {
			if remain != 0 {
				buf := make([]byte, remain)
				len, err := serialPort.Read(buf)
				if err != nil {
					log.Fatal(err)
					panic(err)
				}
				ch <- buf
				fmt.Println("len : ", len)

			}

		}
	}
}

func RS232Tx(ch <-chan []byte) {

	for {
		buf := <-ch

		if len(buf) != 0 {
			fmt.Println(">>>tx msg : ", string(buf))
			_, err := serialPort.Write(buf)

			if err != nil {
				fmt.Println(">>>tx err : ",err)
				continue
			}
		}

	}
}

func httpget(ch chan<- []byte) {
	var jconf []jconfig

	for {
		client := &http.Client{}

		resp, err := client.Get(url)

		if err != nil {
			fmt.Println(err)

		} else {
			defer resp.Body.Close()
			contexts, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(">>>contexts : ", string(contexts), " for the url : ", url)
			json.Unmarshal(contexts, &jconf)

			for _, j:=range jconf {

				if j.Command == "tx" {
					fmt.Println(">>>recv data :", j.Data)
					ch <- []byte(j.Data)
				}
			}

		}

	}
}

func httpPost(ch <-chan []byte) {

	var jsconf jconfig
	jsconf.Command = "rx"

	for {
		jsconf.Data = string(<-ch)
		str, _ := json.Marshal(jsconf)

		buf := []byte("["+string(str)+"]")
		fmt.Println(">>>post : ",string(buf))

		client := &http.Client{}
		req, err := http.NewRequest("POST", url, strings.NewReader(string(buf)))
		resp, err := client.Do(req)
		if err != nil {

			fmt.Println(err)

		} else {
			defer resp.Body.Close()
			contexts, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(">>>context : ",string(contexts))
		}

	}
}

func main() {

	channel := make(chan bool)

	rxchan := make(chan []byte)
	txchan := make(chan []byte)

	RS232Path = "/dev/ttyS1"
	RS232Open()

	url = "Http://nor.kr:3000/v01/rs232/80"

	go RS232Rx(rxchan)
	go httpPost(rxchan)
	go httpget(txchan)
	go RS232Tx(txchan)

	<-channel

	RS232Close()

}
