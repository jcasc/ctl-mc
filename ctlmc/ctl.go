package ctlmc

import (
	"fmt"
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
	Valid(*Kripke, int) bool
}

// p
func (p AP) Valid(K *Kripke, s int) bool {
	_, ok := K.L[s][p]
	return ok
}

func (p AP) String() string {
	return fmt.Sprintf("\"%v\"", string(p))
}

// true
func (phi PhiTrue) Valid(K *Kripke, s int) bool {
	return true
}

func (phi PhiTrue) String() string {
	return "true"
}

// ¬ϕ

// internal type
func (phi _PhiNot) Valid(K *Kripke, s int) bool {
	return !phi.phi1.Valid(K, s)
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
func (phi PhiAnd) Valid(K *Kripke, s int) bool {
	return phi.phi1.Valid(K, s) && phi.phi2.Valid(K, s)
}

func (phi PhiAnd) String() string {
	return fmt.Sprintf("(%v ∧ %v)", phi.phi1, phi.phi2)
}

// EX ϕ
func (phi PhiEX) Valid(K *Kripke, s int) bool {
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
			if phi.phi1.Valid(K, t) {
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
func (phi PhiEU) Valid(K *Kripke, s int) bool {
	if _, ok := K.cache[phi]; !ok {
		phi.marking(K)
	}
	_, ok := K.cache[phi][s]
	return ok
}

func (phi PhiEU) marking(K *Kripke) {
	K.cache[phi] = map[int]bool{}
	open := []int{}
	seen := map[int]bool{}
	for s := range K.R {
		if phi.phi2.Valid(K, s) {
			seen[s] = true
			open = append(open, s)
		}
	}
	for len(open) > 0 {
		cur := open[0]
		open = open[1:]
		K.cache[phi][cur] = true
		for _, pred := range K.pred[cur] {
			if _, ok := seen[pred]; !ok && phi.phi1.Valid(K, pred) {
				seen[pred] = true
				open = append(open, pred)
			}
		}
	}
}

func (phi PhiEU) String() string {
	return fmt.Sprintf("E(%v U %v)", phi.phi1, phi.phi2)
}

// A (ϕ U ϕ)
func (phi PhiAU) Valid(K *Kripke, s int) bool {
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
		if phi.phi2.Valid(K, s) {
			open = append(open, s)
		}
	}
	for len(open) > 0 {
		cur := open[0]
		open = open[1:]
		K.cache[phi][cur] = true
		for _, pred := range K.pred[cur] {
			if !phi.phi2.Valid(K, pred) && phi.phi1.Valid(K, pred) {
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

func (K *Kripke) Clear() {
	K.cache = map[Phi]map[int]bool{}
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
