package goginx

import (
	"log"
	"testing"
)

func TestStart(t *testing.T) {
	engine := Start(2)
	log.Println(engine)
}
