package main

import (
	"log"

	"github.com/jcasc/ctl-mc/ctlmc"
)

func main() {
	K, err := ctlmc.ReadKripke("kripke.json")
	if err != nil {
		log.Fatal(err)
	}
	phi, err := ctlmc.ParseCTL("AG(-\"start\"+AF\"heat\")")
	if err != nil {
		log.Fatalf("failed parsing CTL formula: %v", err)
	}
	log.Println(phi)
	log.Println(K.Satisfies(phi))
	log.Printf("%v", K)
}
