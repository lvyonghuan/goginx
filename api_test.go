package goginx

import (
	"log"
	"testing"
)

func TestInit(t *testing.T) {
	engine := Init(2)
	log.Println(engine)
}

func TestStart(t *testing.T) {
	engine := Init(3)
	engine.Start()
}
