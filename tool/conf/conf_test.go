// package gt

package conf

import (
	"log"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	var config = Configger()
	log.Println("config read test: ", config.GetString("app.port"))
}

// can not read privilege field
type dbas struct {
	MaxIdleConn int
	MaxOpenConn int
	User        string
	Password    string
	host        string
	name        string
}

func TestConfig_GetStruct(t *testing.T) {
	dba := &dbas{}
	GetStruct("app.db", dba)
	t.Log(dba)
}

func TestConfigger(t *testing.T) {

	dir, _ := os.Getwd()
	t.Log(dir)
	t.Log(os.Getenv("GOPATH"))
	mode := GetString("app.devMode")
	t.Log(mode)
}
