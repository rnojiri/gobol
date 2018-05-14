package solar

import (
	"github.com/uol/go-solr/solr"
)

/**
* Contains all search related functions.
* @author rnojiri
**/

// buildBasicQuery - builds a basic query
func (ss *SolrService) buildBasicQuery(collection, query, fields string, start, rows int) *solr.Query {

	q := solr.NewQuery()
	q.Q(query)

	if fields != "" {
		q.FieldList(fields)
	}

	q.Start(start)
	q.Rows(rows)

	return q
}

// builFilteredQuery - builds a basic query
func (ss *SolrService) buildFilteredQuery(collection, query, fields string, start, rows int, filterQueries []string) *solr.Query {

	q := ss.buildBasicQuery(collection, query, fields, start, rows)

	if filterQueries != nil && len(filterQueries) > 0 {
		for _, fq := range filterQueries {
			q.FilterQuery(fq)
		}
	}

	return q
}

// SimpleQuery - queries the solr
func (ss *SolrService) SimpleQuery(collection, query, fields string, start, rows int) (*solr.SolrResult, error) {

	si, err := ss.getSolrInterface(collection)
	if err != nil {
		return nil, err
	}

	q := ss.buildBasicQuery(collection, query, fields, start, rows)
	s := si.Search(q)
	r, err := s.Result(nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// FilteredQuery - queries the solr
func (ss *SolrService) FilteredQuery(collection, query, fields string, start, rows int, filterQueries []string) (*solr.SolrResult, error) {

	si, err := ss.getSolrInterface(collection)
	if err != nil {
		return nil, err
	}

	q := ss.buildFilteredQuery(collection, query, fields, start, rows, filterQueries)
	s := si.Search(q)
	r, err := s.Result(nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// addFacets - add facets to the query
func (ss *SolrService) addFacets(q *solr.Query, facetFields []string) {

	if facetFields != nil && len(facetFields) > 0 {
		for _, facetField := range facetFields {
			q.AddFacet(facetField)
		}
	}
}

// Facets - facets the solr
func (ss *SolrService) Facets(collection, query, fields string, start, rows int, facetFields []string) (*solr.SolrResult, error) {

	si, err := ss.getSolrInterface(collection)
	if err != nil {
		return nil, err
	}

	q := ss.buildBasicQuery(collection, query, fields, start, rows)
	ss.addFacets(q, facetFields)

	s := si.Search(q)
	r, err := s.Result(nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// BlockJoinFacets - block join facets the solr
func (ss *SolrService) BlockJoinFacets(collection, query, fields string, start, rows int, filterQueries []string, facetFields []string) (*solr.SolrResult, error) {

	si, err := ss.getSolrInterface(collection)
	if err != nil {
		return nil, err
	}

	q := ss.buildFilteredQuery(collection, query, fields, start, rows, filterQueries)
	ss.addFacets(q, facetFields)

	s := si.Search(q)
	r, err := s.BlockJoinFaceting(nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}
