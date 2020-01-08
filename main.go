package main

import (
	"fmt"
	"github.com/anvlad11/dd-alicization/glagol"
	"os"
)

func main() {
	var err error

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
