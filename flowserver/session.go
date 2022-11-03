package flowserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hexasoftware/flow"
	"github.com/hexasoftware/flow/flowserver/flowbuilder"

	"github.com/gorilla/websocket"
)

// NodeActivity when nodes are processing
type NodeActivity struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // nodeStatus, Running, error, result
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Data      flow.Data `json:"data"`
	Error     string    `json:"error"`
}

// FlowSession Create a session and link clients
type FlowSession struct {
	sync.Mutex
	manager *FlowSessionManager

	ID string // Random handle for sessionID
	// List of clients on this session
	clients []*websocket.Conn
	Chat    ChatRoom

	RawDoc       []byte // Just share data
	nodeActivity map[string]*NodeActivity

	Data    map[interface{}]interface{}
	flow    *flow.Flow
	running bool
}

//NewSession creates and initializes a NewSession
func NewSession(fsm *FlowSessionManager, ID string) *FlowSession {
	// Or load
	//
	//
	fpath, err := fsm.pathFor(ID)
	if err != nil {
		log.Println("Error fetching filepath", err)
	}

	rawDoc, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Println("Warning: unable to read file:", err)
	}
	if rawDoc == nil {
		rawDoc = []byte{}
	}

	s := &FlowSession{
		Mutex:        sync.Mutex{},
		manager:      fsm,
		ID:           ID,
		clients:      []*websocket.Conn{},
		Chat:         ChatRoom{},
		RawDoc:       rawDoc,
		nodeActivity: map[string]*NodeActivity{},
		// Experimental
		Data: map[interface{}]interface{}{},

		flow: nil,
	}
	return s
}

// ClientAdd add a client to session
func (s *FlowSession) ClientAdd(c *websocket.Conn) error {
	s.Lock()
	defer s.Unlock()
	err := c.WriteJSON(SendMessage{OP: "sessionJoin", ID: s.ID})
	if err != nil {
		return err
	}
	desc, err := s.manager.registry.Descriptions()
	if err != nil {
		return err
	}
	err = c.WriteJSON(SendMessage{OP: "registry", Data: desc})
	if err != nil {
		return err
	}
	s.clients = append(s.clients, c)

	if len(s.RawDoc) == 0 {
		return nil
	}
	err = c.WriteJSON(SendMessage{OP: "document", Data: json.RawMessage(s.RawDoc)})
	if err != nil {
		return err
	}

	// Sending activity
	return c.WriteJSON(s.activity())

	// Send registry
}

