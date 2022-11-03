package flowbuilder

// Node flow-ui node representation
type Node struct {
	ID            string            `json:"id"`
	Src           string            `json:"src"`
	Label         string            `json:"label"`
	DefaultInputs map[int]string    `json:"defaultInputs"`
	Prop          map[string]string `json:"prop"`
}

// Link that joins two nodes
type Link struct {
	From string `json:"from"`
	To   string `json:"to"`
	In   int    `json:"in"`
}

// Trigger that join two nodes on state change
type Trigger struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	On   []string `json:"on"`
}

// FlowDocument flow document
type FlowDocument struct {
	Nodes    []Node    `json:"nodes"`
	Links    []Link    `json:"links"`
	Triggers []Trigger `json:"triggers"`
}

// FetchNodeByID retrieve a node by its ID
func (fd *FlowDocument) FetchNodeByID(ID string) *Node {
	for _, n := range fd.Nodes {
		if n.ID == ID {
			return &n
		}
	}
	return nil
}

// FetchNodeBySrc loops the nodes and filter src
func (fd *FlowDocument) FetchNodeBySrc(src string) []Node {
	ret := []Node{}
	for _, n := range fd.Nodes {
		if n.Src == src {
			ret = append(ret, n)
		}
	}
	return ret
}

// FetchTriggerFrom  fetch nodes where trigger comes from ID
func (fd *FlowDocument) FetchTriggerFrom(ID string) []Trigger {
	ret := []Trigger{}
	for _, t := range fd.Triggers {
		if t.From != ID {
			continue
		}
		ret = append(ret, t)
	}
	return ret
}

// FetchLinksTo fetch all links to node ID
func (fd *FlowDocument) FetchLinksTo(ID string) []Link {
	ret := []Link{}
	for _, l := range fd.Links {
		if l.To != ID {
			continue
		}
		ret = append(ret, l)
	}
	return ret
}

// FetchLinkTo fetch a specific link to ID or return nil of none
func (fd *FlowDocument) FetchLinkTo(ID string, n int) *Link {
	for _, l := range fd.Links {
		if l.To != ID || l.In != n {
			continue
		}
		return &l
	}
	return nil
}

// FetchNamedPortalIn fetch a portal in node, return nil of none
func (fd *FlowDocument) FetchNamedPortalIn(name string) *Node {
	for _, n := range fd.Nodes {
		if n.Src == "Portal In" && n.Prop["portal name"] == name {
			return &n
		}
	}
	return nil
}
