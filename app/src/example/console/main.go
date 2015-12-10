/**
 * The MIT License (MIT)
 *
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
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/webconnme/go-webconn"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

/*
#include <stdio.h>
#include <termio.h>

struct termios save;
int saved = 0;

void saveTerm(void) {
    saved = 1;
    tcgetattr(0,&save);
}

void restoreTerm(void) {
    if (saved == 1) {
        tcsetattr(0, TCSAFLUSH, &save);
    }
}

int getch(void) {
    char ch;
    struct termios buf;

    saveTerm();
    buf = save;
    buf.c_lflag &= ~(ICANON|ECHO);
    buf.c_cc[VMIN] = 1;
    buf.c_cc[VTIME] = 0;
    tcsetattr(0, TCSAFLUSH, &buf);

    ch = getchar();

    restoreTerm();
    return ch;
}
*/
import "C"

func HandleSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	raiseCount := 0

	for {
		// Block until a signal is received.
		select {
		case s := <-c:
			//fmt.Println("Signal Raised: ", s)
			switch s {
			case syscall.SIGINT:
				raiseCount++
				//fmt.Printf("Raise count: %d\n", raiseCount)
				if raiseCount >= 10 {
					C.restoreTerm()
					os.Exit(0)
				}
			case syscall.SIGKILL:
				fallthrough
			case syscall.SIGTERM:
				C.restoreTerm()
				os.Exit(0)
			}
		case <-time.NewTicker(100 * time.Millisecond).C:
			if raiseCount > 0 {
				SendByte(byte(3))
			}
			raiseCount = 0
			runtime.Gosched()
		}
	}
}

var context *zmq.Context
var sock *zmq.Socket

func OnReceive(s *zmq.Socket) error {
	buf, err := s.RecvBytes(0)
	if err != nil {
		return err
	}

	var messages []webconn.Message
	err = json.Unmarshal(buf, &messages)
	if err != nil {
		return err
	}

	for _, m := range messages {
		if m.Command == "rx" {
			fmt.Printf(string(m.Data))
		}
	}
	return nil
}

func SendByte(b byte) error {
	buf := []byte{b}

	var messages []webconn.Message
	messages = append(messages, webconn.Message{"tx", string(buf)})

	j, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	_, err = sock.SendBytes(j, 0)
	if err != nil {
		return err
	}

	return nil
}

func HandleNetwork() {
	var err error
	context, err = zmq.NewContext()
	if err != nil {
		log.Panic(err)
	}
	defer context.Term()

	sock, err = context.NewSocket(zmq.PAIR)
	if err != nil {
		log.Panic(err)
	}
	defer sock.Close()

	sock.Connect("tcp://nor.kr:3001")

	reactor := zmq.NewReactor()
	reactor.AddSocket(sock, zmq.POLLIN, func(state zmq.State) error { return OnReceive(sock) })

	err = reactor.Run(time.Second)

	if err != nil {
		log.Panic(err)
	}
}

func HandleKeyboard() {
	for {
		ch := C.getch()
		SendByte(byte(ch))
	}
}

func main() {

	done := make(chan bool)

	go HandleNetwork()
	go HandleSignal()
	go HandleKeyboard()

	<-done
}
