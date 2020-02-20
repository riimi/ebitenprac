package ngword

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"golang.org/x/text/unicode/norm"
	"sort"
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
	filtered := make([]string, 0, df.Nrow())
	predict := make([]int, 0, df.Nrow())

	filter := func(s series.Series) series.Series {
		sentence := s.Elem(0).String()
		replaced, changed := la.Replace(sentence)
		//fmt.Println(replaced)
		filtered = append(filtered, replaced)
		if changed {
			predict = append(predict, 1)
		} else {
			predict = append(predict, 0)
		}
		return s
	}
	df = df.Rapply(filter)
	df = df.Mutate(series.New(filtered, series.String, "filtered"))
	//df = df.Mutate(series.New(predict, series.Int, "predict"))

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
		smithCh, endCh := SmithWaterman(origin, w, float32(ng["threshold"].(int))/100.0)

	loop:
		for {
			select {
			case r := <-smithCh:
				changed = true
				for i := r.StartPos; i <= r.EndPos; i++ {
					result[i] = rune('*')
				}
			case <-endCh:
				break loop
			}
		}

	}
	return norm.NFC.String(string(result)), changed
}

type LocalAlignmentDebug struct {
	Ngwords dataframe.DataFrame
	End     []SmithWatermanEnd
}

func NewLocalAlignmentDebug(df dataframe.DataFrame) *LocalAlignmentDebug {
	return &LocalAlignmentDebug{
		Ngwords: df,
		End:     make([]SmithWatermanEnd, 0),
	}
}
func (la *LocalAlignmentDebug) Replace(sentence string) (string, bool) {
	changed := false
	origin := []rune(norm.NFKD.String(sentence))
	ngs := la.Ngwords.Maps()
	la.End = make([]SmithWatermanEnd, 0, len(ngs))

	result := make([]rune, len(origin))
	copy(result, origin)
	for _, ng := range ngs {
		w := []rune(norm.NFKD.String(ng["word"].(string)))
		smithCh, endCh := SmithWaterman(origin, w, float32(ng["threshold"].(int))/100.0)

	loop:
		for {
			select {
			case r := <-smithCh:
				changed = true
				for i := r.StartPos; i <= r.EndPos; i++ {
					result[i] = rune('*')
				}
			case e := <-endCh:
				la.End = append(la.End, e)
				break loop
			}
		}
	}

	sort.Sort(BySimilar(la.End))
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
