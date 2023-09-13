/**
 * @Author: koulei
 * @Description:
 * @File: main
 * @Version: 1.0.0
 * @Date: 2023/9/4 20:24
 */

package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/flash520/pusher/pkg/connector"
	"github.com/flash520/pusher/pkg/pusher"
)

var hub *pusher.Hub

func main() {
	app := gin.Default()

	hub = pusher.NewHub()
	// hub.SetHandleRequest(func(data []byte, client pusher.Client) {
	// 	response := pusher.NewResponse("error", string(data))
	// 	client.SendMessage(response)
	// })
	app.GET("/ws/connect", Connect)

	config := connector.NewKafkaConfig("group1", "test-topic", "localhost:9092")
	reader := connector.NewKafkaReader(config)
	reader.SetChannel(hub.ReceiveChan())
	hub.SetReader(reader)

	go func() {
		time.Sleep(time.Second * 5)
		c, b := hub.GetReader(reader.Name())
		if b {
			c.Stop()
		}
	}()
	go test()

	_ = app.Run("")
}

var upgrader = websocket.Upgrader{}

func Connect(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	client := pusher.NewClient(hub, conn)
	client.User().SetUser("who am i")
	client.Run()
}

func test() {
	time.Sleep(time.Second * 5)
	logrus.Infof("test start")
	for i := 0; i < 1000000; i++ {
		hub.WriteEvent(pusher.NewData("http", i))
		logrus.Infof("current number: %d", i)
		time.Sleep(time.Millisecond * 20)
	}
}
