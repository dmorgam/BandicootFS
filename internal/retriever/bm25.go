package retriever

import "math"

// bm25 is a prebuilt BM25 index over a set of tokenized documents (chunks).
type bm25 struct {
	tf     []map[string]int   // term frequencies per document
	dl     []int              // document lengths (token counts)
	idf    map[string]float64 // inverse document frequency per term
	avgdl  float64
	k1, b  float64
}

// newBM25 indexes docs with the given BM25 parameters.
func newBM25(docs [][]string, k1, b float64) *bm25 {
	ix := &bm25{
		tf:  make([]map[string]int, len(docs)),
		dl:  make([]int, len(docs)),
		idf: make(map[string]float64),
		k1:  k1,
		b:   b,
	}

	df := make(map[string]int)
	total := 0
	for i, toks := range docs {
		tf := make(map[string]int, len(toks))
		for _, t := range toks {
			tf[t]++
		}
		ix.tf[i] = tf
		ix.dl[i] = len(toks)
		total += len(toks)
		for t := range tf {
			df[t]++
		}
	}

	n := float64(len(docs))
	for t, d := range df {
		// Nonnegative BM25+ IDF: common terms tend toward 0, never negative.
		ix.idf[t] = math.Log(1 + (n-float64(d)+0.5)/(float64(d)+0.5))
	}

	ix.avgdl = float64(total) / n // n > 0: caller guards len(docs) == 0
	if ix.avgdl == 0 {
		ix.avgdl = 1 // all docs empty; avoid division by zero
	}
	return ix
}

// score returns the BM25 score of document i against the (deduplicated) query.
func (ix *bm25) score(i int, query map[string]struct{}) float64 {
	tf := ix.tf[i]
	dl := float64(ix.dl[i])
	norm := ix.k1 * (1 - ix.b + ix.b*dl/ix.avgdl)

	var s float64
	for t := range query {
		f := float64(tf[t])
		if f == 0 {
			continue
		}
		s += ix.idf[t] * (f * (ix.k1 + 1)) / (f + norm)
	}
	return s
}
