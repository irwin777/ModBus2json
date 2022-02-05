package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/goburrow/modbus"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

var (
	mx       sync.Mutex
	register map[string]uint16
	counter  map[string]float32
)

func main() {
	register = make(map[string]uint16)
	counter = make(map[string]float32)
	register["Voltage"] = 0
	register["Current"] = 8
	register["ActivePower"] = 18
	register["ReactivePower"] = 26
	register["PowerFactor"] = 42
	register["Frequency"] = 54
	register["TotalActivePower"] = 256
	register["TotalReactivePower"] = 1024
	opros()
	CronTicker := time.NewTicker(time.Second * 20)
	go func() {
		for range CronTicker.C {
			opros()
		}
	}()
	http.HandleFunc("/", handler)
	fmt.Printf("Starting server at port 8080\n")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	mx.Lock()
	defer mx.Unlock()
	jsonBody, err := json.Marshal(counter)
	if err != nil {
		fmt.Fprintf(w, "Error: %v.\n", err)
		return
	}
	w.Header().Set("accept", "application/json")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonBody)
}

func bin2float(b []byte) float32 {
	bb := binary.BigEndian.Uint32(b)
	f := math.Float32frombits(bb)
	return f
}

func opros() {
	mx.Lock()
	defer mx.Unlock()
		rtu := modbus.NewRTUClientHandler("/dev/ttyUSB0")
		rtu.BaudRate = 9600
		rtu.DataBits = 8
		rtu.Parity = "E"
		rtu.StopBits = 1
		rtu.SlaveId = 3
		rtu.Timeout = 1 * time.Second
		err := rtu.Connect()
		if err != nil {
			log.Println(err)
			for key, _ := range register {
				counter[key] = 0.0
			}
			return
		}
		defer rtu.Close()
		client := modbus.NewClient(rtu)


	for key, val := range register {
		req, err := client.ReadInputRegisters(val, 2)
		if err != nil {
			log.Println("client.ReadFloat32s", err)
			counter[key] = 0.0
			continue
		}
		counter[key] = bin2float(req)
	}
}
