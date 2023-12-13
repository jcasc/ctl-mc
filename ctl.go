package main

import (
	"fmt"
	"log"
)

type phi1 struct {
	phi1 Phi
}

type phi2 struct {
	phi1 Phi
	phi2 Phi
}

type (
	AP      string
	_PhiNot phi1
	PhiAnd  phi2
	PhiEX   phi1
	PhiEU   phi2
	PhiAU   phi2
	PhiTrue struct{}
)

// Phi represents a CTL formula which can be validated on a Kripke structure
type Phi interface {
	valid(*Kripke, int) bool
}

// p
func (p AP) valid(K *Kripke, s int) bool {
	_, ok := K.L[s][p]
	return ok
}

func (p AP) String() string {
	return fmt.Sprintf("\"%v\"", string(p))
}

// true
func (phi PhiTrue) valid(K *Kripke, s int) bool {
	return true
}

func (phi PhiTrue) String() string {
	return "true"
}

// ¬ϕ

// internal type
func (phi _PhiNot) valid(K *Kripke, s int) bool {
	return !phi.phi1.valid(K, s)
}

func (phi _PhiNot) String() string {
	return fmt.Sprintf("¬(%v)", phi.phi1)
}

// public type, automatically collapses ¬¬ϕ into ϕ
func PhiNot(phi1 Phi) Phi {
	if phi, ok := phi1.(_PhiNot); ok {
		return phi.phi1
	}
	return _PhiNot{phi1}
}

// ϕ ∧ ϕ
func (phi PhiAnd) valid(K *Kripke, s int) bool {
	return phi.phi1.valid(K, s) && phi.phi2.valid(K, s)
}

func (phi PhiAnd) String() string {
	return fmt.Sprintf("(%v ∧ %v)", phi.phi1, phi.phi2)
}

// EX ϕ
func (phi PhiEX) valid(K *Kripke, s int) bool {
	if _, ok := K.cache[phi]; !ok {
		phi.marking(K)
	}
	_, ok := K.cache[phi][s]
	return ok
}

func (phi PhiEX) marking(K *Kripke) {
	K.cache[phi] = map[int]bool{}
	for s := range K.R {
		for _, t := range K.R[s] {
			if phi.phi1.valid(K, t) {
				K.cache[phi][s] = true
				break
			}
		}
	}
}

func (phi PhiEX) String() string {
	return fmt.Sprintf("EX(%v)", phi.phi1)
}

// E (ϕ U ϕ)
func (phi PhiEU) valid(K *Kripke, s int) bool {
	if _, ok := K.cache[phi]; !ok {
		phi.marking(K)
	}
	_, ok := K.cache[phi][s]
	return ok
}

func (phi PhiEU) marking(K *Kripke) {
	K.cache[phi] = map[int]bool{}
	open := []int{}
	closed := map[int]bool{}
	for s := range K.R {
		if phi.phi2.valid(K, s) {
			open = append(open, s)
		}
	}
	for len(open) > 0 {
		cur := open[0]
		open = open[1:]
		closed[cur] = true
		K.cache[phi][cur] = true
		for _, pred := range K.pred[cur] {
			if _, ok := closed[pred]; !ok && phi.phi1.valid(K, pred) {
				open = append(open, pred)
			}
		}
	}
}

func (phi PhiEU) String() string {
	return fmt.Sprintf("E(%v U %v)", phi.phi1, phi.phi2)
}

// A (ϕ U ϕ)
func (phi PhiAU) valid(K *Kripke, s int) bool {
	if _, ok := K.cache[phi]; !ok {
		phi.marking(K)
	}
	_, ok := K.cache[phi][s]
	return ok
}

func (phi PhiAU) marking(K *Kripke) {
	K.cache[phi] = map[int]bool{}
	nb := make([]int, len(K.R))
	open := []int{}
	for s := range K.R {
		nb[s] = len(K.R[s])
		if phi.phi2.valid(K, s) {
			open = append(open, s)
		}
	}
	for len(open) > 0 {
		cur := open[0]
		open = open[1:]
		K.cache[phi][cur] = true
		for _, pred := range K.pred[cur] {
			if !phi.phi2.valid(K, pred) && phi.phi1.valid(K, pred) {
				nb[pred] -= 1
				if nb[pred] == 0 {
					open = append(open, pred)
				}
			}
		}
	}
}

func (phi PhiAU) String() string {
	return fmt.Sprintf("A(%v U %v)", phi.phi1, phi.phi2)
}

// Non-basic operators
func PhiOr(phi1, phi2 Phi) Phi {
	return PhiNot(PhiAnd{PhiNot(phi1), PhiNot(phi2)})
}

func PhiAG(phi1 Phi) Phi {
	return PhiNot(PhiEU{PhiTrue{}, PhiNot(phi1)})
}

func PhiAF(phi1 Phi) Phi {
	return PhiAU{PhiTrue{}, phi1}
}

func PhiImpl(phi1, phi2 Phi) Phi {
	return PhiOr(PhiNot(phi1), phi2)
}

type Kripke struct {
	S0    []int
	R     [][]int
	L     map[int]map[AP]bool
	pred  [][]int
	cache map[Phi]map[int]bool
}

func MakeKripke(s0 []int, r [][]int, l map[int]map[AP]bool) *Kripke {
	K := Kripke{
		S0:    s0,
		R:     r,
		L:     l,
		cache: map[Phi]map[int]bool{},
	}
	K.pred = make([][]int, len(K.R))
	for s := range K.R {
		for _, t := range K.R[s] {
			K.pred[t] = append(K.pred[t], s)
		}
	}
	return &K
}

func main() {
	K := MakeKripke(
		[]int{0},
		[][]int{{1, 2}, {4}, {5, 0}, {2, 0, 3}, {1, 2}, {6}, {3}},
		map[int]map[AP]bool{
			1: {"start": true, "error": true}, 2: {"close": true}, 3: {"close": true, "heat": true},
			4: {"start": true, "error": true, "close": true}, 5: {"start": true, "close": true}, 6: {"start": true, "close": true, "heat": true}},
	)

	phi := PhiAG(PhiImpl(AP("start"), PhiAF(AP("heat"))))
	for i := 0; i < 10000; i++ {
		K.cache = map[Phi]map[int]bool{}
		phi.valid(K, K.S0[0])
	}
	log.Println(phi.valid(K, K.S0[0]))
}
