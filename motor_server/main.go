package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

var cmd byte = 'x'
var lastModified time.Time

func afterModify(w http.ResponseWriter) {
	log.Printf("cmd is %q", cmd)
	w.Write([]byte{cmd})
	lastModified = time.Now()
}

func main() {
	http.HandleFunc("/w", func(w http.ResponseWriter, req *http.Request) {
		cmd = 'w'
		afterModify(w)
	})
	http.HandleFunc("/a", func(w http.ResponseWriter, req *http.Request) {
		cmd = 'a'
		afterModify(w)
	})
	http.HandleFunc("/s", func(w http.ResponseWriter, req *http.Request) {
		cmd = 's'
		afterModify(w)
	})
	http.HandleFunc("/d", func(w http.ResponseWriter, req *http.Request) {
		cmd = 'd'
		afterModify(w)
	})
	http.HandleFunc("/x", func(w http.ResponseWriter, req *http.Request) {
		cmd = 'x'
		afterModify(w)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte{cmd})
	})

	go func() {
		for true { // 1초 이상 변화가 없으면 정지 루프
			time.Sleep(time.Second)
			if time.Now().After(
				lastModified.Add(time.Second),
			) {
				if cmd != 'x' {
					log.Printf("no changes within a second. stop")
					cmd = 'x'
				}
			}
		}
	}()

	go serialWriter() // 시리얼 통신하는 루프

	log.Printf("server is Listening")
	http.ListenAndServe(":8080", nil)
}

func serialWriter() {
	tty := "/dev/ttyUSB0"
	if len(os.Args) == 2 {
		tty = os.Args[1]
	}
	options := serial.OpenOptions{
		PortName:        tty, // change this!
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	}
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer port.Close()

	recv := []byte{}

	_, err = port.Read(recv)
	if err != nil {
		log.Fatalf("serial.Read: %v", err)
	}
	log.Printf("recv: %q", recv)

	b := []byte{'x'}
	_, err = port.Write(b)
	if err != nil {
		log.Fatalf("serial.Write: %v", err)
	}

	for true {
		_, err = port.Write([]byte{cmd})
		if err != nil {
			log.Fatalf("serial.Write: %v", err)
		}
		time.Sleep(time.Millisecond * 30) // 30ms 쉼
	}
}
