package glagol

import (
	"context"
	"errors"
	"github.com/grandcat/zeroconf"
	"log"
	"strconv"
	"strings"
	"time"
)

/**
Собрано прямо из примера модуля grandcat/zeroconf
https://github.com/grandcat/zeroconf/README.md
*/
func PopulateDevicesFromLocalDiscovery(devices DeviceList) (DeviceList, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return devices, errors.New("Failed to initialize resolver: " + err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			entryMap := map[string]interface{}{}
			for _, s := range entry.Text {
				entryData := strings.Split(s, "=")
				if len(entryData) == 2 {
					entryMap[entryData[0]] = entryData[1]
				}
			}
			/**
			 Т.к. у меня только Я.Станция, то ограничил доступные платформы.
			 Слышал, что Я.Модуль так же работает, но мне не на чем проверить.
			upd: я проверил, работает :)
			**/
			if platform, ok := entryMap["platform"]; ok && platform == "yandexstation" || platform == "yandexmodule"{
				for _, device := range devices {
					if device.Id == entryMap["deviceId"] {
						device.Discovery = DeviceLocalDiscovery{
							Discovered:   true,
							LocalAddress: entry.AddrIPv4[0].String(), //Прикрутить бы IPv6
							LocalPort:    strconv.FormatInt(int64(entry.Port), 10),
						}
					}
				}
			}
		}
		log.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	//_yandexio._tcp это устройства Умного дома Яндекса
	err = resolver.Browse(ctx, "_yandexio._tcp", "local.", entries)
	if err != nil {
		return devices, errors.New("Failed to browse: " + err.Error())
	}

	<-ctx.Done()

	return devices, nil
}
