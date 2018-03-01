package flowserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/hexasoftware/flow"
	"github.com/hexasoftware/flow/registry"

	"github.com/gorilla/websocket"
)

//FlowSessionManager or FlowServerCore
type FlowSessionManager struct {
	name     string
	registry *registry.R
	store    string
	// List of flow sessions?
	sessions map[string]*FlowSession
	chats    map[string]*ChatRoom

	sync.Mutex
}

//NewFlowSessionManager creates a New initialized FlowSessionManager
func NewFlowSessionManager(r *registry.R, store string) *FlowSessionManager {
	return &FlowSessionManager{
		registry: r,
		store:    store,
		sessions: map[string]*FlowSession{},
	}
}

//CreateSession creates a new session
func (fsm *FlowSessionManager) CreateSession() *FlowSession {
	fsm.Lock()
	defer fsm.Unlock()
	for {
		ID := flow.RandString(10)
		sess, ok := fsm.sessions[ID]
		if !ok {
			sess = NewSession(fsm, ID)
			fsm.sessions[ID] = sess // XXX: Make this sync
			return sess
		}
	}
}

//LoadSession loads or creates a new Session
func (fsm *FlowSessionManager) LoadSession(ID string) (*FlowSession, error) {
	fsm.Lock()
	defer fsm.Unlock()
	if ID == "" {
		return nil, errors.New("Cannot be null")
	}
	sess, ok := fsm.sessions[ID]
	if !ok {
		sess = NewSession(fsm, ID)
		fsm.sessions[ID] = sess // Make this sync
	}
	return sess, nil

}

var upgrader = websocket.Upgrader{}

func (fsm *FlowSessionManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Contextual flowsession
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()

	// Room
	var sess *FlowSession
	defer func() {
		if sess == nil {
			return
		}
		// Remove client on exit
		sess.ClientRemove(c)
	}()

	/////////////////
	// Message handling
	// ////////

	// Websocket IO loop
	for {
		mt, data, err := c.ReadMessage()
		if err != nil {
			log.Println("Err:", err)
			break
		}
		if mt != websocket.TextMessage {
			log.Println("Not a text message?")
			break
		}

		m := RecvMessage{}
		err = json.Unmarshal(data, &m)
		if e(err) {
			log.Println("Err parsing message:", err)
			c.WriteJSON("bye")
			break
		}
		err = func() error {
			switch m.OP {
			/////////////////////////////
			// NEWSESSION request
			//////////////////
			case "sessionNew":
				log.Println("We want a new session so")
				sess = fsm.CreateSession()
				return sess.ClientAdd(c)
				//////////////////////////////////
				// LOADSESSION request
				////////////////////////////////////
			case "sessionLoad":
				sessID := string(m.ID)

				if sess != nil {
					sess.ClientRemove(c)
				}
				sess, err = fsm.LoadSession(sessID) // Set our session
				if e(err) {
					return err
				}
				return sess.ClientAdd(c)

			///////////////////////
			// DOCUMENTUPDATE Receive a document
			//////
			case "documentUpdate":
				sess, err = fsm.LoadSession(m.ID)
				if e(err) {
					return err
				}
				return sess.DocumentUpdate(c, m.Data)
			case "documentSave":
				if sess == nil {
					return errors.New("documentSave: invalid session")
				}
				return sess.DocumentSave(m.Data)

			//////////////////
			// NODE operations
			/////////
			case "nodeUpdate":
				if sess == nil {
					return errors.New("nodeUpdate: invalid session")
				}
				return sess.Broadcast(c, SendMessage{OP: m.OP, Data: m.Data})
			case "nodeProcess":
				if sess == nil {
					return errors.New("nodeRun: invalid session")
				}
				return sess.NodeProcess(c, m.Data)
			case "nodeTrain":
				if sess == nil {
					return errors.New("nodeTrain: invalid session")
				}
				return sess.NodeTrain(c, m.Data)
			////////////////////
			// CHAT operations
			/////////
			case "chatJoin":
				if sess == nil {
					return nil
				}
				var handle string
				json.Unmarshal(m.Data, &handle)
				log.Println("Joining with:", handle)
				sess.ChatJoin(c, handle)
			case "chatRename":
				if sess == nil {
					return nil
				}
				var handle string
				json.Unmarshal(m.Data, &handle)
				sess.Chat.ClientRename(c, handle)

			case "chatEvent":
				if sess == nil {
					return errors.New("invalid session")
				}
				sess.Chat.Event(c, m.Data)
			}
			return nil
		}()

	}
	log.Println("ws Is disconnecting", r.RemoteAddr)
}

const (
	storePath = "store"
)

func (fsm *FlowSessionManager) pathFor(ID string) (string, error) {
	{
		err := os.MkdirAll(storePath, 0755)
		if err != nil {
			return "", err
		}
		err = os.MkdirAll(filepath.Join(storePath, fsm.store), 0755)
		if err != nil {
			return "", err
		}
	}
	fpath := filepath.Clean(ID)
	_, fpath = filepath.Split(fpath)
	fpath = filepath.Join(storePath, fsm.store, fpath)
	return fpath, nil
}

func e(err error) bool {
	if err == nil {
		return false
	}

	_, file, line, _ := runtime.Caller(1)
	file = filepath.Base(file)

	log.Printf("Err: %s:%d %v", file, line, err)
	return true
}
