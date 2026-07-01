package retriever

// stopwords are common English words dropped before scoring; they add noise
// and carry little retrieval signal.
var stopwords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true, "not": true,
	"you": true, "all": true, "any": true, "can": true, "has": true, "had": true,
	"was": true, "were": true, "with": true, "that": true, "this": true, "then": true,
	"than": true, "they": true, "them": true, "their": true, "there": true, "here": true,
	"from": true, "into": true, "onto": true, "over": true, "under": true, "your": true,
	"our": true, "out": true, "off": true, "its": true, "it's": true, "have": true,
	"has'": true, "which": true, "what": true, "when": true, "where": true, "who": true,
	"whom": true, "how": true, "why": true, "will": true, "would": true, "should": true,
	"could": true, "shall": true, "may": true, "might": true, "must": true, "does": true,
	"did": true, "done": true, "doing": true, "been": true, "being": true, "about": true,
	"above": true, "below": true, "between": true, "each": true, "few": true, "more": true,
	"most": true, "some": true, "such": true, "only": true, "own": true, "same": true,
	"other": true, "both": true, "because": true, "while": true, "during": true, "before": true,
	"after": true, "again": true, "once": true, "very": true, "just": true, "also": true,
	"too": true, "one": true, "two": true, "get": true, "got": true, "let": true,
	"use": true, "used": true, "using": true, "via": true, "per": true,
}
