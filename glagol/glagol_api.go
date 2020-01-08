package glagol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type APIClient struct {
	oauthToken string
}

const GLAGOL_API_BASE_URL = "https://quasar.yandex.net/glagol"

func NewAPIClient(oauthToken string) APIClient {
	glagol := APIClient{}
	glagol.oauthToken = oauthToken

	return glagol
}

func (api *APIClient) GetLocalDevices() (DeviceList, error) {
	var err error
	devices, err := api.GetDeviceList()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		fmt.Println("No devices found at account")
		return devices, nil
	}

	for _, device := range devices {
		token, err := api.GetJwtTokenForDevice(device)
		if err != nil {
			return nil, err
		}
		device.Token = token
	}

	devices, err = api.DiscoverDevices(devices)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		fmt.Println("No devices found in local network")
		return devices, nil
	}

	result := DeviceList{}
	for _, device := range devices {
		if device.Discovery.Discovered {
			result = append(result, device)
		}
	}
	fmt.Printf("%d devices found in local network\n", len(devices))

	return result, nil
}

func (api *APIClient) usePreconfiguredStation() bool {
	return os.Getenv("GLAGOL_USE_STATION_CONFIG") == "1"
}

func (api *APIClient) getPreconfiguredLocalDiscovery() DeviceLocalDiscovery {
	return DeviceLocalDiscovery{
		Discovered:   true,
		LocalAddress: os.Getenv("GLAGOL_STATION_ADDRESS"),
		LocalPort:    os.Getenv("GLAGOL_STATION_PORT"),
	}
}

func (api *APIClient) getPreconfiguredDeviceId() string {
	return os.Getenv("GLAGOL_STATION_ID")
}

func (api *APIClient) GetDeviceList() (DeviceList, error) {
	url := GLAGOL_API_BASE_URL + "/device_list"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Oauth "+api.oauthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response DeviceListSuccessfulResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	deviceList := DeviceList{}
	for _, device := range response.Devices {
		device.Discovery = DeviceLocalDiscovery{
			Discovered: false,
		}
		deviceList = append(deviceList, device)
	}

	return deviceList, nil
}

func (api *APIClient) DiscoverDevices(devices DeviceList) (DeviceList, error) {
	if api.usePreconfiguredStation() {
		preconfig := api.getPreconfiguredLocalDiscovery()
		for _, device := range devices {
			if device.Id == api.getPreconfiguredDeviceId() {
				device.Discovery = preconfig
			}
		}
		return devices, nil
	}

	return PopulateDevicesFromLocalDiscovery(devices)
}

func (api *APIClient) GetJwtTokenForDevice(device *Device) (string, error) {
	url := GLAGOL_API_BASE_URL + "/token?device_id=" + device.Id + "&platform=" + device.Platform
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Oauth "+api.oauthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response TokenSuccessfulResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}

type DeviceListSuccessfulResponse struct {
	Devices DeviceList `json:"devices"`
	Status  string     `json:"status"`
}

type TokenSuccessfulResponse struct {
	Token  string `json:"token"`
	Status string `json:"status"`
}
