package registry

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
)

// Global
var (
	Global       = New()
	Descriptions = Global.Descriptions
	GetEntry     = Global.Entry
	Add          = Global.Add
)

// M Alias for map[string]interface{}
type M = map[string]interface{}

// R the function registry
type R struct {
	entries map[string]*Entry
}

// New creates a new registry
func New() *R {
	r := &R{map[string]*Entry{}}
	// create a base function here?
	return r
}

// Clone an existing registry
func (r *R) Clone() *R {
	newR := &R{map[string]*Entry{}}
	for k, v := range r.entries {
		newR.entries[k] = v
	}
	return newR
}

// Merge other registry
func (r *R) Merge(or *R) {
	for k, v := range or.entries {
		r.entries[k] = v
	}
}

// Add function to registry
func (r *R) Add(args ...interface{}) *EDescriber {

	var nextName string
	var name string
	var fn interface{}

	d := &EDescriber{[]*Entry{}, nil}
	for _, a := range args {
		switch f := a.(type) {
		case string:
			if nextName != "" {
				d.Err = ErrNotAFunc
				return d
			}
			nextName = f
			continue
		default:
			fn = a
			// consume name
			if nextName != "" {
				name = nextName
				nextName = ""
			} else {
				if reflect.TypeOf(fn).Kind() != reflect.Func {
					d.Err = ErrNotAFunc
					return d
				}
				// Automatic naming
				name = runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
				name = path.Ext(name)[1:]
			}
		}

		e, err := r.register(name, fn)
		if err != nil {
			d.Err = err
			return d
		}
		d.entries = append(d.entries, e)
	}
	if nextName != "" {
		d.Err = ErrNotAFunc
	}

	return d
}

//Register should be a func only register
func (r *R) register(name string, v interface{}) (*Entry, error) {
	e, err := NewEntry(r, v)
	if err != nil {
		return nil, err
	}
	r.entries[name] = e
	return e, nil
}

// Get an entry
func (r *R) Get(name string, params ...interface{}) (interface{}, error) {
	e, ok := r.entries[name]
	if !ok {
		return nil, fmt.Errorf("Entry not found '%s'", name)
	}
	v := e.fn
	// We already know this is a function
	// and that returns 1 or more values

	vtyp := reflect.TypeOf(v)
	if vtyp.NumOut() == 0 || vtyp.Out(0).Kind() != reflect.Func {
		return v, nil
	}
	// Constructor
	fparam := make([]reflect.Value, len(params))
	for i := range params {
		fparam[i] = reflect.ValueOf(params[i])
	}
	// Call the func and return the thing
	v = reflect.ValueOf(v).Call(fparam)[0].Interface()

	return v, nil
}

// Entry fetches entries from the register
func (r *R) Entry(name string) (*Entry, error) {
	e, ok := r.entries[name]
	if !ok {
		return nil, ErrNotFound
	}
	return e, nil
}

//Describe named fn

// Descriptions Description list
func (r *R) Descriptions() (map[string]Description, error) {
	ret := map[string]Description{}
	for k, e := range r.entries {
		ret[k] = e.Description
	}
	return ret, nil
}
