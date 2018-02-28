package registry

// Description of an entry
type Description struct {
	Name string   `json:"name"`
	Desc string   `json:"description"`
	Tags []string `json:"categories"`

	//InputType
	Inputs []DescType `json:"inputs"`
	Output DescType   `json:"output"`

	Extra map[string]interface{} `json:"extra"`
}

//EDescriber helper to batch set properties
type EDescriber struct {
	entries []*Entry
	Err     error
}

// Describer returns a batch of entries for easy manipulation
func Describer(params ...interface{}) *EDescriber {
	ret := &EDescriber{[]*Entry{}, nil}
	for _, el := range params {
		switch v := el.(type) {
		case *EDescriber:
			ret.entries = append(ret.entries, v.entries...)
		case *Entry:
			ret.entries = append(ret.entries, v)
		}
	}
	return ret
}

// Entries return entries from describer
func (d *EDescriber) Entries() []*Entry {
	return d.entries
}

// Description set node description
func (d *EDescriber) Description(m string) *EDescriber {
	for _, e := range d.entries {
		e.Description.Desc = m
	}
	return d

}

//Tags set categories of the group
func (d *EDescriber) Tags(tags ...string) *EDescriber {
	for _, e := range d.entries {
		e.Description.Tags = tags
	}
	return d
}

// Inputs describe inputs
func (d *EDescriber) Inputs(inputs ...string) *EDescriber {
	for _, e := range d.entries {

		for i, dstr := range inputs {
			if i >= len(e.Description.Inputs) { // do nothing
				break // next entry
			}
			curDesc := e.Description.Inputs[i]
			e.Description.Inputs[i] = DescType{curDesc.Type, dstr}
		}
	}
	return d
}

// Output describe the output
func (d *EDescriber) Output(output string) *EDescriber {
	for _, e := range d.entries {
		e.Description.Output = DescType{e.Description.Output.Type, output}
	}
	return d
}

// Extra set extras of the group
func (d *EDescriber) Extra(name string, value interface{}) *EDescriber {
	for _, e := range d.entries {
		e.Description.Extra[name] = value
	}
	return d
}

/*/ Describer
type Describer struct {
	target *Description
}

// Description set module description
func (d *Describer) Description(m string) *Describer {
	d.target.Desc = m
	return d

}

//Tags of the entry
func (d *Describer) Tags(cat ...string) *Describer {
	d.target.Tags = cat
	return d
}

// Inputs description for Inputs
func (d *Describer) Inputs(desc ...string) *Describer {
	for i, dstr := range desc {
		if i >= len(d.target.Inputs) { // do nothing
			return d
		}
		curDesc := d.target.Inputs[i]
		d.target.Inputs[i] = DescType{curDesc.Type, dstr}
	}
	return d
}

// Output description for Input
func (d *Describer) Output(desc string) *Describer {
	d.target.Output = DescType{
		d.target.Output.Type,
		desc,
	}
	return d
}

// Extra information on entry
func (d *Describer) Extra(name string, extra interface{}) *Describer {
	d.target.Extra[name] = extra
	return d
}*/
