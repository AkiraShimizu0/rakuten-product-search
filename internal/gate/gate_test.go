package gate

import (
	"math"
	"testing"
)

func TestRecallMaxRejectAndBoundary(t *testing.T) {
	rows := []Row{}
	for i := 0; i < 20; i++ {
		rows = append(rows, Row{Raw: float64(i) / 20, AI: 4})
	}
	rows = append(rows, Row{Raw: -1, AI: 2})
	p, curve, fn, e := Calibrate(rows, .95)
	if e != nil || p.Threshold != .05 || p.Recall != .95 || p.Reject != 2 || len(fn) != 1 {
		t.Fatal(p, fn, e)
	}
	for _, x := range curve {
		if x.Recall >= .95 && x.Reject > p.Reject {
			t.Fatal("not maximal")
		}
	}
	c := Config{Threshold: .8133}
	if !c.Pass(.8132999999999999) || c.Pass(.813299) {
		t.Fatal("boundary tolerance")
	}
}
func TestCalibrationNoPositive(t *testing.T) {
	if _, _, _, e := Calibrate([]Row{{Raw: 1, AI: 3}}, .95); e == nil {
		t.Fatal("no positives")
	}
	if _, _, _, e := Calibrate([]Row{{Raw: 1, AI: 4}}, math.NaN()); e == nil {
		t.Fatal("NaN target")
	}
}
