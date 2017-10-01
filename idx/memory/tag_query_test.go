package memory

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/raintank/metrictank/idx"
	"gopkg.in/raintank/schema.v1"
)

func getTestIndex() (TagIndex, map[string]*idx.Archive) {
	data := [][]string{
		{"id1", "key1=value1", "key2=value2"},
		{"id2", "key1=value1", "key3=value3"},
		{"id3", "key1=value1", "key4=value4"},
		{"id4", "key1=value1", "key4=value3", "key3=value3"},
		{"id5", "key2=value1", "key5=value4", "key3=value3"},
		{"id6", "key2=value2", "key4=value5"},
		{"id7", "key3=value1", "key4=value4"},
	}

	tagIdx := make(TagIndex)
	byId := make(map[string]*idx.Archive)

	for _, d := range data {
		byId[d[0]] = &idx.Archive{}
		byId[d[0]].Tags = d[1:]
		for _, tag := range d[1:] {
			tagSplits := strings.Split(tag, "=")
			if _, ok := tagIdx[tagSplits[0]]; !ok {
				tagIdx[tagSplits[0]] = make(map[string]map[string]struct{})
			}

			if _, ok := tagIdx[tagSplits[0]][tagSplits[1]]; !ok {
				tagIdx[tagSplits[0]][tagSplits[1]] = make(map[string]struct{})
			}

			tagIdx[tagSplits[0]][tagSplits[1]][d[0]] = struct{}{}
		}
	}

	return tagIdx, byId
}

func queryAndCompareResults(t *testing.T, q *TagQuery, expectedData map[string]struct{}) {
	t.Helper()
	tagIdx, byId := getTestIndex()

	res, err := q.Run(tagIdx, byId)
	if err != nil {
		t.Fatalf("Unexpected error when running query: %q", err)
	}

	if !reflect.DeepEqual(expectedData, res) {
		t.Fatalf("Returned data does not match expected data:\nExpected: %+v\nGot: %+v", expectedData, res)
	}
}

func TestQueryByTagSimpleEqual(t *testing.T) {
	q, _ := NewTagQuery([]string{"key1=value1", "key3=value3"})
	expect := make(map[string]struct{})
	expect["id2"] = struct{}{}
	expect["id4"] = struct{}{}
	queryAndCompareResults(t, q, expect)
}

func TestQueryByTagSimplePattern(t *testing.T) {
	q, _ := NewTagQuery([]string{"key4=~value[43]", "key3=~value[1-3]"})
	expect := make(map[string]struct{})
	expect["id7"] = struct{}{}
	expect["id4"] = struct{}{}
	queryAndCompareResults(t, q, expect)
}

func TestQueryByTagSimpleUnequal(t *testing.T) {
	q, _ := NewTagQuery([]string{"key1=value1", "key4!=value4"})
	expect := make(map[string]struct{})
	expect["id1"] = struct{}{}
	expect["id2"] = struct{}{}
	expect["id4"] = struct{}{}
	queryAndCompareResults(t, q, expect)
}

func TestQueryByTagSimpleNotPattern(t *testing.T) {
	q, _ := NewTagQuery([]string{"key1=~value?", "key4!=~value[0-9]", "key2!=~va.+"})
	expect := make(map[string]struct{})
	expect["id2"] = struct{}{}
	queryAndCompareResults(t, q, expect)
}

func TestQueryByTagWithEqualEmpty(t *testing.T) {
	q, _ := NewTagQuery([]string{"key1=value1", "key2=", "key2=~"})
	expect := make(map[string]struct{})
	expect["id2"] = struct{}{}
	expect["id3"] = struct{}{}
	expect["id4"] = struct{}{}
	queryAndCompareResults(t, q, expect)
}

func TestQueryByTagWithUnequalEmpty(t *testing.T) {
	q, _ := NewTagQuery([]string{"key1=value1", "key3!=", "key3!=~"})
	expect := make(map[string]struct{})
	expect["id2"] = struct{}{}
	expect["id4"] = struct{}{}
	queryAndCompareResults(t, q, expect)
}

func TestQueryByTagInvalidQuery(t *testing.T) {
	_, err := NewTagQuery([]string{"key!=value1"})
	if err != errInvalidQuery {
		t.Fatalf("Expected an error, but didn't get it")
	}
}

