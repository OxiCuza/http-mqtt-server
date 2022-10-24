package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"unicode/utf8"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/fhmq/hmq/broker"
	"github.com/fhmq/hmq/logger"
	"go.uber.org/zap"
)

var log = logger.Get()

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config, err := broker.ConfigureConfig(os.Args[1:])
	if err != nil {
		log.Error("configure broker config error: ", zap.Error(err))
	}

	b, err := broker.NewBroker(config)
	if err != nil {
		log.Fatal("New Broker error: ", zap.Error(err))
	}
	b.Start()
	print("[done mqtt]\n")

	http.HandleFunc("/sensor/temperature", func(writer http.ResponseWriter, request *http.Request) {
		res, err := io.ReadAll(request.Body)
		if err != nil {
		}
		fmt.Printf("host: %v method: %v uri: %v: data: %v", request.Host, request.Method, trimFirstRune(request.URL.Path), string(res))
		switch request.Method {
		case "GET":

			break
		case "POST":
			b.PublishMessage(&packets.PublishPacket{
				FixedHeader: packets.FixedHeader{
					MessageType:     3,
					Dup:             false,
					Qos:             2,
					Retain:          true,
					RemainingLength: 50,
				},
				TopicName: "sensor/temperature",
				MessageID: 1,
				Payload:   res,
			})
			fmt.Println("Success send HTTP to MQTT")
			fmt.Fprintf(writer, "Success !")
			break
		default:
			break
		}
	})

	go http.ListenAndServe(":8090", nil)
	print("[done http]\n")

	s := waitForSignal()
	fmt.Print("signal received, broker closed.", s)
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}
