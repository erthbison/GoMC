package server

import (
	"testing"
)

func TestRegisterModule(t *testing.T) {
	for i, test := range registerModule {
		srv := NewServer()
		var isErr bool
		for _, mod := range test.modules {
			err := srv.RegisterModule(mod.name, mod.mod)
			isErr = err != nil
			if isErr {
				break
			}
		}
		if test.err != isErr {
			if test.err {
				t.Errorf("Test %v: Expected to receive an error", i)
			} else {
				t.Errorf("Test %v: Expected not to receive an error", i)
			}
		}
	}
}

var registerModule = []struct {
	modules []struct {
		name string
		mod  any
	}
	err bool
}{
	{
		[]struct {
			name string
			mod  any
		}{{"1", struct{}{}}, {"2", struct{}{}}, {"3", struct{}{}}},
		false,
	},
	{
		// Multiple modules with the same name
		[]struct {
			name string
			mod  any
		}{{"1", struct{}{}}, {"1", struct{}{}}, {"1", struct{}{}}},
		true,
	},
}
