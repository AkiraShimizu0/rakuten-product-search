package cli

import (
	"flag"
	"jev-money-engine/internal/filter"
)

type Options struct {
	DB     string
	Sample int
	Filter filter.Config
}

func Bind(f *flag.FlagSet) *Options {
	o := &Options{Filter: filter.Default()}
	f.StringVar(&o.DB, "db", "data/money.db", "SQLite database path")
	f.IntVar(&o.Sample, "sample", 0, "random eligible products to display")
	f.Int64Var(&o.Filter.MinPrice, "min-price", o.Filter.MinPrice, "minimum price (JPY, inclusive)")
	f.Int64Var(&o.Filter.MaxPrice, "max-price", o.Filter.MaxPrice, "maximum price (JPY, inclusive)")
	f.Int64Var(&o.Filter.MinReviews, "min-reviews", o.Filter.MinReviews, "minimum review count")
	f.IntVar(&o.Filter.MinCaptionRunes, "min-caption", o.Filter.MinCaptionRunes, "minimum description length (Unicode characters)")
	return o
}
