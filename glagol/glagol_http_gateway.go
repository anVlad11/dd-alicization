package glagol

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"os"
	"time"
)

type HttpGateway struct {
	*gin.Engine
	conversation *Conversation
}

const GLAGOL_CTX_KEY = "glagol"

func StartWebServer(conversation *Conversation) error {
	server := NewHttpGateway(conversation)
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "OPTION", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	server.POST("/", SendCommand)
	server.GET("/", GetLastState)
	err := server.Start(os.Getenv("HTTP_HOST"))

	return err
}

func NewHttpGateway(conversation *Conversation) *HttpGateway {
	httpGateway := new(HttpGateway)
	httpGateway.Engine = gin.New()
	httpGateway.conversation = conversation

	logger := gin.LoggerWithWriter(gin.DefaultWriter, "/api/v1/status", "/metrics", "/healthz")

	httpGateway.Engine.Use(logger, gin.Recovery(), func(context *gin.Context) {
		context.Set(GLAGOL_CTX_KEY, httpGateway.conversation)
	})

	return httpGateway
}

func (httpGateway *HttpGateway) Start(addr string) error {
	return httpGateway.Engine.Run(addr)
}

func GetLastState(ctx *gin.Context) {
	glagolCtx, ok := ctx.Get(GLAGOL_CTX_KEY)
	if !ok {
		ctx.JSON(500, "Glagol conversation context not found")
		return
	}
	glagolCtxO := glagolCtx.(*Conversation)

	ctx.JSON(200, glagolCtxO.Device.LastState)
}

func SendCommand(ctx *gin.Context) {
	glagolCtx, ok := ctx.Get(GLAGOL_CTX_KEY)
	if !ok {
		ctx.JSON(500, "Glagol conversation context not found")
		return
	}
	glagolCtxO := glagolCtx.(*Conversation)

	var msg map[string]interface{}
	err := ctx.BindJSON(&msg)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	payload := DeviceRequestWrapper{
		ConversationToken: glagolCtxO.Device.Token,
		Id:                uuid.New().String(),
		SentTime:          time.Now().UnixNano(),
		Payload:           msg,
	}

	err = glagolCtxO.Connection.WriteJSON(payload)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	ctx.JSON(200, glagolCtxO.Device.LastState)
}
