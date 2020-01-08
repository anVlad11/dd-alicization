package glagol

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type Conversation struct {
	Device     *Device
	Connection *websocket.Conn
}

func NewConversation(device *Device) Conversation {
	conversation := Conversation{Device: device}

	return conversation
}

func (conversation *Conversation) Init() {
	go func() {
		err := conversation.runWsConnection()
		if err != nil {
			panic(err)
		}
	}()
}

func (conversation *Conversation) runWsConnection() error {
	var err error
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: conversation.Device.Discovery.GetHost(), Path: "/"}

	dialer := websocket.DefaultDialer
	rootCAs, err := conversation.rootCAs()
	if err != nil {
		return err
	}
	dialer.TLSClientConfig = &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: true,
	}

	log.Printf("connecting to %s", u.String())
	conversation.Connection, _, err = dialer.Dial(u.String(), http.Header{"Origin": {"http://yandex.ru/"}})
	if err != nil {
		return errors.New("dial: " + err.Error())
	}
	defer conversation.Connection.Close()

	done := make(chan struct{})

	var locked = false

	go func() {
		defer close(done)
		for {
			_, message, err := conversation.Connection.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			go func() {
				for locked {

				}
				locked = true
				var latestState map[string]interface{}
				_ = json.Unmarshal(message, &latestState)
				conversation.Device.LastState = latestState
				locked = false
			}()
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	guid := uuid.NewV4()
	pingMessage := DeviceRequestWrapper{
		ConversationToken: conversation.Device.Token,
		Id:                guid.String(),
		SentTime:          time.Now().UnixNano(),
		Payload: map[string]interface{}{
			"command": "ping",
		},
	}
	err = conversation.Connection.WriteJSON(pingMessage)
	if err != nil {
		return errors.New("write: " + err.Error())

	}
	if os.Getenv("GLAGOL_CONFIRM_CONNECTION") == "1" {
		connectionConfirmationRequest := DeviceRequestWrapper{
			ConversationToken: conversation.Device.Token,
			Id:                guid.String(),
			SentTime:          time.Now().UnixNano(),
			Payload: map[string]interface{}{
				"command": "sendText",
				"text":    "Повтори за мной 'Локальный сервер подключен к станции'",
			},
		}
		err = conversation.Connection.WriteJSON(connectionConfirmationRequest)
		if err != nil {
			return errors.New("write: " + err.Error())
		}
	}
	log.Printf("connected to %s", u.String())

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			//TODO: добавить сюда пинг (а надо?)

		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := conversation.Connection.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				return errors.New("write close:" + err.Error())
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

func (conversation *Conversation) rootCAs() (*x509.CertPool, error) {
	certs := x509.NewCertPool()
	block, _ := pem.Decode([]byte(conversation.Device.Glagol.Security.ServerCertificate))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse certificate: " + err.Error())
	}
	certs.AddCert(cert)
	return certs, nil
}

func (conversation *Conversation) StartHttpGateway() error {
	return StartWebServer(conversation)
}

type DeviceStatusResponse struct {
	Extra          interface{}   `json:"extra"`
	Id             string        `json:"id"`
	SentTime       int64         `json:"sentTime"`
	State          ResponseState `json:"state"`
	Status         string        `json:"status"`
	ProcessingTime int64         `json:"processingTime"`
}

type ResponseState struct {
	AliceState                 string              `json:"aliceState"`
	CanStop                    bool                `json:"canStop"`
	PlayerState                ResponsePlayerState `json:"playerState"`
	Playing                    bool                `json:"playing"`
	TimeSinceLastVoiceActivity int64               `json:"timeSinceLastVoiceActivity"`
	Volume                     float64             `json:"volume"`
}

type ResponsePlayerState struct {
	Duration       float64     `json:"duration"`
	Extra          interface{} `json:"extra"`
	HasNext        bool        `json:"hasNext"`
	HasPause       bool        `json:"hasPause"`
	HasPlay        bool        `json:"hasPlay"`
	HasPrev        bool        `json:"hasPrev"`
	HasProgressBar bool        `json:"hasProgressBar"`
	LiveStreamText string      `json:"liveStreamText"`
	Progress       float64     `json:"progress"`
	Subtitle       string      `json:"subtitle"`
	Title          string      `json:"title"`
}

type DeviceRequestWrapper struct {
	ConversationToken string      `json:"conversationToken"`
	Id                string      `json:"id"`
	SentTime          int64       `json:"sentTime"`
	Payload           interface{} `json:"payload"`
}
