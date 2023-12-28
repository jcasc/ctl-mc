# A CTL (Computation Tree Logic) Model Checker

## Kripke Structures

The following definition of Krikpe structure is used:
A Krikpe structure K is a tuple (S, S_0, R, L) where
- S is the set of states
- S_0 is the set of initial states
- R is the left-total transition relation
- L is the labeling function S->2^AP where AP is the set of Atomic Propositions

Krikpe structures are read from a json file with 3 top-level keys `S0`, `R`, and `L` where
- R is a json object mapping each state (as string) to an array of successor states in the total (!) transition relation
- R implicitly defines S as the set of R's keys
- S0 is the array of initial states (as strings)
- L is a json object mapping each state (as strings) to an array Atomic Propositions which are true for the given state

The set AP is effectively the set of all strings not containing whitespace.

See `kripke.json` for an example of the format.

## CTL Formulas

Formulas can be parsed using the provided function from the following syntax:

```
ϕ ::= true | false | p | -ϕ | (ϕ * ϕ) | (ϕ + ϕ) | Eψ | Aψ
ψ ::= Xϕ | Fϕ | Gϕ | (ϕ U ϕ) | (ϕ W ϕ) | (ϕ R ϕ) | (ϕ S ϕ)
```

The common CTL semantics apply, where:
```
- = NOT
* = AND
+ = OR
U = Until
W = Weak Until
R = Release
S = Strong Release
```


where `p` is a `"`-delimited string. The parantheses delimiting the subformulas of operators taking two operands are mandatory!
The trivial sub-formulas `true` and `false` are *not* quote-delimited.
Whitespace is ignored entirely, even within AP-strings.

Internally, formulas as represented as a tree structure of subformulas of the equivalent formula of "basic" CTL logic in which quantor-modalities are limited to `EX`, `EU` and `AU`.

## Example

```go
K, err := ctlmc.ReadKripke("kripke.json")
if err != nil {
  log.Fatal(err)
}
phi, err := ctlmc.ParseCTL(`AG(-"start"+AF"heat")`)
if err != nil {
  log.Fatalf("failed parsing CTL formula: %v", err)
}
log.Println(phi) // ¬E(true U ("start" ∧ ¬A(true U "heat")))
log.Println(phi.Check(K, K.S0[0])) // Check Formula phi on K's first initial state.
log.Println(K.Satisfies(phi)) // Check whether K satisfies phi, i.e. whether phi holds for every initial state of K.
```



