package ngword

import (
	"fmt"
	"strings"
	"unicode"
)

type Trie struct {
	Root *TrieNode
}

type TrieNode struct {
	Level    int
	Value    rune
	Children map[rune]*TrieNode
	End      bool
}

func NewTrie() Trie {
	var r Trie
	r.Root = NewTrieNode(0, '^')
	return r
}

func NewTrieNode(l int, v rune) *TrieNode {
	n := new(TrieNode)
	n.Children = make(map[rune]*TrieNode)
	n.Value = v
	n.Level = l
	return n
}

func (this *Trie) Append(txt string) {
	if len(txt) < 1 {
		return
	}
	node := this.Root
	key := []rune(txt)
	for i := 0; i < len(key); i++ {
		if _, exists := node.Children[key[i]]; !exists {
			node.Children[key[i]] = NewTrieNode(node.Level+1, key[i])
		}
		node = node.Children[key[i]]
	}

	node.End = true
}

func isNoneChar(r rune) bool {
	return (unicode.In(r, unicode.Han) || unicode.IsLetter(r) || unicode.IsDigit(r)) &&
		!unicode.IsPunct(r) && !unicode.IsSpace(r)
}

func replaceChars(words []rune, start, end int, rep rune) {
	for ; start < end; start++ {
		words[start] = rep
	}
}

func (this *Trie) Replace(txt string, rep rune) (string, bool) {
	if txt == "" {
		return "", false
	}
	origin := []rune(strings.ToLower(txt))
	words := []rune(txt)
	replace := false
	var (
		ok   bool
		node *TrieNode
	)
	for i, word := range origin {
		if node, ok = this.Root.Children[word]; !ok {
			continue
		}
		j := i + 1
		if node.End {
			replaceChars(words, i, j, rep)
		}
		for ; j < len(origin); j++ {
			if !isNoneChar(origin[j]) {
				continue
			}
			if v, ok := node.Children[origin[j]]; !ok {
				break
			} else {
				node = v
			}
		}
		if node.End {
			replace = true
			replaceChars(words, i, j, rep)
		}
	}
	return string(words), replace
}

func (this *Trie) PreOrder() <-chan *TrieNode {
	nodeCh := make(chan *TrieNode, 10)
	go func() {
		for _, child := range this.Root.Children {
			preOrder(child, nodeCh)
		}
		close(nodeCh)
	}()
	return nodeCh
}

func preOrder(node *TrieNode, ch chan<- *TrieNode) {
	ch <- node
	for _, child := range node.Children {
		preOrder(child, ch)
	}
}

func (this *Trie) Print() {
	node := this.Root
	for k, v := range node.Children {
		for k1, v1 := range v.Children {
			for k2, _ := range v1.Children {
				fmt.Printf("%s%s%s", string(k), string(k1), string(k2))
			}
		}
	}
}
