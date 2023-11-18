package hub

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	name         string
	value        string
    shuffleValue string
	ready        bool
	group        string
	isAlive      bool
	stateChan    chan StateMessage
	shuffleChan  chan ShuffleMessage
	conn         *websocket.Conn
	room         *Room
}

type ClientState struct {
	Name  string
    Group string
	Ready bool
}

func NewClient(name string, conn *websocket.Conn, room *Room) *Client {
	return &Client{
		name:         name,
		value:        "",
        shuffleValue: "",
		ready:        false,
        group:        "",
		isAlive:      true,
		stateChan:    make(chan StateMessage),
		shuffleChan:  make(chan ShuffleMessage),
		conn:         conn,
		room:         room,
	}
}

func (c *Client) Close() {
	c.room.hub.unregister <- c
	c.conn.Close()
}

func (c *Client) Write() {
	defer c.Close()

	for {
		select {
		case msg := <-c.stateChan:
			c.UpdateState(msg)
		case msg := <-c.shuffleChan:
			c.UpdateShuffleValue(msg)
		}
	}
}

func (c *Client) UpdateState(msg StateMessage) {
	var buf bytes.Buffer
	templ := template.Must(template.ParseFiles("templates/roomState.html"))
	templ.Execute(&buf, msg)
	s := buf.String()

	c.conn.WriteMessage(
		websocket.TextMessage,
		[]byte(s),
	)
}

func (c *Client) UpdateShuffleValue(msg ShuffleMessage) {
	c.shuffleValue = msg.Value

    var buf bytes.Buffer
    templ := template.Must(template.ParseFiles("templates/shuffleResult.html"))
    err := templ.Execute(&buf, msg)
    if (err != nil) {
        log.Printf("Error while parsing %s", err)
    }
    s := buf.String()

    c.conn.WriteMessage(
        websocket.TextMessage,
        []byte(s),
    )
}

func (c *Client) Read() {
	defer c.Close()

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)

		log.Printf("Received %s", msg.Kind)

		if err != nil {
			log.Printf("Error while reading: %s", err)
			break
		}

		switch msg.Kind {
		case MSG_SET:
			c.value = msg.Value
            c.group = msg.Group
            c.ready = true
            c.room.SendUpdate(c)
		case MSG_SHUFFLE:
            ok := c.room.ShuffleSend()
            if (!ok) {
                log.Print("Shuffling failed")
            }
		}
	}
}

func ConnectClient(
	wr http.ResponseWriter,
	req *http.Request,
	clientName string,
	room *Room,
) (*Client, error) {
	ws, err := upgrader.Upgrade(wr, req, nil)
	if err != nil {
		log.Printf("Failed connection: %s", err)
		return nil, errors.New("Failed connection")
	}

	client := NewClient(clientName, ws, room)
	room.RegisterClient(client)
	return client, nil
}
