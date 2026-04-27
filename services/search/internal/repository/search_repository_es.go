package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lppduy/ecom-poc/services/search/internal/domain"
)

const indexName = "products"

type ESSearchRepository struct {
	client *elasticsearch.Client
}

func NewESSearchRepository(client *elasticsearch.Client) *ESSearchRepository {
	return &ESSearchRepository{client: client}
}

func (r *ESSearchRepository) Index(product domain.Product) error {
	doc, err := json.Marshal(product)
	if err != nil {
		return err
	}

	res, err := r.client.Index(
		indexName,
		bytes.NewReader(doc),
		r.client.Index.WithDocumentID(product.ID),
		r.client.Index.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES index error: %s", res.String())
	}
	return nil
}

func (r *ESSearchRepository) BulkIndex(products []domain.Product) error {
	if len(products) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, p := range products {
		// Action line
		meta := fmt.Sprintf(`{"index":{"_index":%q,"_id":%q}}`, indexName, p.ID)
		buf.WriteString(meta + "\n")

		// Document line
		doc, err := json.Marshal(p)
		if err != nil {
			return err
		}
		buf.Write(doc)
		buf.WriteByte('\n')
	}

	res, err := r.client.Bulk(
		strings.NewReader(buf.String()),
		r.client.Bulk.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES bulk error: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}
	if errors, ok := result["errors"].(bool); ok && errors {
		return fmt.Errorf("ES bulk had item-level errors")
	}
	return nil
}

func (r *ESSearchRepository) Search(query string, minPrice, maxPrice int) (domain.SearchResult, error) {
	must := []map[string]any{}

	if query != "" {
		must = append(must, map[string]any{
			"match": map[string]any{
				"name": map[string]any{
					"query":                query,
					"fuzziness":            "AUTO",
					"minimum_should_match": "75%",
				},
			},
		})
	} else {
		must = append(must, map[string]any{"match_all": map[string]any{}})
	}

	filter := []map[string]any{}
	if minPrice > 0 || maxPrice > 0 {
		priceRange := map[string]any{}
		if minPrice > 0 {
			priceRange["gte"] = minPrice
		}
		if maxPrice > 0 {
			priceRange["lte"] = maxPrice
		}
		filter = append(filter, map[string]any{
			"range": map[string]any{"price": priceRange},
		})
	}

	esQuery := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filter,
			},
		},
	}

	body, err := json.Marshal(esQuery)
	if err != nil {
		return domain.SearchResult{}, err
	}

	res, err := r.client.Search(
		r.client.Search.WithIndex(indexName),
		r.client.Search.WithBody(bytes.NewReader(body)),
		r.client.Search.WithContext(context.Background()),
	)
	if err != nil {
		return domain.SearchResult{}, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return domain.SearchResult{}, fmt.Errorf("ES search error: %s", res.String())
	}

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return domain.SearchResult{}, err
	}

	products := make([]domain.Product, 0, len(esResp.Hits.Hits))
	for _, h := range esResp.Hits.Hits {
		products = append(products, h.Source)
	}

	return domain.SearchResult{
		Total:    esResp.Hits.Total.Value,
		Products: products,
	}, nil
}
