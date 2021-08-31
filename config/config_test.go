package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	assertion := assert.New(t)
	cfg, err := LoadFile("../config.yml")

	assertion.Nil(err)
	assertion.Equal(1, len(cfg.App))
	assertion.Equal("app", cfg.App[0].Name)
	assertion.NotEmpty(cfg.App[0].Format)
	assertion.Equal(1, len(cfg.App[0].SourceFiles))
	assertion.Equal("foo", cfg.App[0].ExternalLabels["foo"])
	assertion.NotEmpty(cfg.App[0].HistogramBuckets)
}
