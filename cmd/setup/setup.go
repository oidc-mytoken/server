package main

import (
	"fmt"
	"io/ioutil"

	"github.com/zachmann/mytoken/internal/jws"
)

func main() {
	sk, _ := jws.GenerateRSAKeyPair()
	str := jws.ExportRSAPrivateKeyAsPemStr(sk)
	filepath := "/tmp/mytoken.key"
	ioutil.WriteFile(filepath, []byte(str), 0600)
	fmt.Printf("Wrote key to '%s'. Copy the keyfile to a secure location.\n", filepath)
}
