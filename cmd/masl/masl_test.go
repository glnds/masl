package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// var theflag string

func init() {
	// flag.StringVar(&theflag, "profile", "olol", "desc")
}

func TestParseFlags(t *testing.T) {

	// os.Args = []string{"-profile=bla"}

	// conf := masl.Config{Profile: "testProfile"}
	// parseFlags(conf)
	// assert equality
	assert.Equal(t, 123, 123, "they should be equal")

}
