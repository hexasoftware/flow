package flow_test

import (
	"bytes"
	"encoding/json"
	"testing"

	vecasm "github.com/gohxs/vec-benchmark/asm"

	"github.com/hexasoftware/flow"
	"github.com/hexasoftware/flow/internal/assert"
	"github.com/hexasoftware/flow/registry"
)

func init() {
	assert.Quiet = true
}

func TestInput(t *testing.T) {
	a := assert.A(t)
	f := flow.New()

	opIn := f.In(0)
	a.NotEq(opIn, nil, "input err should not be nil")

	d, err := opIn.Process([]float32{2, 2, 2})
	a.Eq(d, []float32{2, 2, 2}, "array should be equal")

	op := f.Op("vecadd", []float32{1, 1, 1}, opIn)
	_, err = op.Process([]float32{1, 2, 3})
	a.Eq(err, nil, "result should not error")
	a.NotEq(op, nil, "operation should not be nil")

	d, err = op.Process([]float32{2, 2, 2})
	a.Eq(err, nil, "should not error passing an input")

	a.Eq(d, []float32{3, 3, 3}, "array should be equal")

}

func TestSerialize(t *testing.T) {
	// Does not text yet
	f := flow.New()
	var1 := f.Var("var1", []float32{4, 4, 4})

	c1 := f.Const([]float32{1, 2, 3})
	c2 := f.Const([]float32{2, 2, 2})

	op1 := f.Op("vecmul", // op:0 - expected: [12,16,20,24]
		f.Var("vec1", []float32{4, 4, 4, 4}),
		f.Op("vecadd", // op:1 - expected: [3,4,5,6]
			f.Const([]float32{1, 2, 3, 4}),
			f.Const([]float32{2, 2, 2, 2}),
		),
	)
	mul1 := f.Op("vecmul", c1, op1)       // op:2 - expected 12, 32, 60, 0
	mul2 := f.Op("vecmul", mul1, var1)    // op:3 - expected 48, 128, 240, 0
	mul3 := f.Op("vecmul", c2, mul2)      // op:4 - expected 96, 256, 480, 0
	mul4 := f.Op("vecmul", mul3, f.In(0)) // op:5 - expected 96, 512, 1440,0

	s := bytes.NewBuffer(nil)
	f.Analyse(s, []float32{1, 2, 3, 4})
	t.Log(s)
	res, _ := mul4.Process([]float32{1, 2, 3, 4})

	t.Log("Res:", res)
	t.Log("Flow:\n", f)

	ret := bytes.NewBuffer(nil)
	e := json.NewEncoder(ret)
	e.SetIndent(" ", " ")
	e.Encode(f)

	// Deserialize

	t.Log("Flow:", ret)

}
func TestConst(t *testing.T) {
	a := assert.A(t)
	f := flow.New()

	c := f.Const(1)
	res, err := c.Process()
	a.Eq(res, 1, "It should be one")
	a.Eq(err, nil, "const should not error")
}
func TestOp(t *testing.T) {
	a := assert.A(t)
	f := flow.New()

	add := f.Op("vecadd",
		f.Op("vecmul",
			[]float32{1, 2, 3},
			[]float32{2, 2, 2},
		),
		[]float32{1, 2, 3},
	)
	res, err := add.Process()
	a.Eq(err, nil)

	test := []float32{3, 6, 9}
	a.Eq(test, res)
}

/*
* TODO: Create variable test
func TestVariable(t *testing.T) {
	a := assert.A(t)
	f := flow.New()
	v := f.Var("v1", 1)

	res, err := v.Process()
	a.Eq(err, nil)
	a.Eq(res, 1)

	v.Set(2)
	res, err = v.Process()
	a.Eq(err, nil)
	a.Eq(res, 2)
}*/

// Test context
func TestCache(t *testing.T) {
	a := assert.A(t)
	f := flow.New()
	{
		r := f.Op("inc")
		a.NotEq(r, nil, "should not error giving operation")

		for i := 1; i < 5; i++ {
			res, err := r.Process()
			a.Eq(err, nil)
			a.Eq(res, i)
		}
	}
	{
		var res flow.Data
		inc := f.Op("inc")

		add := f.Op("add", inc, inc)
		res, _ = add.Process() // 1+1
		assert.Eq(t, res, 2)
		res, _ = add.Process() // 2+2
		assert.Eq(t, res, 4)
	}
}

// XXX: Create proper test
/*func TestHandler(t *testing.T) {
	f, op := prepareComplex()
	f.Hook(flow.Hook{
		Wait:   func(op flow.Operation, triggerTime time.Time) { t.Logf("[%s] Wait", op) },
		Start:  func(op flow.Operation, triggerTime time.Time) { t.Logf("[%s]Start", op) },
		Finish: func(op flow.Operation, triggerTime time.Time, res flow.Data) { t.Logf("[%s] Finish %v", op, res) },
		Error:  func(op flow.Operation, triggerTime time.Time, err error) { t.Logf("[%s] Error %v", op, err) },
	})
	op.Process()
}*/

func TestLocalRegistry(t *testing.T) {
	a := assert.A(t)

	r := registry.New()
	e := r.Add("test", func() string { return "" })
	a.NotEq(e, nil, "registered in a local register")

	f := flow.New()
	f.UseRegistry(r)
	op := f.Op("test")
	a.NotEq(op, nil, "operation should be valid")

	op = f.Op("none")
	a.NotEq(op, nil, "operation should not be nil")
	_, err := op.Process()

	a.NotEq(err, nil, "flow should contain an error")
}

func init() {
	registry.Add("vecmul", VecMul)
	registry.Add("vecadd", VecAdd)
	registry.Add("vecdiv", VecDiv)
	registry.Add("inc", Inc)
	registry.Add("add", Add)
}

func prepareComplex() (*flow.Flow, flow.Operation) {
	vecsize := 5
	v1 := make([]float32, vecsize)
	v2 := make([]float32, vecsize)
	for i := range v1 {
		v1[i], v2[i] = float32(i+1), 2
	}

	f := flow.New()
	f1 := f.Var("f1", v1)
	f2 := f.Var("f2", v2)

	mul := f.Op("vecmul", f1, f2)      // Doubles 2,4,6,8...
	add := f.Op("vecadd", mul, f2)     // Sum 4,8,10,12...
	mul2 := f.Op("vecmul", mul, add)   // mul again
	mul3 := f.Op("vecmul", mul2, f1)   // mul with f1
	div1 := f.Op("vecdiv", mul3, mul2) // div

	return f, div1
}

func VecMul(a, b []float32) []float32 {

	sz := Min(len(a), len(b))

	out := make([]float32, sz)
	vecasm.VecMulf32x8(a, b, out)
	return out
}
func VecAdd(a, b []float32) []float32 {
	sz := Min(len(a), len(b))
	out := make([]float32, sz)
	for i := 0; i < sz; i++ {
		out[i] = a[i] + b[i]
	}
	return out
}
func VecDiv(a, b []float32) []float32 {
	sz := Min(len(a), len(b))
	out := make([]float32, sz)
	for i := 0; i < sz; i++ {
		out[i] = a[i] / b[i]
	}
	return out
}

// ScalarInt
// Every time this operator is called we increase 1
// Constructor
func Inc() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}
func Add(a, b int) int {
	return a + b
}

// Utils
func Min(p ...int) int {
	min := p[0]
	for _, v := range p[1:] {
		if min < v {
			min = v
		}
	}
	return min
}
