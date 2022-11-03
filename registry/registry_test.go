package registry_test

import (
	"strings"
	"testing"

	"github.com/hexasoftware/flow/internal/assert"
	"github.com/hexasoftware/flow/registry"

	"github.com/gohxs/prettylog"
)

func init() {
	prettylog.Global()
}
func TestRegistry(t *testing.T) {
	a := assert.A(t)

	r := registry.New()
	r.Add("vecadd", dummy1)

	e, err := r.Entry("vecadd")
	a.Eq(err, nil, "fetching entry")

	d := e.Description

	a.Eq(len(d.Inputs), 2, "should have 2 outputs")
	a.Eq(d.Output.Type, "[]float32", "output type")

	t.Log(d)

}

func TestEntry(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	e, err := r.Entry("bogus")

	a.NotEq(err, nil, "should get an error")
	a.Eq(e, nil, "entry should be nil")

}

func TestRegisterDuplicate(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	r.Add("func", func(a, b int) int { return 0 })
	d := r.Add("func", func(b int) int { return 0 })
	a.Eq(d.Err, nil, "should allow duplicate")
}
func TestRegisterOutput(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	d := r.Add("func", func(a int) {})
	a.Eq(d.Err, nil, "should not give output error")
}

func TestRegisterInvalidInput(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	d := r.Add("func", func() int { return 0 })
	a.Eq(d.Err, nil, "should register a func without params")
}

func TestRegistryGet(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	d := r.Add("func", func() int { return 0 })
	a.Eq(d.Err, nil, "should register func")

	fn, err := r.Get("func")
	a.Eq(err, nil, "should fetch a function")
	a.NotEq(fn, nil, "fn should not be nil")
}

func TestRegistryGetEmpty(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	fn, err := r.Get("notfoundfunc")
	a.NotEq(err, nil, "should fail fetching a unregistered func")
	a.Eq(fn, nil, "should return a <nil> func")
}
func TestRegistryGetConstructor(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	d := r.Add("func", func() func() int {
		return func() int {
			return 0
		}
	})
	a.Eq(d.Err, nil, "should register the constructor func")

	ifn, err := r.Get("func")
	a.Eq(err, nil, "get should not error")

	fn, ok := ifn.(func() int)
	a.Eq(ok, true, "function should be the constructor type")

	ret := fn()
	a.Eq(ret, 0, "function should return 0")
}

func TestRegistryGetConstructorParam(t *testing.T) {
	a := assert.A(t)

	r := registry.New()
	d := r.Add("func2", func(a, b int) func() int {
		return func() int {
			return a + b
		}
	})
	a.Eq(d.Err, nil)
	ifn, err := r.Get("func2", 1, 1)
	a.Eq(err, nil, "should not fail passing params to constructor")
	a.NotEq(ifn, nil, "should return a function")

	fn, ok := ifn.(func() int)
	a.Eq(ok, true, "function should match the type")

	ret := fn()
	a.Eq(ret, 2, "function should execute")

}

func TestDescriptions(t *testing.T) {
	a := assert.A(t)

	r := registry.New()
	r.Add("vecadd", dummy1)
	r.Add("vecstr", dummy2)

	d, err := r.Descriptions()
	a.Eq(err, nil, "should fetch Descriptions")
	a.Eq(len(d), 2, "should contain 2 descriptions")

	t.Log(d)
}
func TestClone(t *testing.T) {
	a := assert.A(t)
	r := registry.New()
	r.Add("vecadd", dummy1)

	desc, err := r.Descriptions()
	a.Eq(err, nil, "should not error fetching description")

	a.Eq(len(desc), 1, "should contain 1 descriptions")

	r2 := r.Clone()
	r2.Add("vecmul", dummy2)
	a.Eq(len(desc), 1, "should contain 1 descriptions")

	d2, err := r2.Descriptions()
	a.Eq(err, nil, "should not error fetching descriptions")

	a.Eq(len(d2), 2, "should contain 2 descriptions")
	_, ok := d2["vecmul"]
	a.Eq(ok, true, "should be equal")
}

func TestNotAFunc(t *testing.T) {
	a := assert.A(t)
	r := registry.New()

	d := r.Add("test", []string{})
	a.Eq(d.Err, registry.ErrNotAFunc, "should give a func error")

	d = r.Add("test")
	a.Eq(d.Err, registry.ErrNotAFunc, "should give a func error")

	d = r.Add([]string{})
	a.Eq(d.Err, registry.ErrNotAFunc, "should give a func error")

}

func TestRegisterErr(t *testing.T) {
	r := registry.New()

	d := r.Add("name", "notfunc")
	assert.Eq(t, d.Err, registry.ErrNotAFunc, "should give a func error")
}

func TestMerge(t *testing.T) {
	a := assert.A(t)
	r1 := registry.New()
	r1.Add(strings.Join)

	r2 := registry.New()
	r2.Add(strings.Split)

	r2.Merge(r1)

	m, err := r2.Descriptions()
	a.Eq(err, nil, "should not error fetching descriptions")
	a.Eq(len(m), 2, "Number of descriptions should be 2")

}

/*func TestAddEntry(t *testing.T) {
	a := assert.A(t)
	r := registry.New()

	g, err := r.Register("vecadd", dummy1)
	//a.Eq(err, ErrNotAFunc, "should error giving a func")
	a.Eq(len(g), 0, "should contain 1 entry")

	r.Add(dummy2)

	d, err := r.Descriptions()
	a.Eq(err, nil, "should not error fetching descriptions")
	a.Eq(len(d), 2, "should contain 2 descriptions")

}*/

func dummy1([]float32, []float32) []float32 {
	return []float32{1, 3, 3, 7}
}
func dummy2([]float32) string {
	return ""
}
