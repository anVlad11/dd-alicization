package glagol

type DeviceList []*Device

type Device struct {
	ActivationCode     int64                `json:"activation_code"`
	ActivationRegion   string               `json:"activation_region"`
	Config             DeviceConfig         `json:"config"`
	Id                 string               `json:"id"`
	Name               string               `json:"name"`
	Platform           string               `json:"platform"`
	PromocodeActivated bool                 `json:"promocode_activated"`
	Glagol             DeviceGlagolSettings `json:"glagol"`
	Tags               []string             `json:"tags"`

	Discovery DeviceLocalDiscovery `json:"-"`
	Token     string               `json:"-"`
	LastState interface{}          `json:"-"`
}

type DeviceLocalDiscovery struct {
	Discovered   bool
	LocalAddress string
	LocalPort    string
}

func (deviceLocalDiscovery *DeviceLocalDiscovery) GetHost() string {
	return deviceLocalDiscovery.LocalAddress + ":" + deviceLocalDiscovery.LocalPort
}

type DeviceConfig struct {
	Name              string                        `json:"name"`
	ScreenSaverConfig DeviceConfigScreenSaverConfig `json:"screen_saver_config"`
}

type DeviceConfigScreenSaverConfig struct {
	Type string `json:"type"`
}

type DeviceGlagolSettings struct {
	Security DeviceGlagolSettingsSecurity `json:"security"`
}

type DeviceGlagolSettingsSecurity struct {
	ServerCertificate string `json:"server_certificate"`
	ServerPrivateKey  string `json:"server_private_key"`
}
