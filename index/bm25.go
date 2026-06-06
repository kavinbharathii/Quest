
package index

import (
	"math"
	"sort"
	"strings"

	"github.com/kavinbharathii/quest/db"
)

const (
	bm25K1 = 1.5
	bm25B  = 0.75
)

type Result struct {
	Command		string
	Score		float64
	Frequency	int
}

type BM25 struct {
	docs	[]db.Command
	idf		map[string]float64
	avgLen	float64
}

func Tokenize(cmd string) []string {
	cmd = strings.ToLower(cmd)

	for _, ch := range []string{"-", "_", ".", "/", "=", ":", ",", "\"", "'"} {
		cmd = strings.ReplaceAll(cmd, ch, " ")
	}

	raw := strings.Fields(cmd)
	var tokens []string

	for _, t := range raw {
		t = strings.TrimSpace(t)
		if len(t) == 0 {
			continue
		}

		if isNumeric(t) && len(t) > 5 {
			continue
		}

		tokens = append(tokens, t)
	}

	return tokens
}

func TokenizeToString(cmd string) string {
	return strings.Join(Tokenize(cmd), " ")
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func Build(cmds []db.Command) *BM25 {
	b := &BM25 {
		docs: cmds,
		idf:  make(map[string]float64),
	}

	N := float64(len(cmds))
	if N == 0 {
		return b
	}

	df := make(map[string]int)
	totalLen := 0

	for _, cmd := range cmds {
		tokens := strings.Fields(cmd.Tokens)
		totalLen += len(tokens)
		seen := make(map[string]bool)

		for _, t := range tokens {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}

	b.avgLen = float64(totalLen) / N

	for term, freq := range df {
		b.idf[term] = math.Log((N - float64(freq) + 0.5)/(float64(freq) + 0.5) + 1)
	}

	return b
}

func (b *BM25) Search (query string, topN int) []Result {
	queryTokens := Tokenize(query)

	if len(queryTokens) == 0 || len(b.docs) == 0 {
		return nil
	}

	results := make([]Result, 0, len(b.docs))

	for _, cmd := range b.docs {
		docTokens := strings.Fields(cmd.Tokens)
		docLen := float64(len(docTokens))

		tf := make(map[string]float64)
		for _, t := range docTokens {
			tf[t] ++
		}

		score := 0.0
		for _, qt := range queryTokens {
			if tf[qt] == 0 {
				continue
			}

			idf := b.idf[qt]
			tfScore := tf[qt] * (bm25K1 + 1) /
				(tf[qt] + bm25K1 * (1 - bm25B + bm25B * (docLen / b.avgLen)))
			score += idf * tfScore
		}

		// Penalize cmds that occur a lot
		score *= 1 + math.Log1p(float64(cmd.Frequency)) * 0.1

		if score > 0 {
			results = append(results, Result {
				Command:	cmd.Command,
				Score:		score,
				Frequency:	cmd.Frequency,
			})
		}
	}

	sort.Slice(results, func (i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topN > len(results) {
		topN = len(results)
	}

	return results[:topN]
}


