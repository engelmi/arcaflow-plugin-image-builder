package dto_test

import (
	"fmt"
	"go.arcalot.io/assert"
	"go.arcalot.io/imagebuilder/internal/dto"
	arcalog "go.arcalot.io/log"
	"log"
	"os"
	"testing"
)

func TestLookupEnvVar(t *testing.T) {
	logger := arcalog.NewLogger(arcalog.LevelInfo, arcalog.NewNOOPLogger())
	envvar_key := "i_hope_this_isnt_used"
	envvar_val := ""

	v := dto.LookupEnvVar(envvar_key, logger)
	assert.Equals(t, v.Msg, fmt.Sprintf("%s not set", envvar_key))
	assert.Equals(t, v.Return_value, "")

	err := os.Setenv(envvar_key, envvar_val)
	if err != nil {
		log.Fatal(err)
	}
	v = dto.LookupEnvVar(envvar_key, logger)
	assert.Equals(t, v.Msg, fmt.Sprintf("%s is empty", envvar_key))
	assert.Equals(t, v.Return_value, "")

	envvar_val = "robot"
	err = os.Setenv(envvar_key, envvar_val)
	if err != nil {
		log.Fatal(err)
	}
	v = dto.LookupEnvVar(envvar_key, logger)
	assert.Equals(t, v.Msg, "")
	assert.Equals(t, v.Return_value, envvar_val)

	err = os.Unsetenv(envvar_key)
	if err != nil {
		log.Fatal(err)
	}
}
