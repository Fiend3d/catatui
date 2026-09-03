// Port of examples/apps/volatility-surface/src/volatility.rs @ ratatui-v0.30.2

package main

import (
	"math"
	"math/rand/v2"
)

// The shape of the grid: 25 strikes across 70% to 130% of the spot price, and
// 20 expiries from a week out to two years.
const (
	strikeCount = 25
	expiryCount = 20
)

// volatilityEngine makes up an implied volatility surface and moves it along.
//
// None of this is a real model — it is the shape of one, enough to give the 3D
// renderer a surface that ripples the way a market's does. A real surface comes
// out of option prices; this one comes out of four terms and some noise.
type volatilityEngine struct {
	strikes       []float64 // moneyness, 0.7 to 1.3
	expirations   []float64 // time to expiry in years
	termStructure []float64 // the base volatility at each expiry
	surface       [][]float64
	baseVol       float64
	skew          float64
	time          float64
	rng           *rand.Rand
}

func newVolatilityEngine() *volatilityEngine {
	e := &volatilityEngine{
		baseVol: 20,
		skew:    0.3,
		rng:     rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
	for i := range strikeCount {
		e.strikes = append(e.strikes, 0.7+float64(i)*0.025)
	}
	for i := range expiryCount {
		e.expirations = append(e.expirations, 0.02+float64(i)*0.1)
	}

	// The term structure: volatility rises with time to expiry and flattens
	// out, which is what a calm market looks like.
	for _, t := range e.expirations {
		e.termStructure = append(e.termStructure, e.baseVol+5*(1-math.Exp(-t*2)))
	}

	e.regenerate()
	return e
}

// update advances the clock one tick and rebuilds the surface.
func (e *volatilityEngine) update() {
	e.time += 0.05
	e.regenerate()
}

// reset puts the clock back to zero.
func (e *volatilityEngine) reset() {
	e.time = 0
	e.regenerate()
}

// regenerate rebuilds the whole surface at the current time.
func (e *volatilityEngine) regenerate() {
	e.surface = make([][]float64, 0, len(e.expirations))

	for expiryIndex, expiry := range e.expirations {
		termVol := e.termStructure[expiryIndex]

		// Two slow waves over the whole surface: one that ripples along the
		// expiry axis, and one shock that lifts and drops everything at once.
		timeWave := math.Sin(e.time*0.5+float64(expiryIndex)*0.1) * 20
		volShock := math.Sin(e.time*0.3) * 1.5

		row := make([]float64, 0, len(e.strikes))
		for _, moneyness := range e.strikes {
			logMoneyness := math.Log(moneyness)
			sqrtExpiry := math.Sqrt(expiry)

			// The skew: puts cost more than calls at the same distance, so the
			// surface leans. Negative because the skew is on the put side.
			skew := -e.skew * logMoneyness * 100 / sqrtExpiry

			// The smile: options far from the money cost more either way, so
			// the lean has a curve in it.
			smile := 5 * logMoneyness * logMoneyness / sqrtExpiry

			// The wings, which lift the far ends further still.
			wing := 0.0
			if moneyness < 0.95 || moneyness > 1.05 {
				wing = (math.Abs(moneyness-1) - 0.05) * 20
			}

			// Volatility clusters: a busy patch stays busy for a while.
			cluster := math.Abs(math.Sin(e.time*2+moneyness*10)) * 1.5

			noise := (e.rng.Float64() - 0.5) * 0.5

			iv := termVol + skew + smile + wing + timeWave + volShock + cluster + noise
			row = append(row, min(max(iv, 5), 80))
		}
		e.surface = append(e.surface, row)
	}
}

// getSurface is the grid as [expiry][strike].
func (e *volatilityEngine) getSurface() [][]float64 { return e.surface }
