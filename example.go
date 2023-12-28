package main

import (
	"log"

	. "github.com/jcasc/ctl-mc/ctlmc"
)

func main() {
	K := MakeKripke(
		[]int{0},
		append([][]int{{1, 2}, {4}, {5, 0}, {2, 0, 3}, {1, 2}, {6}, {3}}, make([][]int, 7000)...),
		map[int]map[AP]bool{
			1: {"start": true, "error": true}, 2: {"close": true}, 3: {"close": true, "heat": true},
			4: {"start": true, "error": true, "close": true}, 5: {"start": true, "close": true}, 6: {"start": true, "close": true, "heat": true}},
	)
	for i, v := range K.R {
		if v == nil {
			K.R[i] = make([]int, 7000)
			for j := range K.R[i] {
				K.R[i][j] = j
			}
		}
	}
	for i := 0; i < len(K.R); i++ {
		K.R[0] = append(K.R[0], i)
	}
	phi_ := PhiAG(PhiImpl(AP("start"), PhiAF(AP("heat"))))
	phi, _ := ParseCTL("A                      G(-\"start\"oAF\"heat\")")
	log.Println(phi)
	log.Println(phi == phi_)
	log.Println(K.Satisfies(phi))
}
