package ctlmc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type phi1 struct {
	phi1 Phi
}

type phi2 struct {
	phi1 Phi
	phi2 Phi
}

type (
	AP       string
	_PhiNot  phi1
	PhiAnd   phi2
	PhiEX    phi1
	PhiEU    phi2
	PhiAU    phi2
	_PhiTrue struct{}
)

// Phi represents a CTL formula which can be checked on a Kripke structure
type Phi interface {
	Check(*Kripke, int) bool
}

// p
func (p AP) Check(K *Kripke, s int) bool {
	_, ok := K.L[s][p]
	return ok
}

func (p AP) String() string {
	return fmt.Sprintf("\"%v\"", string(p))
}

// true
func (phi _PhiTrue) Check(K *Kripke, s int) bool {
	return true
}

func (phi _PhiTrue) String() string {
	return "true"
}

var PhiTrue Phi = _PhiTrue{}

// false
var PhiFalse Phi = _PhiNot{PhiTrue}

// ¬ϕ

// internal type
func (phi _PhiNot) Check(K *Kripke, s int) bool {
	return !phi.phi1.Check(K, s)
}

func (phi _PhiNot) String() string {
	return fmt.Sprintf("¬%v", phi.phi1)
}

// public type, automatically collapses ¬¬ϕ into ϕ
func PhiNot(phi1 Phi) Phi {
	if phi, ok := phi1.(_PhiNot); ok {
		return phi.phi1
	}
	return _PhiNot{phi1}
}

// ϕ ∧ ϕ
func (phi PhiAnd) Check(K *Kripke, s int) bool {
	return phi.phi1.Check(K, s) && phi.phi2.Check(K, s)
}

func (phi PhiAnd) String() string {
	return fmt.Sprintf("(%v ∧ %v)", phi.phi1, phi.phi2)
}

