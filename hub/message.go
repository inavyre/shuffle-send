package hub

const MESSAGE_SEND = "send"
const MESSAGE_RECEIVE = "receive"
const MESSAGE_SHUFFLE = "shuffle"

const MSG_SET = "set_value";
const MSG_GROUP = "set_group";
const MSG_SHUFFLE = "shuffle"; 

// Message sent by the client
type Message struct {
    Kind string `json:"kind"`
    Value string `json:"value"`
    Group string `json:"group"`
} 

// Message to signal a new state
type StateMessage struct {
    Clients []ClientState
    Ready bool
}

// Message to signal a shuffling 
type ShuffleMessage struct {
    Value string
}
