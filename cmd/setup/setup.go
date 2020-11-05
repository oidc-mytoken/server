package main

import (
	"io/ioutil"
	"log"

	"github.com/zachmann/mytoken/internal/config"

	"github.com/zachmann/mytoken/internal/jws"
)

func main() {
	config.LoadForSetup()
	sk, _, err := jws.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}
	str := jws.ExportPrivateKeyAsPemStr(sk)
	filepath := config.Get().Signing.KeyFile
	ioutil.WriteFile(filepath, []byte(str), 0600)
	log.Printf("Wrote key to '%s'.\n", filepath)
}
