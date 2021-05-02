package routes

import (
	"encoding/json"
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/utils"
)

func webSocketConnect(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Connection event 1 - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketDisconnect(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketClose(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Close event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketError(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Error event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketMessage(ep *ikisocket.EventPayload) {
	fmt.Println("Web socket message")
	// fmt.Println(fmt.Sprintf("Message event - User: %s - Message: %s", ep.Kws.GetStringAttribute("user_id"), string(ep.Data)))
}

const setAddressesMessage = "setAddresses"

type TapperAddressesMessage struct {
	MessageType string `json:"messageType"`
	Addresses []string `json:"addresses"`
}

func sendWebsocketMessage(addresses []string) {
	webSocketConn, resp, err1 := websocket.DefaultDialer.Dial("ws://127.0.0.1:8080", nil)
	defer func() {
		_ = webSocketConn.Close()
	}()

	fmt.Printf("%v", resp)
	utils.CheckErr(err1)
	t := &TapperAddressesMessage{
		MessageType: setAddressesMessage,
		Addresses: addresses,
	}
	msgText, _ := json.Marshal(t)
	err := webSocketConn.WriteMessage(websocket.TextMessage, msgText)
	utils.CheckErr(err)
}

type ChangeAddressesBody struct {
	Addresses []string `json:"addresses"`
}
func WebSocketRoutes(app *fiber.App) {

	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		// kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true)
		// kws.Emit([]byte(fmt.Sprintf("Hello user with UUID: %s", kws.UUID)))
		kws.SetAttribute("user_id", kws.UUID)
	}))


	app.Post("/proxy/tapper", func(c *fiber.Ctx) error {
		// u := url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/"}
		addressesBody := new(ChangeAddressesBody)

		if err := c.BodyParser(addressesBody); err != nil {
			return err
		}

		go sendWebsocketMessage(addressesBody.Addresses)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"msg": "Success",
		})
	})

	ikisocket.On(ikisocket.EventMessage, webSocketMessage)
	ikisocket.On(ikisocket.EventConnect, webSocketConnect)
	ikisocket.On(ikisocket.EventDisconnect, webSocketDisconnect)
	ikisocket.On(ikisocket.EventClose, webSocketClose) // This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventError, webSocketError) // On error event
}
