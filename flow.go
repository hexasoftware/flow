package flow

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hexasoftware/flow/registry"
)

// Data interface
type Data = interface{}

// Flow structure
// We could Create a single array of operations
// refs would only mean id, types would be embed in operation
type Flow struct {
	//sync.Mutex // Needed?
	registry   *registry.R
	Data       sync.Map // Should be named, to fetch later
	consts     []Data
	operations []*operation

	// Experimental run Event
	hooks Hooks
}

// New create a new flow
func New() *Flow {
	return &Flow{
		registry:   registry.Global,
		Data:       sync.Map{},
		operations: []*operation{},
		consts:     []Data{},
	}
}

//UseRegistry use the registry specified
func (f *Flow) UseRegistry(r *registry.R) *Flow {
	f.registry = r
	// chain
	return f
}

// Analyse every operations
func (f *Flow) Analyse(w io.Writer, params ...Data) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "Ops analysis:\n")

	for k, op := range f.operations {
		fw := bytes.NewBuffer(nil)
		//fmt.Fprintf(w, "  [%s] (%v)", k, op.name)
		fmt.Fprintf(fw, "  [%v] %s(", k, op.name)
		for j, in := range op.inputs {
			//ref := in.(Op)
			if j != 0 {
				fmt.Fprintf(fw, ", ")
			}
			ires, err := in.Process(params...)
			if err != nil {
				fmt.Fprintf(w, "Operator: %s error#%s\n", op.name, err)
				break
			}
			fmt.Fprintf(fw, " %s(%v)", in.kind, ires)
		}
		fmt.Fprintf(fw, ") - ")
		// Create OpProcessor and execute
		//
		res, err := op.Process(params...)
		if err != nil {
			fmt.Fprintf(fw, "ERR\n")
		}
		fmt.Fprintf(fw, "%v\n", res)

		fmt.Fprintf(w, "%s", fw.String())
	}
}

/////////////////////////////
// Serializers inspectors
//////////////////////

func (f *Flow) String() string {
	ret := bytes.NewBuffer(nil)
	fmt.Fprintf(ret, "Flow\n")
	// consts
	fmt.Fprintf(ret, "consts:\n")
	for i, v := range f.consts {
		fmt.Fprintf(ret, "  [%v] %v\n", i, v)
	}
	fmt.Fprintf(ret, "data:\n") // Or variable

	f.Data.Range(func(k, v interface{}) bool {
		fmt.Fprintf(ret, "  [%v] %v\n", k, v)
		return true
	})

	fmt.Fprintf(ret, "operations:\n")
	for k, op := range f.operations {
		fmt.Fprintf(ret, "  [%v] %s(", k, op.name)
		for j, in := range op.inputs {
			if j != 0 {
				fmt.Fprintf(ret, ", ")
			}
			if in.kind == "const" {
				v, _ := in.Process()
				fmt.Fprintf(ret, "%s[%v](%v)", in.kind, j, v)
				continue
			}
			// Find operation index
			for t := range f.operations {
				if f.operations[t] == in {
					fmt.Fprintf(ret, "%s[%v]", in.kind, t)
					break
				}
			}
		}
		fmt.Fprintf(ret, ")\n")
	}

	return ret.String()
}

//////////////////////////////////////////////
// Experimental event hooks
////////////////

// Hook attach the node event hooks
func (f *Flow) Hook(hook Hook) {
	f.hooks.Attach(hook)
}