// ClientRemove remove client from Session
func (s *FlowSession) ClientRemove(c *websocket.Conn) {
	s.Lock()
	defer s.Unlock()
	for i, cl := range s.clients {
		if cl == c {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
	s.Chat.ClientRemove(c)
	/*if len(s.clients) == 0 && s.flow == nil {
		log.Println("No more clients, remove session")
		delete(s.mgr.sessions, s.ID) // Clear memory session
	}*/
}

// ChatJoin the chat room on this session
func (s *FlowSession) ChatJoin(c *websocket.Conn, handle string) {
	s.Chat.ClientAdd(c, handle)
}

// DocumentUpdate client c Updates the session document
func (s *FlowSession) DocumentUpdate(c *websocket.Conn, data []byte) error {
	s.Lock()
	defer s.Unlock()

	s.RawDoc = make([]byte, len(data))
	copy(s.RawDoc, data)

	return s.broadcast(c, SendMessage{OP: "document", Data: json.RawMessage(s.RawDoc)})
}

// DocumentSave persist document in a file
func (s *FlowSession) DocumentSave(data []byte) error {
	s.Lock()
	defer s.Unlock()

	s.RawDoc = make([]byte, len(data))
	copy(s.RawDoc, data)

	fpath, err := s.manager.pathFor(s.ID)
	if err != nil {
		log.Println("path error", err)
		return err
	}

	err = ioutil.WriteFile(fpath, s.RawDoc, os.FileMode(0600))
	if err != nil {
		log.Println("writing file", err)
		return err
	}

	s.notify("Session saved")
	return s.broadcast(nil, SendMessage{OP: "documentSave", Data: "saved"})
}

// Document send document to client c
func (s *FlowSession) Document(c *websocket.Conn) error {
	s.Lock()
	defer s.Unlock()

	return c.WriteJSON(SendMessage{OP: "document", Data: json.RawMessage(s.RawDoc)})
}

// NodeProcess a node triggering results
// Build a flow and run
func (s *FlowSession) NodeProcess(c *websocket.Conn, data []byte) error {

	if s.flow != nil {
		s.Notify("flow is already running")
		return errors.New("nodes already running")
	}

	ids := []string{}
	err := json.Unmarshal(data, &ids)
	if err != nil {
		return err
	}

	// *New* 25-02-2018 node Array
	//ID := string(data[1 : len(data)-1]) // remove " instead of unmarshalling json

	// Clear activity
	s.nodeActivity = map[string]*NodeActivity{}
	s.Broadcast(nil, s.activity()) // Empty activity in clients

	build := func() error {

		localR := s.manager.registry.Clone()
		//Add our log func that is not in global registry
		localR.Add("Notify", func(v flow.Data, msg string) flow.Data {
			log.Println("Notify:", msg)
			s.Notify(msg)
			return v
		})
		localR.Add("Log", func() io.Writer {
			return s
		})
		// this will be disabled
		localR.Add("Output", func(d interface{}) interface{} {
			//r := fmt.Sprint("Result:", d)
			// Do something
			//s.Notify(r)
			//s.Write([]byte(r))
			return d
		})

		builder := flowbuilder.New(localR)
		builder.Load(s.RawDoc)

		ops := make([]flow.Operation, len(ids))

		for i, id := range ids {
			ops[i] = builder.Build(id)
		}
		// Multiple ops
		if builder.Err != nil {
			return builder.Err
		}

		s.flow = builder.Flow()
		log.Println("Flow:", s.flow)

		log.Println("Experimental: Loading data")
		for k, v := range s.Data {
			s.flow.Data.Store(k, v)
		}

		defer func() { // After routing gone
			s.flow = nil
		}()

		// Flow hooks
		// Flow activity TODO: needs improvements as it shouldn't send the overall activity to client
		// instead should send singular events
		s.flow.Hook(flow.Hook{
			Any: func(name string, hookOp flow.Operation, triggerTime time.Time, extra ...flow.Data) {
				s.Lock()
				defer s.Unlock()

				nodeIDs := builder.GetOpIDs(hookOp)
				updated := true
				for _, nodeID := range nodeIDs {
					act, ok := s.nodeActivity[nodeID]
					if !ok {
						act = &NodeActivity{ID: nodeID}
						s.nodeActivity[nodeID] = act
					}
					status := ""
					switch name {
					case "Wait":
						status = "waiting"
						act.StartTime = time.Time{}
						act.EndTime = time.Time{}
					case "Start":
						status = "running"
						act.EndTime = time.Time{}
						act.StartTime = triggerTime
					case "Finish":
						status = "finish"
						act.EndTime = triggerTime
						// only load data from requested node
						// Or if node has the data retrieval flag
						// if running ids contains the nodeID
						// we add the data
						for _, id := range ids {
							if nodeID == id {
								act.Data = extra[0]
							}
						}
						//node := builder.Doc.FetchNodeByID(ID)
						//log.Println("Should we add data:", ID, node)
						//if node.Prop["data"] == "true" || nodeID == ID {
						//	act.Data = extra[0]
						//}
					case "Error":
						status = "error"
						act.EndTime = triggerTime
						act.Error = fmt.Sprint(extra[0])
					}
					if act.Status == status {
						continue
					}
					updated = true
					act.Status = status
				}
				if !updated {
					return
				}
				s.broadcast(nil, s.activity())

			},
		})

		/*op, ok := builder.OperationMap[ID]
		if !ok {
			return fmt.Errorf("Operation not found %v", ID)
		}*/
		log.Println("Processing operation")
		sess := s.flow.NewSession()
		_, err = sess.Run(ops...)
		if err != nil {
			log.Println("Error operation", err)
			return err
		}

		log.Println("Experimental storing data to session")
		// Copy Data from flow
		s.flow.Data.Range(func(k, v interface{}) bool {
			s.Data[k] = v
			return true
		})
		log.Println("Operation finish")
		log.Println("Flow now:", s.flow)
		return nil
	}

	// Background running
	go func() {
		err := build()
		if err != nil {
			s.Notify(fmt.Sprint("ERR:", err))
		}
	}()

	return nil
}

// NodeTrain temporary operation for repeating a node
// this is for the demo purposes
func (s *FlowSession) NodeTrain(c *websocket.Conn, data []byte) error {
	ID := string(data[1 : len(data)-1]) // remove " instead of unmarshalling json
	if s.flow != nil {
		s.Notify("a node is already running")
		return errors.New("node already running")
	}

	// Clear activity
	s.nodeActivity = map[string]*NodeActivity{}
	s.Broadcast(nil, s.activity()) // Ampty activity

	build := func() error {
		localR := s.manager.registry.Clone()
		//Add our log func that is not in global registry
		localR.Add("Notify", func(v flow.Data, msg string) flow.Data {
			log.Println("Notify:", msg)
			s.Notify(msg)
			return v
		})
		localR.Add("Log", func() io.Writer {
			return s
		})
		localR.Add("Output", func(d interface{}) interface{} {
			//r := fmt.Sprint("Result:", d)
			// Do something
			//s.Notify(r)
			//s.Write([]byte(r))
			return d
		})
		builder := flowbuilder.New(localR)
		builder.Load(s.RawDoc).Build(ID)
		if builder.Err != nil {
			return builder.Err
		}

		s.flow = builder.Flow()
		log.Println("Flow:", s.flow)

		// XXX: Possibly remove
		log.Println("Experimental: Loading global data")
		for k, v := range s.Data {
			s.flow.Data.Store(k, v)
		}

		defer func() { // After routing gone
			s.flow = nil
		}()
		// Flow activity

		op, ok := builder.OperationMap[ID]
		if !ok {
			return fmt.Errorf("Operation not found %v", ID)
		}
		log.Println("Processing operation")

		epochs := 5000
		s.Notify(fmt.Sprintf("Training for %d epochs", epochs))
		for i := 0; i < epochs; i++ {
			res, err := op.Process()
			if err != nil {
				log.Println("Error operation", err)
				return err
			}
			if i%1000 == 0 {
				fmt.Fprintf(s, "Res: %v", res)
				fmt.Fprintf(s, "Training... %d/%d", i, epochs)
				outs := builder.Doc.FetchNodeBySrc("Output")
				if len(outs) == 0 {
					continue
				}
			}
		}
		fmt.Fprintf(s, "%v", s.flow)
		// Copy Data from flow
		s.flow.Data.Range(func(k, v interface{}) bool {
			s.Data[k] = v
			return true
		})
		log.Println("Operation finish")
		log.Println("Flow now:", s.flow)
		return nil
	}

	// Parallel building
	go func() {
		err := build()
		if err != nil {
			s.Notify(fmt.Sprint("ERR:", err))
		}
	}()

	return nil

}

func (s *FlowSession) activity() *SendMessage {

	msg := SendMessage{OP: "nodeActivity",
		Data: map[string]interface{}{
			"serverTime": time.Now(),
			"nodes":      s.nodeActivity,
		},
	}

	return &msg
}

// Notify broadcast a notification to clients
func (s *FlowSession) Notify(v interface{}) error {
	s.Lock()
	defer s.Unlock()
	return s.notify(v)
}

func (s *FlowSession) notify(v interface{}) error {
	return s.broadcast(nil, SendMessage{OP: "sessionNotify", Data: v})
}

// Write io.Writer implementation to send event logging
func (s *FlowSession) Write(data []byte) (int, error) {
	err := s.Broadcast(nil, SendMessage{OP: "sessionLog", Data: string(data)})
	if err != nil {
		return -1, err
	}
	return len(data), nil
}

// Broadcast broadcast a message in session besides C
func (s *FlowSession) Broadcast(c *websocket.Conn, v interface{}) error {
	s.Lock()
	defer s.Unlock()

	return s.broadcast(c, v)
}

//
func (s *FlowSession) broadcast(c *websocket.Conn, v interface{}) error {
	for _, sc := range s.clients {
		if sc == c { // ours
			continue
		}
		err := sc.WriteJSON(v)
		if err != nil {
			return err
		}
	}
	return nil

}