// EX ϕ
func (phi PhiEX) Check(K *Kripke, s int) bool {
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
			if phi.phi1.Check(K, t) {
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
func (phi PhiEU) Check(K *Kripke, s int) bool {
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
		if phi.phi2.Check(K, s) {
			seen[s] = true
			open = append(open, s)
		}
	}
	for len(open) > 0 {
		cur := open[0]
		open = open[1:]
		K.cache[phi][cur] = true
		for _, pred := range K.pred[cur] {
			if _, ok := seen[pred]; !ok && phi.phi1.Check(K, pred) {
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
func (phi PhiAU) Check(K *Kripke, s int) bool {
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
		if phi.phi2.Check(K, s) {
			open = append(open, s)
		}
	}
	for len(open) > 0 {
		cur := open[0]
		open = open[1:]
		K.cache[phi][cur] = true
		for _, pred := range K.pred[cur] {
			if !phi.phi2.Check(K, pred) && phi.phi1.Check(K, pred) {
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
	return _PhiNot{PhiAnd{PhiNot(phi1), PhiNot(phi2)}}
}

func PhiEF(phi1 Phi) Phi {
	return PhiEU{PhiTrue, phi1}
}

func PhiEG(phi1 Phi) Phi {
	return _PhiNot{PhiAU{PhiTrue, PhiNot(phi1)}}
}

// E(pWq) = E(pUq or Gp) = E(pUq) or EGp = E(pUq) or -AF-p = E(pUq) or -A(tU-p)
func PhiEW(phi1, phi2 Phi) Phi {
	return PhiOr(PhiEU{phi1, phi2}, _PhiNot{PhiAU{PhiTrue, PhiNot(phi1)}})
}

// E(pRq) = E-(-pU-q) = -A(-pU-q)
func PhiER(phi1, phi2 Phi) Phi {
	return _PhiNot{PhiAU{PhiNot(phi1), PhiNot(phi2)}}
}

// E(pSq) = E(qU(p a q))
func PhiES(phi1, phi2 Phi) Phi {
	return PhiEU{phi2, PhiAnd{phi1, phi2}}
}

func PhiAX(phi1 Phi) Phi {
	return _PhiNot{PhiEX{PhiNot(phi1)}}
}

func PhiAF(phi1 Phi) Phi {
	return PhiAU{PhiTrue, phi1}
}

func PhiAG(phi1 Phi) Phi {
	return _PhiNot{PhiEU{PhiTrue, PhiNot(phi1)}}
}

// A(pRq) = A-(-pU-q) = -E(-pU-q)
func PhiAR(phi1, phi2 Phi) Phi {
	return _PhiNot{PhiEU{PhiNot(phi1), PhiNot(phi2)}}
}

// A(pSq) = A(pRq a Fp) = A(pRq) a AFp = -E(-pU-q) a A(tUp)
func PhiAS(phi1, phi2 Phi) Phi {
	return PhiAnd{_PhiNot{PhiEU{PhiNot(phi1), PhiNot(phi2)}}, PhiAU{PhiTrue, phi1}}
}

// A(pWq) = A-(-pS-q) = -E(-qU(-p a -q))
func PhiAW(phi1, phi2 Phi) Phi {
	nq := PhiNot(phi2)
	return _PhiNot{PhiEU{nq, PhiAnd{PhiNot(phi1), nq}}}
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

func (K *Kripke) Satisfies(phi Phi) bool {
	for _, i := range K.S0 {
		if !phi.Check(K, i) {
			return false
		}
	}
	return true
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

func parseCTL(s string, i int) (Phi, int, error) {
	if i >= len(s) {
		return nil, i, fmt.Errorf("unexpected end of input at %v", i)
	}

	if s[i:i+4] == "true" {
		return PhiTrue, i + 4, nil
	}

	if s[i:i+5] == "false" {
		return PhiFalse, i + 5, nil
	}

	// p
	if s[i] == '"' {
		if end := strings.Index(s[i+1:], "\""); end < 1 {
			return nil, i, fmt.Errorf("empty or unterminated AP at %v", i)
		} else {
			return AP(s[i+1 : i+1+end]), i + end + 2, nil
		}
	}

	// ¬ϕ
	if s[i] == '-' {
		if phi_, next, err := parseCTL(s, i+1); err != nil {
			return nil, next, err
		} else {
			return PhiNot(phi_), next, nil
		}
	}

	// (ϕ)
	if s[i] == '(' {
		if phi1, next, err := parseCTL(s, i+1); err != nil {
			return nil, next, err
		} else {
			if next >= len(s) {
				return nil, i, fmt.Errorf("unterminated parantheses block starting at %v", i)
			} else if s[next] == ')' {
				return phi1, next + 1, nil
			} else if s[next] == 'a' || s[next] == 'o' {
				if phi2, end, err := parseCTL(s, next+1); err != nil {
					return nil, end, err
				} else {
					if end >= len(s) || s[end] != ')' {
						return nil, i, fmt.Errorf("unterminated con-/disjunction block starting at %v", i)
					}
					if s[next] == 'a' {
						return PhiAnd{phi1, phi2}, end + 1, nil
					} else { // o
						return PhiOr(phi1, phi2), end + 1, nil
					}
				}
			} else {
				return nil, i, fmt.Errorf("malformed parantheses block starting at %v", i)
			}
		}
	}

	switch s[i : i+2] {
	case "EX", "EF", "EG", "AX", "AF", "AG":
		if phi1, next, err := parseCTL(s, i+2); err != nil {
			return nil, next, err
		} else {
			switch s[i : i+2] {
			case "EX":
				return PhiEX{phi1}, next, nil
			case "EF":
				return PhiEF(phi1), next, nil
			case "EG":
				return PhiEG(phi1), next, nil
			case "AX":
				return PhiAX(phi1), next, nil
			case "AF":
				return PhiAF(phi1), next, nil
			case "AG":
				return PhiAG(phi1), next, nil
			}
		}
	}

	if s[i:i+2] == "E(" || s[i:i+2] == "A(" {
		if phi1, next, err := parseCTL(s, i+2); err != nil {
			return nil, next, err
		} else {
			if next >= len(s) {
				return nil, i, fmt.Errorf("incomplete quantor-modality block starting at %v", i)
			}
			switch s[next] {
			case 'U', 'W', 'R', 'S':
				if phi2, end, err := parseCTL(s, next+1); err != nil {
					return nil, end, err
				} else {
					if end >= len(s) || s[end] != ')' {
						return nil, i, fmt.Errorf("unterminated quantor-modality block starting at %v", i)
					}
					if s[i] == 'E' {
						switch s[next] {
						case 'U':
							return PhiEU{phi1, phi2}, end + 1, nil
						case 'W':
							return PhiEW(phi1, phi2), end + 1, nil
						case 'R':
							return PhiER(phi1, phi2), end + 1, nil
						case 'S':
							return PhiES(phi1, phi2), end + 1, nil
						}
					} else { // A
						switch s[next] {
						case 'U':
							return PhiAU{phi1, phi2}, end + 1, nil
						case 'W':
							return PhiAW(phi1, phi2), end + 1, nil
						case 'R':
							return PhiAR(phi1, phi2), end + 1, nil
						case 'S':
							return PhiAS(phi1, phi2), end + 1, nil
						}
					}
				}
			}
			return nil, i, fmt.Errorf("malformed quantor-modality block starting at %v", i)
		}
	}
	return nil, 0, fmt.Errorf("unrecognized syntax at %v", i)
}

func ParseCTL(s string) (Phi, error) {
	s = regexp.MustCompile(`\s`).ReplaceAllLiteralString(s, "")
	if phi, end, err := parseCTL(s, 0); err != nil {
		return nil, err
	} else if end != len(s) {
		return nil, fmt.Errorf("unexpected input after pos. %v", end)
	} else {
		return phi, nil
	}
}

func ReadKripke(path string) *Kripke {
	type KripkeJson struct {
		S0 []int
		R  map[int][]int
		L  map[int][]string
	}
	var msg KripkeJson

	infile, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to parse Kripke structure from json: %v", err)
	}
	defer infile.Close()

	dec := json.NewDecoder(infile)
	if err := dec.Decode(&msg); err != nil {
		log.Fatalf("failed to parse Kripke structure from json: %v", err)
	}

	R := make([][]int, len(msg.R))
	L := make(map[int]map[AP]bool)
	for k, v := range msg.R {
		R[k] = v
	}
	for k, vs := range msg.L {
		L[k] = make(map[AP]bool)
		for _, v := range vs {
			L[k][AP(v)] = true
		}
	}
	return MakeKripke(msg.S0, R, L)
}
