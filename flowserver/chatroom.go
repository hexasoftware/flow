package flowserver

import (
	"errors"
	"log"
	"sync"

	"github.com/hexasoftware/flow/flowserver/flowmsg"

	"github.com/gorilla/websocket"
)

//ChatClient structure
type ChatClient struct {
	Handle string
	ws     *websocket.Conn
}

//ChatRoom will have an external ID
type ChatRoom struct {
	sync.Mutex
	clients []*ChatClient
	Events  []interface{} // Persistent chat, temporary datatype
}

// NewChatRoom create a chatRoom
func NewChatRoom() *ChatRoom {
	return &ChatRoom{sync.Mutex{}, []*ChatClient{}, []interface{}{}}
}

// ClientAdd add a client to session
func (r *ChatRoom) ClientAdd(c *websocket.Conn, handle string) error {
	r.Lock()
	defer r.Unlock()

	// Already exists
	for _, cl := range r.clients {
		if cl.ws == c {
			return nil
		}
	}

	r.clients = append(r.clients, &ChatClient{handle, c})

	r.broadcastClientList()

	for _, ev := range r.Events {
		if err := c.WriteJSON(flowmsg.SendMessage{OP: "chatEvent", Data: ev}); err != nil {
			return err
		}
	}
	return nil

}

// ClientRemove remove client from Session
func (r *ChatRoom) ClientRemove(c *websocket.Conn) {
	r.Lock()
	defer r.Unlock()

	for i, cl := range r.clients {
		if cl.ws == c {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			break
		}
	}

	r.broadcastClientList()
}

// ClientRename renames a client
func (r *ChatRoom) ClientRename(c *websocket.Conn, newHandle string) error {
	r.Lock()
	defer r.Unlock()

	var cli *ChatClient
	for _, cl := range r.clients {
		if cl.ws == c {
			cli = cl
			break
		}
	}
	if cli == nil {
		log.Println("Cli:Err")
		return errors.New("Client not found")
	}
	cli.Handle = newHandle

	return r.broadcastClientList()

}

// Event Message broadcast a message to clients in session
func (r *ChatRoom) Event(cl *websocket.Conn, v interface{}) error {
	r.Lock()
	defer r.Unlock()

	r.Events = append(r.Events, v)

	// Every one including self
	return r.broadcast(nil, flowmsg.SendMessage{OP: "chatEvent", Data: v})
}

func (r *ChatRoom) broadcast(c *websocket.Conn, v interface{}) error {
	for _, sc := range r.clients {
		if sc.ws == c { // ours
			continue
		}
		err := sc.ws.WriteJSON(v)
		if err != nil {
			return err
		}
	}
	return nil
}
func (r *ChatRoom) broadcastClientList() error {

	clients := make([]string, len(r.clients))
	for i, cl := range r.clients {
		clients[i] = cl.Handle
	}

	return r.broadcast(nil, flowmsg.SendMessage{OP: "chatUserList", Data: clients})
}
