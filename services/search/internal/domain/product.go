package domain

// Product is the document shape stored in Elasticsearch.
type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// SearchResult wraps a list of matched products with total count.
type SearchResult struct {
	Total    int64     `json:"total"`
	Products []Product `json:"products"`
}
