package hub

import (
	"log"
	"math/rand"
	"sort"
)

type Hub struct {
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) CreateRoom(token string) *Room {
	if h.rooms[token] == nil {
		h.rooms[token] = NewRoom(token, h)
	}
	return h.rooms[token]
}

func (h *Hub) GetRoom(token string) *Room {
	return h.rooms[token]
}

func (h *Hub) HasRoom(token string) bool {
	return (h.rooms[token] != nil)
}

func (h *Hub) CloseRoom(token string) {
	delete(h.rooms, token)
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			client.room.RegisterClient(client)
		case client := <-h.unregister:
			client.room.UnregisterClient(client)
		}
	}
}

type Room struct {
	hub     *Hub
	token   string
	clients map[*Client]bool
}

func NewRoom(token string, hub *Hub) *Room {
	return &Room{
		hub:     hub,
		token:   token,
		clients: make(map[*Client]bool),
	}
}

func (r *Room) GetState() StateMessage {
	var clients []ClientState
	var ready bool = true

	for client := range r.clients {
		if !client.ready {
			ready = false
		}
		clients = append(clients, ClientState{
			client.name,
			client.group,
			client.ready,
		})
	}

    if len(clients) == 0 {
        ready = false
    }

	return StateMessage{
		Clients: clients,
		Ready:   ready,
	}
}

func RemoveIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func (r *Room) ShuffleSend() bool {
    log.Printf("[%s] Shuffling!", r.token)

    // Count number of clients in each group
    groupsCount := make(map[string]int)
    for client := range r.clients {
        _, ok := groupsCount[client.group]
        if !ok { groupsCount[client.group] = 0 }
        groupsCount[client.group] += 1
    }
   
    // Define order
    groupsOrder := make([]string, 0, len(groupsCount))
    for g := range groupsCount {
        groupsOrder = append(groupsOrder, g)
    }
    sort.SliceStable(groupsOrder, func(i, j int) bool {
        if groupsOrder[i] == "" { return false }
        return groupsCount[groupsOrder[i]] > groupsCount[groupsOrder[j]]
    })

    // Create array of clients where more numerous groups are first
    targets := make([]*Client, 0, len(r.clients))
    for _, group := range groupsOrder {
        for client := range r.clients {
            if client.group == group {
                targets = append(targets, client)
            }
        }
    }

	// Create shuffle array and map of values
	clients := []*Client{}
	clientsMap := make(map[*Client]string)
	for client := range r.clients {
		clients = append(clients, client)
		clientsMap[client] = ""
	}
	// Shuffle array
	rand.Shuffle(len(clients), func(i, j int) { clients[i], clients[j] = clients[j], clients[i] })

	// Try to create the map
	for _, target := range targets {
		value := ""
		for index, source := range clients {
			if target != source && (target.group == "" || target.group != source.group) {
				value = source.value
				clients = RemoveIndex[*Client](clients, index)
				break
			}
		}

		if value == "" {
			return false
		}
		clientsMap[target] = value
	}

	// Send
	for client, value := range clientsMap {
        log.Printf("Sending value %s", value)
		client.shuffleChan <- ShuffleMessage{Value: value}
	}

	return true
}

func (r *Room) SendUpdate(c *Client) {
	state := r.GetState()
	c.UpdateState(state)
	for client := range r.clients {
		if client != c {
			client.stateChan <- state
		}
	}
}

func (r *Room) RegisterClient(client *Client) {
	r.clients[client] = true
	client.room.SendUpdate(client)
}

func (r *Room) UnregisterClient(client *Client) {
    log.Printf("[%s] Unregistering client", r.token)
	close(client.stateChan)
	close(client.shuffleChan)
	delete(r.clients, client)
	client.room.SendUpdate(client)

	if len(r.clients) == 0 {
        log.Printf("[%s] Closing", r.token)
		r.hub.CloseRoom(r.token)
	}
}
