package ngword

import "golang.org/x/text/unicode/norm"

const (
	SCORE_MATCH    = 5
	SCORE_SIMILAR  = 4
	SCORE_SPACE    = 0
	SCORE_MISMATCH = -2

	NEXT_END = 0
	NEXT_IJ  = 1
	NEXT_I   = 2
	NEXT_J   = 3
)

type SmithWatermanResult struct {
	MatchWord                           string
	CompleteAgreement, AppliedAgreement int
	SimilarScore                        float32
	StartPos, EndPos                    int
}

type SmithWatermanEnd struct {
	MatchWord         string
	LastNodes         []Node
	CompleteAgreement int
	MaxAgreement      int
	ThreshAgreement   int
}

type BySimilar []SmithWatermanEnd

func (b BySimilar) Len() int {
	return len(b)
}
func (b BySimilar) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b BySimilar) Less(i, j int) bool {
	f1 := float64(b[i].MaxAgreement) / float64(b[i].CompleteAgreement)
	f2 := float64(b[j].MaxAgreement) / float64(b[j].CompleteAgreement)
	return f1 > f2
}

type Node struct {
	Score int
	Next  int
}

var MatchTable = map[rune]int{
	0x1100: 1,   // ㄱ
	0x1101: 1,   // ㄲ
	0x110f: 1,   // ㅋ
	0x1109: 4,   // ㅅ 초성
	0x110a: 4,   // ㅆ 초성
	0x1107: 5,   // ㅂ 초성
	0x1108: 5,   // ㅃ 초성
	0x1111: 5,   // ㅍ 초성
	0x1103: 6,   // ㄷ 초성
	0x1104: 6,   // ㄸ 초성
	0x1110: 6,   // ㅌ 초성
	0x110c: 7,   // ㅈ 초성
	0x110d: 7,   // ㅉ 초성
	0x110e: 7,   // ㅊ 초성
	0x1161: 100, // ㅏ 중성
	0x1163: 100, // ㅑ 중성
	0x1165: 101, // ㅓ 중성
	0x1167: 101, // ㅕ 중성
	0x1169: 102, // ㅗ 중성
	0x116d: 102, // ㅛ 중성
	0x116e: 103, // ㅜ 중성
	0x1172: 103, // ㅠ 중성
	0x1171: 104, // ㅟ 중성
	0x1174: 104, // ㅢ 중성
	0x1175: 104, // ㅣ 중성
}

func Match(a, b rune) int {
	if a == b {
		return SCORE_MATCH
	} else if a == ' ' || b == ' ' {
		return SCORE_SPACE
	}

	g1, ok1 := MatchTable[a]
	g2, ok2 := MatchTable[b]
	if ok1 && ok2 && g1 == g2 {
		return SCORE_SIMILAR
	}

	return SCORE_MISMATCH
}

func History(stMatrix [][]Node, lenWord, endPos int) int {
	s := endPos
	t := lenWord
	for p := stMatrix[t][s]; p.Next != NEXT_END; p = stMatrix[t][s] {
		if p.Next == NEXT_IJ {
			t, s = t-1, s-1
		} else if p.Next == NEXT_I {
			t = t - 1
		} else if p.Next == NEXT_J {
			s = s - 1
		}
	}

	return s
}

