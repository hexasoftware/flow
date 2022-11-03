package registry

import (
	"fmt"
	"reflect"
)

//DescType type Description
type DescType struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// Entry contains a function description params etc
type Entry struct {
	registry    *R
	fn          interface{}
	Inputs      []reflect.Type
	Output      reflect.Type
	Description Description
	Err         error
}

// NewEntry creates and describes a New Entry
func NewEntry(r *R, fn interface{}) (*Entry, error) {
	e := &Entry{registry: r, fn: fn}

	fntyp := reflect.TypeOf(e.fn)
	if fntyp.Kind() != reflect.Func {
		return nil, ErrNotAFunc
	}
	var Output DescType
	if fntyp.NumOut() > 0 {
		outTyp := fntyp.Out(0)
		if outTyp.Kind() == reflect.Func {
			outTyp.Out(0)
		}
		Output = DescType{fmt.Sprint(outTyp), ""}
		e.Output = outTyp // ?

	}

	fnTyp := reflect.TypeOf(e.fn)
	nInputs := fnTyp.NumIn()

	Inputs := make([]DescType, nInputs)
	for i := 0; i < nInputs; i++ {
		inTyp := fnTyp.In(i)
		Inputs[i] = DescType{fmt.Sprint(inTyp), ""}
		e.Inputs = append(e.Inputs, inTyp) // ?
	}

	e.Description = Description{
		Tags:   []string{"generic"},
		Inputs: Inputs,
		Output: Output,
		Extra:  map[string]interface{}{},
	}
	return e, nil
}

// Describer return a description builder
func (e *Entry) Describer() *EDescriber {
	return Describer(e)
}
