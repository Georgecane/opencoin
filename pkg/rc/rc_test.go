package rc

import "testing"

func TestEffectiveTimeClamp(t *testing.T) {
	last := []int64{100, 101, 102, 103, 104}
	eff := EffectiveTime(200, last, 5)
	if eff != 107 { // median 102, max 107
		t.Fatalf("expected 107 got %d", eff)
	}
	eff = EffectiveTime(90, last, 5)
	if eff != 97 { // median 102, min 97
		t.Fatalf("expected 97 got %d", eff)
	}
}

func TestRegen(t *testing.T) {
	p := Params{Alpha: 10, Beta: 2}
	rc, last := p.Regen(0, 5, 0, 3)
	if last != 3 {
		t.Fatalf("expected last 3 got %d", last)
	}
	// regen = beta*stake*dt = 2*5*3 = 30, rc_max = 50
	if rc != 30 {
		t.Fatalf("expected rc 30 got %d", rc)
	}
}
