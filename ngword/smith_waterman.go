package ngword

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
	LastNodes                           []Node
	StartPos, EndPos                    int
}

type Node struct {
	Score int
	Next  int
}

var MatchTable = map[rune]int{
	0x1100: 1, // ㄱ
	0x1101: 1, // ㄲ
	0x110f: 1, // ㅋ
	0x1109: 4, // ㅅ 초성
	0x110a: 4, // ㅆ 초성
	0x1107: 5, // ㅂ 초성
	0x1111: 5, // ㅍ 초성
	0x11ac: 6, // ㅓ 중성
	0x11ae: 6, // ㅕ 중성
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

func SmithWaterman(sentence, word []rune, thresh float32) <-chan SmithWatermanResult {
	smithCh := make(chan SmithWatermanResult, 10)

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
		for i := len(stMatrix[lenWord]) - 1; i >= 0; i-- {
			v := stMatrix[lenWord][i]
			if float32(v.Score) > threshAgreement {
				s := History(stMatrix, lenWord, i)
				smithCh <- SmithWatermanResult{
					MatchWord:         string(word),
					CompleteAgreement: completeAgreement,
					AppliedAgreement:  v.Score,
					SimilarScore:      float32(v.Score) / float32(completeAgreement),
					LastNodes:         nil,
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
		smithCh <- SmithWatermanResult{
			MatchWord:         "",
			LastNodes:         stMatrix[lenWord],
			CompleteAgreement: completeAgreement,
		}
		close(smithCh)
	}()
	return smithCh
}
