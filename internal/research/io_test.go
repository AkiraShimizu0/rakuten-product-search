package research

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicCheckpointAndReadOnly(t *testing.T) {
	p := filepath.Join(t.TempDir(), "checkpoint.json")
	if e := JSON(p, map[string]int{"n": 1}); e != nil {
		t.Fatal(e)
	}
	if e := JSON(p, map[string]int{"n": 2}); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(p)
	if e != nil || string(b) != "{\n  \"n\": 2\n}" {
		t.Fatal(string(b), e)
	}
	dbpath := filepath.Join(t.TempDir(), "immutable.db")
	d, e := sql.Open("sqlite", dbpath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = d.Exec("CREATE TABLE frozen(id INTEGER)"); e != nil {
		t.Fatal(e)
	}
	d.Close()
	ro, e := ReadOnly(dbpath)
	if e != nil {
		t.Fatal(e)
	}
	defer ro.Close()
	if _, e = ro.Exec("INSERT INTO frozen VALUES(1)"); e == nil {
		t.Fatal("read only violation")
	}
}
func TestStatistics(t *testing.T) {
	if Quantile([]float64{3, 1, 4, 2}, .5) != 2.5 {
		t.Fatal("median")
	}
	n, largest, hhi := Counts([]string{"a", "a", "b", "c"})
	if n != 3 || largest != .5 || hhi != .375 {
		t.Fatal(n, largest, hhi)
	}
}
