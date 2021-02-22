package logging

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/grafana/loki/pkg/promtail/config"
	"gopkg.in/yaml.v2"
)

func TestLoadSrapeConfig(t *testing.T) {
	var config config.Config

	filepath := path.Join("../../assets", PromtailScrapeConfig)
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Errorf("error loading %v", err)
	}
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		t.Errorf("error parsing yaml %v", err)
	}
}
