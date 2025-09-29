package gql

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		input     Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: Graph("FinGraph").
				Match(Path().
					From(Node().Labels("Account")).
					Via(Edge().Labels("Transfers").RightDirection()).
					To(Node().Variable("account").Labels("Account")),
				).
				Return(
					Column("account"),
					Column("COUNT(*)").As("num_incoming_transfers"),
				).
				GroupBy(
					Column("account"),
				).
				Next().
				Match(Path().
					To(Node().Variable("account").Labels("Account")).
					Via(Edge().Labels("Owns").LeftDirection()).
					From(Node().Variable("owner").Labels("Person")),
				).
				Return(
					Column("account.id").As("account_id"),
					Column("owner.name").As("owner_name"),
					Column("num_incoming_transfers"),
				),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (:Account)-[:Transfers]->(account:Account)
RETURN account, COUNT(*) AS ` + "`num_incoming_transfers`" + `
GROUP BY account

NEXT

MATCH (account:Account)<-[:Owns]-(owner:Person)
RETURN account.id AS ` + "`account_id`, owner.name AS `owner_name`, num_incoming_transfers",
		},
		{
			input: Graph("FinGraph").
				Match(Path().
					From(Node().Variable("p").Labels("Person")).
					Via(Edge().Variable("o").Labels("Owns").RightDirection()).
					To(Node().Variable("a").Labels("Account")),
				).
				Filter(
					NEQ("p.Id", "1"),
				).
				Return(
					Column("p.name"),
					Column("a.Id").As("account_id"),
				),
			wantArgs: []any{"1"},
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (p:Person)-[o:Owns]->(a:Account)
FILTER ` + "`p.Id` <> ?" + `
RETURN p.name, a.Id AS ` + "`account_id`",
		},
		{
			input: Graph("FinGraph").
				Match(Path().
					From(Node().Variable("p").Labels("Person")).
					Via(Edge().Variable("o").Labels("Owns").RightDirection()).
					To(Node().Variable("a").Labels("Account")),
				).
				For(
					For("element").In(Expr("[\"all\",\"some\"]")).WithOffset(),
				).
				Return(
					Column("p.Id"),
					Column("element").As("alert_type"),
					Column("offset"),
				).
				OrderBy(
					Column("p.Id"),
					Column("element"),
					Column("offset"),
				),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (p:Person)-[o:Owns]->(a:Account)
FOR element IN ["all","some"] WITH OFFSET
RETURN p.Id, element AS ` + "`alert_type`, offset" + `
ORDER BY p.Id, element, offset`,
		},
		{
			input: Graph("FinGraph").
				Match(Path().
					From(Node().Variable("source").Labels("Account")).
					Via(Edge().Variable("e").Labels("Transfers").RightDirection()).
					To(Node().Variable("destination").Labels("Account")),
				).
				Let(
					Assign("a", Expr("source")),
				).
				Return(
					Column("a.id").As("a_id"),
				),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (source:Account)-[e:Transfers]->(destination:Account)
LET a = source
RETURN a.id AS ` + "`a_id`",
		},
		{
			input: Graph("FinGraph").
				Match(Path().
					From(Node().Variable("source").Labels("Account")).
					Via(Edge().Variable("e").Labels("Transfers").RightDirection()).
					To(Node().Variable("destination").Labels("Account")),
				).
				OrderBy(
					Column("source.Id"),
				).
				Limit(3).
				Return(
					Column("source.Id"),
					Column("source.nick_name"),
				),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (source:Account)-[e:Transfers]->(destination:Account)
ORDER BY source.Id
LIMIT 3
RETURN source.Id, source.nick_name`,
		},
		{
			input: Graph("FinGraph").
				Match(
					Node().Variable("p").Labels("Person"),
				).
				Offset(2).
				Return(
					Column("p.name"),
					Column("p.id"),
				),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (p:Person)
OFFSET 2
RETURN p.name, p.id`,
		},
		{
			input: Graph("FinGraph").
				Match(
					Node().Variable("p").Labels("Person"),
				).
				Return(
					Column("p.name"),
					Column("p.id"),
				).
				Limit(1).
				Offset(1),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (p:Person)
RETURN p.name, p.id
LIMIT 1
OFFSET 1`,
		},
		{
			input: Graph("FinGraph").
				Match(Path().
					From(Node().Variable("src").Labels("Account")).
					Via(Edge().Variable("transfer").Labels("Transfers").RightDirection()).
					To(Node().Variable("dst").Labels("Account")),
				).
				WithDistinct(
					Column("dst"),
				).
				Return(
					Column("dst.id").As("destination_id"),
				),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (src:Account)-[transfer:Transfers]->(dst:Account)
WITH DISTINCT dst
RETURN dst.id AS ` + "`destination_id`",
		},
		{
			input: Graph("FinGraph").
				Match(
					Node().Variable("p").Labels("Person"),
				).
				Return(
					Column("p.name"),
					Column("1").As("group_id"),
				).
				UnionAll().
				Match(
					Node().Variable("p").Labels("Person"),
				).
				Return(
					Column("2").As("group_id"),
					Column("p.name"),
				),
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (p:Person)
RETURN p.name, 1 AS ` + "`group_id`" + `
UNION ALL
MATCH (p:Person)
RETURN 2 AS ` + "`group_id`, p.name",
		},
		{
			input: Graph("FinGraph").
				Match(Path().
					From(Node().Variable("p").Labels("Person").Property("id", 1)).
					Via(Edge().Labels("Owns").RightDirection()).
					To(Node().Variable("a").Labels("Account")),
				).
				MatchWithHint(Hint{"JOIN_METHOD": "APPLY_JOIN"}, Path().
					From(Node().Variable("a").Labels("Account")).
					Via(Edge().Variable("e").Labels("Transfers").RightDirection()).
					To(Node().Variable("oa").Labels("Account")),
				).
				Return(
					Column("oa.id"),
				),
			wantArgs: []any{1},
			wantQuery: "GRAPH `FinGraph`" + `
MATCH (p:Person {id: ?})-[:Owns]->(a:Account)
MATCH @{JOIN_METHOD=APPLY_JOIN} (a:Account)-[e:Transfers]->(oa:Account)
RETURN oa.id`,
		},
		{
			input: GraphTable(Graph("FinGraph").
				Match(
					Node().Variable("n").Labels("Person"),
				).
				Return(
					Column("n.name"),
				),
			).As("PersonNames"),
			wantQuery: "GRAPH_TABLE(" + `
	` + "`FinGraph`" + `
	MATCH (n:Person)
	RETURN n.name
) AS ` + "`PersonNames`",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			query, args := tt.input.Query()
			require.Equal(t, tt.wantQuery, query)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}
