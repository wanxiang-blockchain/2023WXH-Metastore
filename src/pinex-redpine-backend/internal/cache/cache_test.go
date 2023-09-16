package cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	InitJsonCode()
	InitGlobalCache(MAP)
	m.Run()
}

var called bool

func testFunc(var1 int, var2 string) (string, error) {
	called = true
	val := fmt.Sprintf("%d/%s", var1, var2)
	return val, nil
}

type Teststruct struct {
	I      int
	S      string
	B      []byte
	Struct *struct {
		I int
	}
}

func testFunc2(var1 *int, var2 []byte) ([]byte, *Teststruct) {
	called = true
	return []byte{1}, &Teststruct{
		I: 123,
		S: "pinex",
		B: nil,
		Struct: &struct{ I int }{
			I: 456,
		},
	}
}

func TestWithCache(t *testing.T) {
	f := WithCache(testFunc)
	res, err := f(123, "pinex")
	require.Equal(t, res, fmt.Sprintf("%d/%s", 123, "pinex"))
	require.Empty(t, err)
	require.True(t, called)
	called = false
	res, err = f(123, "pinex")
	require.Equal(t, res, fmt.Sprintf("%d/%s", 123, "pinex"))
	require.Empty(t, err)
	require.False(t, called)
}

func TestWithCacheComplex(t *testing.T) {
	called = false
	f2 := WithCache(testFunc2)
	var i int = 1024
	expectdSlice := []byte{1}
	expectedStruct := &Teststruct{
		I: 123,
		S: "pinex",
		B: nil,
		Struct: &struct{ I int }{
			I: 456,
		},
	}
	outSlice, outStruct := f2(&i, []byte("pinex"))
	require.Equal(t, outSlice, expectdSlice)
	require.Equal(t, outStruct, expectedStruct)
	require.True(t, called)

	called = false
	outSlice, outStruct = f2(&i, []byte("pinex"))
	require.Equal(t, outSlice, expectdSlice)
	require.Equal(t, outStruct, expectedStruct)
	require.False(t, called)
}
