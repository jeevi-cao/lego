package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"lego/components/config"
)

func TestApplication_GetConfig(t *testing.T) {
	cf, err := App.GetConfig()
	assert.Equal(t, cf,  (*config.Config)(nil), "config need equal nil")
	assert.NotEqual(t, err, nil, "not init config")
}
