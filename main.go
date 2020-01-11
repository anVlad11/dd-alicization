package main

import (
	"fmt"
	"github.com/anvlad11/dd-alicization/glagol"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	var err error

	if err := godotenv.Load(".env.local"); err != nil {
		log.Print("No .env file found")
	}
	//TODO: Получение токена по логину и паролю
	yandexApiToken := os.Getenv("YANDEX_OAUTH_TOKEN")

	glagolInstance := glagol.NewAPIClient(yandexApiToken)
	devices, err := glagolInstance.GetLocalDevices()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(devices)

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return
	}

	device := devices[0]

	conversation := glagol.NewConversation(device)
	conversation.Init()

	err = conversation.StartHttpGateway()

	if err != nil {
		fmt.Println(err)
	}
}
