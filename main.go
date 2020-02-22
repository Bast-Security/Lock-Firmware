package main

import (
	// Install with `go get github.com/eclipse/paho.mqtt.golang
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"time"
	"fmt"
)

func main() {
	defaultHandler := func(c mqtt.Client, m mqtt.Message) {
		fmt.Println(m)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://bastc.local:1883")
	opts.SetClientID("go thingy")
	opts.SetDefaultPublishHandler(defaultHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	handleQuit := func(c mqtt.Client, m mqtt.Message) {
		os.Exit(0)
	}

	if token := client.Subscribe("quit", 0, handleQuit); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for {
		fmt.Println("Publishing \"hello\" to \"hello\"")

		if token := client.Publish("hello", 0, false, "This is it"); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}

		//for topic /keyboard when it is used in a lock
		if token := client.Publish("/bast/csulb-bast/backdoor/keyboard", 0, false, "backdoor keyboard engaged"); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}

		//for top /card when it is used in a lcok
		if token := client.Publish("/bast/csulb-bast/backdoor/card", 0, false, "backdoor card engaged"); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}


		time.Sleep(time.Second)
	}
}