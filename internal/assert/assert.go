package assert

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// Global setup
var (
	Quiet = false
)

//A return a asserter with testing.T
func A(t *testing.T) *Checker {
	return &Checker{T: t}
}

// Eq global
func Eq(t *testing.T, a, b interface{}, params ...interface{}) *Checker {
	return A(t).Eq(a, b, params...)
}

// NotEq global
func NotEq(t *testing.T, a, b interface{}, params ...interface{}) *Checker {
	return A(t).NotEq(a, b, params...)
}

// Checker test checking helper
type Checker struct {
	*testing.T
}

//Eq check if equal
func (t *Checker) Eq(a, b interface{}, params ...interface{}) *Checker {
	msg := fmt.Sprintf("(%s) Expect Eq '%v' got '%v'", fmt.Sprint(params...), b, a)

	if (a == nil || (reflect.ValueOf(a).Kind() == reflect.Ptr && reflect.ValueOf(a).IsNil())) &&
		(b == nil || (reflect.ValueOf(b).Kind() == reflect.Ptr && reflect.ValueOf(b).IsNil())) {
		t.pass(msg)
		return t
	}
	if !reflect.DeepEqual(a, b) {
		t.fail(msg)
	}
	t.pass(msg)
	return t
}

//NotEq check if different
func (t *Checker) NotEq(a, b interface{}, params ...interface{}) *Checker {
	msg := fmt.Sprintf("(%s) Expect NotEq '%v' got '%v'", fmt.Sprint(params...), b, a)
	if (a == nil || (reflect.ValueOf(a).Kind() == reflect.Ptr && reflect.ValueOf(a).IsNil())) &&
		(b == nil || (reflect.ValueOf(b).Kind() == reflect.Ptr && reflect.ValueOf(b).IsNil())) {
		t.fail(msg)
	}
	if reflect.DeepEqual(a, b) {
		t.fail(msg)
	}
	t.pass(msg)
	return t
}

func (t *Checker) fail(msg string) {
	file, line := getCaller(3)
	fmt.Fprintf(os.Stderr, "    %s:%-4s \033[31m[FAIL] \033[01;31m%s\033[0m\n", file, fmt.Sprintf("%d:", line), msg)
	t.FailNow()
	//t.FailNow()
}
func (t *Checker) pass(msg string) {

	if Quiet {
		return
	}
	file, line := getCaller(3)
	file = filepath.Base(file)
	fmt.Fprintf(os.Stderr, "    %s:%-4s \033[32m[PASS]\033[m %s\n", file, fmt.Sprintf("%d:", line), msg)
}

func getCaller(offs int) (string, int) {
	var file string
	var line int
	for count := offs; ; count++ {
		_, file, line, _ = runtime.Caller(count)
		file = filepath.Base(file)
		if file != "assert.go" {
			break
		}
	}
	return file, line
}
