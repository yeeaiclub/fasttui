package components

import "math"

// glowNoise is a compact value-noise sampler.
// sample(x/20, y/20 + t*speed) mapped to [0,1].
type glowNoise struct {
	seed int
}

func newGlowNoise(seed int) glowNoise {
	if seed == 0 {
		seed = 1337
	}
	return glowNoise{seed: seed}
}

func (g glowNoise) sample(x, y, t, speed float64) float64 {
	// noise2D(x/20, y/20 + t*speed) → [0,1],
	// with a light second octave for smoother swirl.
	const scale = 20.0
	nx := x / scale
	ny := y/scale + t*speed
	n := valueNoise2D(nx, ny, g.seed)*0.7 + valueNoise2D(nx*2.1+3.1, ny*2.1, g.seed+17)*0.3
	return (n + 1) * 0.5
}

func valueNoise2D(x, y float64, seed int) float64 {
	x0 := math.Floor(x)
	y0 := math.Floor(y)
	fx := x - x0
	fy := y - y0
	// Smoothstep
	ux := fx * fx * (3 - 2*fx)
	uy := fy * fy * (3 - 2*fy)

	n00 := hash2(int(x0), int(y0), seed)
	n10 := hash2(int(x0)+1, int(y0), seed)
	n01 := hash2(int(x0), int(y0)+1, seed)
	n11 := hash2(int(x0)+1, int(y0)+1, seed)

	nx0 := n00*(1-ux) + n10*ux
	nx1 := n01*(1-ux) + n11*ux
	return nx0*(1-uy) + nx1*uy
}

func hash2(x, y, seed int) float64 {
	n := x*374761393 + y*668265263 + seed*1442695041
	n = (n ^ (n >> 13)) * 1274126177
	n = n ^ (n >> 16)
	return float64(n&0xffff)/32767.5 - 1 // [-1, 1]
}
