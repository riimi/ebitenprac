package ngword

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"golang.org/x/text/unicode/norm"
	"sort"
	"strings"
	"sync"
)

type Filter interface {
	Do(frame dataframe.DataFrame) (dataframe.DataFrame, error)
}

type LocalAlignment struct {
	Ngwords dataframe.DataFrame
	Result  SmithWatermanResult
}

func NewLocalAlignment(df dataframe.DataFrame) *LocalAlignment {
	normalize := func(s series.Series) series.Series {
		words := s.Records()
		nor := make([]string, len(words))
		for i, w := range words {
			nor[i] = norm.NFKD.String(w)
		}
		return series.Strings(nor)
	}
	df = df.Select([]string{"word"}).Capply(normalize)
	return &LocalAlignment{Ngwords: df}
}
func (la *LocalAlignment) Do(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	filtered := make([]string, df.Nrow())
	predict := make([]int, df.Nrow())

	filter := func(s series.Series) series.Series {
		stcs := s.Records()
		for i, s := range stcs {
			replaced, changed := la.Replace(s)
			filtered[i] = replaced
			if changed {
				predict[i] = 1
			} else {
				predict[i] = 0
			}
		}
		return s
	}
	df.Capply(filter)
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
		//w := []rune(norm.NFKD.String(ng["word"].(string)))
		w := []rune(ng["word"].(string))
		smithCh, endCh := SmithWaterman(origin, w, -0.005*float32(len(w))+0.95)

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

type LocalAlignmentTrie struct {
	trie Trie
}

func NewLocalAlignmentTrie(df dataframe.DataFrame) *LocalAlignmentTrie {
	trie := NewTrie()
	normalize := func(s series.Series) series.Series {
		words := s.Records()
		for _, w := range words {
			trie.Append(norm.NFKD.String(w))
		}
		return s
	}
	df.Select([]string{"word"}).Capply(normalize)
	return &LocalAlignmentTrie{trie: trie}
}
func (la *LocalAlignmentTrie) Do(df dataframe.DataFrame) (dataframe.DataFrame, error) {
	filtered := make([]string, df.Nrow())
	predict := make([]int, df.Nrow())

	filter := func(s series.Series) series.Series {
		token := make(chan struct{}, 8)
		stcs := s.Records()
		wg := &sync.WaitGroup{}
		wg.Add(len(stcs))

		for i, s := range stcs {
			token <- struct{}{}
			go func(idx int, str string) {
				replaced, changed := la.Replace(str)
				filtered[idx] = replaced
				if changed {
					predict[idx] = 1
				} else {
					predict[idx] = 0
				}
				<-token
				wg.Done()
			}(i, s)
		}

		wg.Wait()
		//for i, s := range stcs {
		//	replaced, changed := la.Replace(s)
		//	filtered[i] = replaced
		//	if changed {
		//		predict[i] = 1
		//	} else {
		//		predict[i] = 0
		//	}
		//}
		return s
	}
	df.Capply(filter)
	df = df.Mutate(series.New(filtered, series.String, "filtered"))
	//df = df.Mutate(series.New(predict, series.Int, "predict"))

	return df, nil
}
func (la *LocalAlignmentTrie) Replace(sentence string) (string, bool) {
	changed := false
	origin := []rune(norm.NFKD.String(sentence))

	result := make([]rune, len(origin))
	copy(result, origin)
	smithCh := SmithWatermanTrie(origin, la.trie, 0.9)
	for r := range smithCh {
		changed = true
		for i := r.StartPos; i <= r.EndPos; i++ {
			result[i] = rune('*')
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