//func SmithWatermanScore(sentence, word []rune) SmithWatermanResult {
//	lenStc := len(sentence)
//	lenWord := len(word)
//
//	stMatrix := _p.Get().([][]int)
//	for t := 1; t <= lenWord; t++ {
//		stMatrix[t][0] = -t
//	}
//
//	for t := 1; t <= lenWord; t++ {
//		for s := 1; s <= lenStc; s++ {
//			ijscore := stMatrix[t-1][s-1] + Match(sentence[s-1], word[t-1])
//			iscore := stMatrix[t-1][s] + Match(rune(0), word[t-1])
//			jscore := stMatrix[t][s-1] + Match(sentence[s-1], rune(0))
//
//			if ijscore >= iscore && ijscore >= jscore {
//				stMatrix[t][s] = ijscore
//			} else if iscore >= ijscore && iscore >= jscore {
//				stMatrix[t][s] = iscore
//			} else {
//				stMatrix[t][s] = jscore
//			}
//			if stMatrix[t][s] < 0 {
//				stMatrix[t][s] = 0
//			}
//		}
//	}
//
//	completeAgreement := lenWord * SCORE_MATCH
//	maxAgreement := 0
//	startPos := 0
//	for i, s := range stMatrix[lenWord] {
//		if s > maxAgreement {
//			maxAgreement = s
//			startPos = i
//		}
//	}
//
//	for t := 0; t <= lenWord; t++ {
//		for s := 0; s <= lenStc; s++ {
//			stMatrix[t][s] = 0
//		}
//	}
//	_p.Put(stMatrix)
//	return SmithWatermanResult{
//		CompleteAgreement: completeAgreement,
//		AppliedAgreement:  maxAgreement,
//		SimilarScore:      float32(maxAgreement) / float32(completeAgreement),
//		StartPos:          startPos,
//	}
//}

func SmithWaterman(sentence, word []rune, thresh float32) (<-chan SmithWatermanResult, <-chan SmithWatermanEnd) {
	smithCh := make(chan SmithWatermanResult, 10)
	endCh := make(chan SmithWatermanEnd)

	go func() {
		lenStc := len(sentence)
		lenWord := len(word)

		//stMatrix := _p.Get().([][]Node)
		stMatrix := make([][]Node, lenWord+1)
		for i := range stMatrix {
			stMatrix[i] = make([]Node, lenStc+1)
		}

		for t := 1; t <= lenWord; t++ {
			for s := 1; s <= lenStc; s++ {
				ijscore := stMatrix[t-1][s-1].Score + Match(sentence[s-1], word[t-1])
				iscore := stMatrix[t-1][s].Score + Match(rune(0), word[t-1])
				jscore := stMatrix[t][s-1].Score + Match(sentence[s-1], rune(0))

				if ijscore >= iscore && ijscore >= jscore {
					stMatrix[t][s].Score = ijscore
					stMatrix[t][s].Next = NEXT_IJ
				} else if iscore >= jscore {
					stMatrix[t][s].Score = iscore
					stMatrix[t][s].Next = NEXT_I
				} else {
					stMatrix[t][s].Score = jscore
					stMatrix[t][s].Next = NEXT_J
				}
				/*
					if stMatrix[t][s].Score < 0 {
						stMatrix[t][s].Score = 0
					}
				*/
			}
		}

		completeAgreement := lenWord * SCORE_MATCH
		threshAgreement := float32(completeAgreement) * thresh
		maxAgreement := -987654321
		for i := len(stMatrix[lenWord]) - 1; i >= 0; i-- {
			v := stMatrix[lenWord][i]
			if v.Score > maxAgreement {
				maxAgreement = v.Score
			}
		}
		for i := len(stMatrix[lenWord]) - 1; i >= 0; i-- {
			v := stMatrix[lenWord][i]
			if float32(v.Score) > threshAgreement {
				s := History(stMatrix, lenWord, i)
				smithCh <- SmithWatermanResult{
					MatchWord:         string(word),
					CompleteAgreement: completeAgreement,
					AppliedAgreement:  v.Score,
					SimilarScore:      float32(v.Score) / float32(completeAgreement),
					StartPos:          s,
					EndPos:            i - 1,
					//EndPos:            i,
				}
				i = s
			}
		}

		//for t := 0; t <= lenWord; t++ {
		//	for s := 0; s <= lenStc; s++ {
		//		stMatrix[t][s].Score = 0
		//		stMatrix[t][s].Next = NEXT_END
		//	}
		//}
		//_p.Put(stMatrix)
		endCh <- SmithWatermanEnd{
			MatchWord:         norm.NFC.String(string(word)),
			LastNodes:         stMatrix[lenWord],
			CompleteAgreement: completeAgreement,
			MaxAgreement:      maxAgreement,
			ThreshAgreement:   int(float32(completeAgreement) * thresh),
		}
		close(smithCh)
		close(endCh)
	}()
	return smithCh, endCh
}
