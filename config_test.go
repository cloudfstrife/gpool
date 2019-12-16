package gpool

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCreateDefaultConfig(t *testing.T) {
	want := Config{
		InitialPoolSize:      5,
		MinPoolSize:          2,
		MaxPoolSize:          15,
		AcquireRetryAttempts: 5,
		AcquireIncrement:     5,
		TestDuration:         1000,
		TestOnGetItem:        false,
		Params:               make(map[string]string),
	}
	got := DefaultConfig()
	if !cmp.Equal(want, got) {
		t.Errorf("TestCreateDefaultConfig Failed \n want : %#v \n got %#v", want, got)
	}
}

func TestString(t *testing.T) {
	want := "InitialPoolSize : 5 \n MinPoolSize : 2 \n MaxPoolSize : 15 \n AcquireRetryAttempts : 5 \n AcquireIncrement : 5 \n TestDuration : 1000 \n TestOnGetItem : false \nParams:\n\tserver : 127.0.0.1 \n"
	defaultConfig := DefaultConfig()
	defaultConfig.Params["server"] = "127.0.0.1"
	got := defaultConfig.String()
	if !cmp.Equal(want, got) {
		t.Errorf("TestString Failed \n want : %#v \n got  : %#v", want, got)
	}
}

func TestLoadToml(t *testing.T) {
	want := Config{
		InitialPoolSize:      10,
		MinPoolSize:          4,
		MaxPoolSize:          30,
		AcquireRetryAttempts: 10,
		AcquireIncrement:     10,
		TestDuration:         60000,
		TestOnGetItem:        false,
		Params: map[string]string{
			"host": "127.0.0.1",
			"port": "11211",
		},
	}
	v := DefaultConfig()
	err := v.LoadToml("./testing/testing_config.toml")
	if err != nil {
		t.Errorf("TestLoadToml Failed : %#v", err)
	}
	if !cmp.Equal(want, v) {
		t.Errorf("TestLoadToml Failed \n want : %#v \n got  : %#v", want, v)
	}
}

func TestLoadTomlNotExsist(t *testing.T) {
	v := DefaultConfig()
	err := v.LoadToml("./testing/testing.toml")
	if err == nil {
		t.Errorf("TestLoadTomlNotExsist Failed \n want : PathError \n got  : %#v", err)
	}
}

func TestLoadTomlNotToml(t *testing.T) {
	want := errors.New("FILE TYPE ERROR")
	v := DefaultConfig()
	err := v.LoadToml("./testing/goline.report")
	if err == nil {
		t.Errorf("TestLoadTomlNotExsist Failed \n want : %#v \n got  : %#v", want, err)
	}
}
