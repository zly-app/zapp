package serializer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Temp struct {
	A string `json:"AA"`
}

func TestBaseMarshal(t *testing.T) {
	serializer := GetSerializer(BaseSerializerName)

	marshalArgs := []struct {
		A      interface{}
		Expect string
	}{
		{"hello", "hello"},
		{1, "1"},
		{true, "true"},
		{float64(1.2), "1.2"},
		{[]byte("hello"), "hello"},
		{Temp{"xx"}, `{"AA":"xx"}`},
	}
	for i := range marshalArgs {
		var buf bytes.Buffer
		err := serializer.Marshal(marshalArgs[i].A, &buf)
		assert.Nil(t, err)
		assert.Equal(t, marshalArgs[i].Expect, buf.String())
	}
	for i := range marshalArgs {
		bs, err := serializer.MarshalBytes(marshalArgs[i].A)
		assert.Nil(t, err)
		assert.Equal(t, marshalArgs[i].Expect, string(bs))
	}
}

func TestBaseUnmarshal(t *testing.T) {
	serializer := GetSerializer(BaseSerializerName)

	type Args struct {
		Data   string
		A      interface{}
		Expect interface{}
	}
	unmarshalArgs := []func() Args{
		func() Args {
			var a string
			expect := "hello"
			return Args{"hello", &a, &expect}
		},
		func() Args {
			var a int
			expect := 1
			return Args{"1", &a, &expect}
		},
		func() Args {
			var a bool
			expect := true
			return Args{"1", &a, &expect}
		},
		func() Args {
			var a float64
			expect := float64(1.2)
			return Args{"1.2", &a, &expect}
		},
		func() Args {
			var a []byte
			expect := []byte("hello")
			return Args{"hello", &a, &expect}
		},
		func() Args {
			var a Temp
			expect := Temp{"xx"}
			return Args{`{"AA":"xx"}`, &a, &expect}
		},
	}
	for i := range unmarshalArgs {
		args := unmarshalArgs[i]()
		err := serializer.Unmarshal(strings.NewReader(args.Data), args.A)
		assert.Nil(t, err)
		assert.Equal(t, args.Expect, args.A)
	}
}
