package main

import (
	"log"

	"github.com/jcasc/ctl-mc/ctlmc"
)

func main() {
	K := ctlmc.ReadKripke("kripke.json")
	phi, _ := ctlmc.ParseCTL("AG(-\"start\"oAF\"heat\")")
	log.Println(phi)
	log.Println(K.Satisfies(phi))
	log.Printf("%v", K)
}