func TestGetByTag(t *testing.T) {
	tagSupport := TagSupport
	defer func() { TagSupport = tagSupport }()
	TagSupport = true

	ix := New()
	ix.Init()

	mds := make([]schema.MetricData, 20)
	for i := range mds {
		mds[i].Metric = fmt.Sprintf("metric.%d", i)
		mds[i].Name = mds[i].Metric
		mds[i].Id = mds[i].Metric
		mds[i].OrgId = 1
		mds[i].Interval = 1
	}
	mds[1].Tags = []string{"key1=value1", "key2=value2"}
	mds[11].Tags = []string{"key1=value1"}
	mds[18].Tags = []string{"key1=value2", "key2=value2"}
	mds[3].Tags = []string{"key1=value1", "key3=value3"}

	for _, md := range mds {
		ix.AddOrUpdate(&md, 1)
	}

	type testCase struct {
		expressions []string
		expectation []string
	}

	testCases := []testCase{
		{
			expressions: []string{"key1=value1"},
			expectation: []string{"metric.1", "metric.11", "metric.3"},
		}, {
			expressions: []string{"key1=value2"},
			expectation: []string{"metric.18"},
		}, {
			expressions: []string{"key1=~value[0-9]"},
			expectation: []string{"metric.1", "metric.11", "metric.18", "metric.3"},
		}, {
			expressions: []string{"key1=~value[23]"},
			expectation: []string{"metric.18"},
		}, {
			expressions: []string{"key1=value1", "key2=value1"},
			expectation: []string{},
		}, {
			expressions: []string{"key1=value1", "key2=value2"},
			expectation: []string{"metric.1"},
		}, {
			expressions: []string{"key1=~value[12]", "key2=value2"},
			expectation: []string{"metric.1", "metric.18"},
		}, {
			expressions: []string{"key1=~value1", "key1=value2"},
			expectation: []string{},
		}, {
			expressions: []string{"key1=~value[0-9]", "key2=~", "key3!=value3"},
			expectation: []string{"metric.11"},
		}, {
			expressions: []string{"key2=", "key1=value1"},
			expectation: []string{"metric.11", "metric.3"},
		},
	}

	for _, tc := range testCases {
		tagQuery, err := NewTagQuery(tc.expressions)
		if err != nil {
			t.Fatalf("Got an unexpected error with query %s: %s", tc.expressions, err)
		}
		res := ix.IdsByTagQuery(1, tagQuery)
		if len(res) != len(tc.expectation) {
			t.Fatalf("Result does not match expectation for expressions %+v\nGot:\n%+v\nExpected:\n%+v\n", tc.expressions, res, tc.expectation)
		}
		expectationMap := make(map[string]struct{})
		for _, v := range tc.expectation {
			expectationMap[v] = struct{}{}
		}
		if !reflect.DeepEqual(res, expectationMap) {
			t.Fatalf("Result does not match expectation\nGot:\n%+v\nExpected:\n%+v\n", res, tc.expectation)
		}
	}
}

func TestDeleteTaggedSeries(t *testing.T) {
	tagSupport := TagSupport
	defer func() { TagSupport = tagSupport }()
	TagSupport = true

	ix := New()
	ix.Init()

	orgId := 1

	mds := getMetricData(orgId, 2, 50, 10, "metric.public")
	mds[10].Tags = []string{"key1=value1", "key2=value2"}

	for _, md := range mds {
		ix.AddOrUpdate(md, 1)
	}

	tagQuery, _ := NewTagQuery([]string{"key1=value1", "key2=value2"})
	res := ix.IdsByTagQuery(orgId, tagQuery)

	if len(res) != 1 {
		t.Fatalf("Expected to get 1 result, but got %d", len(res))
	}

	if len(ix.Tags[orgId]) != 2 {
		t.Fatalf("Expected tag index to contain 2 keys, but it does not: %+v", ix.Tags)
	}

	deleted, err := ix.Delete(orgId, mds[10].Metric)
	if err != nil {
		t.Fatalf("Error deleting metric: %s", err)
	}

	if len(deleted) != 1 {
		t.Fatalf("Expected 1 metric to get deleted, but got %d", len(deleted))
	}

	res = ix.IdsByTagQuery(orgId, tagQuery)

	if len(res) != 0 {
		t.Fatalf("Expected to get 0 results, but got %d", len(res))
	}

	if len(ix.Tags[orgId]) > 0 {
		t.Fatalf("Expected tag index to be empty, but it is not: %+v", ix.Tags)
	}
}