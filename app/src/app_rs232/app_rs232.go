/**
 * The MIT License (MIT)
 *
 * Copyright (c) 2015 Jane Lee <jane@webconn.me>
 * Copyright (c) 2015 Edward Kim <edward@webconn.me>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
    "github.com/webconnme/go-webconn"
	"github.com/mikepb/go-serial"
)

import (
	"fmt"
	"log"
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

var client webconn.Webconn

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

func RS232Rx() {
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
                client.Write("rx", buf)
				fmt.Println("len : ", len)

			}

		}
	}
}

func RS232Tx(buf []byte) error {
    fmt.Println(">>>tx msg : ", string(buf))
    _, err := serialPort.Write(buf)

    if err != nil {
        return err
    }

    return nil
}

func main() {
	RS232Path = "/dev/ttyS1"
	RS232Open()
	defer RS232Close()

	url = "Http://nor.kr:3000/v01/rs232/80"

    client = webconn.NewClient("http://nor.kr:3000/v01/rs232/80")
    client.AddHandler("tx", RS232Tx)

	go RS232Rx()

    client.Run()
}
