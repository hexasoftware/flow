package flowbuilder

import (
	"encoding/json"
	"errors"
	"flow"
	"flow/registry"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

// ErrLoop loop error
var ErrLoop = errors.New("Looping is disabled for now")

// FlowBuilder builds a flow from flow-ui json data
type FlowBuilder struct {
	registry     *registry.R
	Doc          *FlowDocument
	flow         *flow.Flow
	OperationMap map[string]flow.Operation
	nodeTrack    map[string]bool
	Err          error
}

// New creates a New builder
func New(r *registry.R) *FlowBuilder {
	return &FlowBuilder{
		registry:     r,
		OperationMap: map[string]flow.Operation{},
		nodeTrack:    map[string]bool{},
	}
}

// GetOpIDs fetches operation IDs
// with portals we can have several ids pointing to same operation
func (fb *FlowBuilder) GetOpIDs(op flow.Operation) []string {
	var ret []string

	for k, v := range fb.OperationMap {
		if op == v {
			ret = append(ret, k)
		}
	}
	return ret
}

// Load document from json into builder
func (fb *FlowBuilder) Load(rawData []byte) *FlowBuilder {
	fb.flow = flow.New()
	fb.flow.UseRegistry(fb.registry)

	doc := &FlowDocument{[]Node{}, []Link{}, []Trigger{}}
	log.Println("Loading document from:", string(rawData))
	err := json.Unmarshal(rawData, doc)
	if err != nil {
		fb.Err = err
		return fb
	}

	fb.Doc = doc

	return fb
}

// Build a flow starting from node
func (fb *FlowBuilder) Build(ID string) flow.Operation {
	if fb.Err != nil {
		op := fb.flow.ErrOp(fb.Err)
		return op
	}
	f := fb.flow
	r := fb.registry
	doc := fb.Doc

	if _, ok := fb.nodeTrack[ID]; ok {
		fb.Err = ErrLoop //fmt.Errorf("[%v] Looping through nodes is disabled:", ID)
		op := fb.flow.ErrOp(fb.Err)
		return op
	}
	// loop detector
	fb.nodeTrack[ID] = true
	defer delete(fb.nodeTrack, ID)

	// If flow already has ID just return
	if op, ok := fb.OperationMap[ID]; ok {
		return op
	}

	node := fb.Doc.FetchNodeByID(ID)
	if node == nil {
		op := fb.flow.ErrOp(fmt.Errorf("node not found [%v]", ID))
		return op
	}

	var op flow.Operation
	var inputs []reflect.Type

	switch node.Src {
	case "Portal From":
		nID := node.Prop["portal from"]
		n := doc.FetchNodeByID(nID)
		if n == nil {
			return f.ErrOp(fmt.Errorf("Invalid portal, id: %v", nID))
		}
		// Fetch existing or build new
		op = fb.Build(nID)
		fb.OperationMap[node.ID] = op
		return op
	case "Input":
		inputID, err := strconv.Atoi(node.Prop["input"])
		if err != nil {
			op := f.ErrOp(errors.New("Invalid inputID value, must be a number"))
			fb.OperationMap[node.ID] = op
			return op
		}
		op := f.In(inputID) // By id perhaps
		fb.OperationMap[node.ID] = op
		return op
	case "Var":
		log.Println("Source is a variable")
		var t interface{}
		inputs = []reflect.Type{reflect.TypeOf(t)}
	case "SetVar":
		log.Println("Source is a setvariable")
		var t interface{}
		inputs = []reflect.Type{reflect.TypeOf(t)}
	default:
		log.Println("Loading entry:", node.Src)
		entry, err := r.Entry(node.Src)
		if err != nil {
			op = f.ErrOp(err)
			fb.OperationMap[node.ID] = op
			return op
		}
		inputs = entry.Inputs

	}

	//// Build inputs ////
	param := make([]flow.Data, len(inputs))
	for i := range param {
		l := doc.FetchLinkTo(node.ID, i)
		if l == nil { // No link we fetch the value inserted
			// Direct input entries
			v, err := parseValue(inputs[i], node.DefaultInputs[i])
			if err != nil {
				param[i] = f.ErrOp(err)
				continue
			}
			param[i] = v
			continue
		}
		param[i] = fb.Build(l.From)
	}

	//Switch again
	switch node.Src {
	case "Var":
		op = f.Var(node.Prop["variable name"], param[0])
	case "SetVar":
		op = f.SetVar(node.Prop["variable name"], param[0])
	default:
		op = f.Op(node.Src, param...)
	}

	fb.OperationMap[node.ID] = op
	fb.buildTriggersFor(node, op)

	return op
}

func (fb *FlowBuilder) buildTriggersFor(node *Node, targetOp flow.Operation) error {
	// Process triggers for this node
	triggers := fb.Doc.FetchTriggerFrom(node.ID)
	log.Println("Operation has this triggers:", triggers)
	for _, t := range triggers {
		log.Println("will build for")
		op := fb.Build(t.To)
		// Register the thing here
		fb.flow.Hook(flow.Hook{
			Any: func(name string, triggerOp flow.Operation, triggerTime time.Time, extra ...interface{}) {
				if name != "Error" && name != "Finish" {
					return
				}
				if triggerOp != targetOp {
					log.Printf("ID triggered [%v], I'm t.From: %v", name, t.From)
					return
				}
				exec := false
				for _, o := range t.On {
					if name == o {
						exec = true
						break
					}
				}
				if !exec {
					log.Println("Mismatching trigger, but its a test")
				}
				//op := opfb.flow.GetOp(t.To) // Repeating
				go op.Process(name) // Background
			},
		})
	}
	return nil
}

// Flow returns the build flow
func (fb *FlowBuilder) Flow() *flow.Flow {
	return fb.flow
}

// Or give a string
func parseValue(typ reflect.Type, raw string) (flow.Data, error) {

	if len(raw) == 0 {
		return nil, nil
	}
	if typ == nil {
		var val flow.Data
		err := json.Unmarshal([]byte(raw), &val)
		if err != nil { // Try to unmarshal as a string?
			val = string(raw)
		}
		return val, nil
	}

	var ret flow.Data
	switch typ.Kind() {
	case reflect.Int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, err
		}
		ret = v
	case reflect.String:
		ret = raw
	default:
		if len(raw) == 0 {
			return nil, nil
		}
		//ret = reflect.Zero(typ)

		refVal := reflect.New(typ)
		err := json.Unmarshal([]byte(raw), refVal.Interface())
		if err != nil {
			return nil, err
		}
		ret = refVal.Elem().Interface()
	}
	return ret, nil
}
