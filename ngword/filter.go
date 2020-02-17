package ngword

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"golang.org/x/text/unicode/norm"
	"strings"
)

type Filter interface {
	Do(frame dataframe.DataFrame) (dataframe.DataFrame, error)
}

type LocalAlignment struct {
	Ngwords dataframe.DataFrame
	Result  SmithWatermanResult
}

func NewLocalAlignment(df dataframe.DataFrame) *LocalAlignment {
	return &LocalAlignment{Ngwords: df}
}
func (la *LocalAlignment) Do(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	filtered := make([]string, df.Nrow())
	predict := make([]int, df.Nrow())
	df = df.Mutate(series.New(filtered, series.String, "filtered"))
	df = df.Mutate(series.New(predict, series.Int, "predict"))

	filter := func(s series.Series) series.Series {
		sentence := s.Elem(0).String()
		replaced, changed := la.Replace(sentence)
		//fmt.Println(replaced)
		s.Set(2, series.Strings([]string{replaced}))
		if changed {
			s.Set(3, series.Ints([]int{1}))
		}
		return s
	}
	df = df.Rapply(filter)

	return df, nil
}
func (la *LocalAlignment) Replace(sentence string) (string, bool) {
	changed := false
	origin := []rune(norm.NFKD.String(sentence))
	ngs := la.Ngwords.Maps()

	result := make([]rune, len(origin))
	copy(result, origin)
	for _, ng := range ngs {
		w := []rune(norm.NFKD.String(ng["word"].(string)))
		smithCh := SmithWaterman(origin, w, float32(ng["threshold"].(int))/100.0)
		for r := range smithCh {
			if r.LastNodes != nil {
				continue
			}
			changed = true
			for i := r.StartPos; i <= r.EndPos; i++ {
				result[i] = rune('*')
				//d := norm.NFC.NextBoundaryInString(string(origin[i:]), true)
				//if d /= 3; d == 0 {
				//	break
				//}
				//for j := i; j < i+d; j++ {
				//	result[j] = rune('*')
				//}
				//i += d
			}
		}
	}
	return norm.NFC.String(string(result)), changed
}

type PerfectMatch struct {
	Ngwords dataframe.DataFrame
}

func NewPerfectMatch(df dataframe.DataFrame) *PerfectMatch {
	df = df.Filter(
		dataframe.F{"lang", "==", "all"},
	).Filter(
		dataframe.F{"country", "==", "all"},
	).Filter(
		dataframe.F{"usage", "==", "all"},
	).Select([]string{"word"})
	return &PerfectMatch{Ngwords: df}
}
func (pm *PerfectMatch) Do(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	ngs := pm.Ngwords.Maps()
	filtered := make([]string, df.Nrow())
	predict := make([]int, df.Nrow())
	df = df.Mutate(series.New(filtered, series.String, "filtered"))
	df = df.Mutate(series.New(predict, series.Int, "predict"))

	filter := func(s series.Series) series.Series {
		sentence := s.Elem(0).String()
		for _, ng := range ngs {
			if strings.Index(sentence, ng["word"].(string)) >= 0 {
				replaced := strings.ReplaceAll(sentence, ng["word"].(string), "***")
				s.Set(2, series.Strings([]string{replaced}))
				s.Set(3, series.Ints([]int{1}))
			}
		}
		return s
	}
	df = df.Rapply(filter)

	return df, nil
}
