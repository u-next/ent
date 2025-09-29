package gql

import (
	"strconv"
	"testing"

	"entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		input     sql.Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: Graph("FinGraph").
				Match(
					Path().
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
				Match(
					Path().
						To(Node().Variable("account").Labels("Account")).
						Via(Edge().Labels("Owns").LeftDirection()).
						From(Node().Variable("owner").Labels("Person")),
				).
				Return(
					Column("account.id").As("account_id"),
					Column("owner.name").As("owner_name"),
					Column("num_incoming_transfers"),
				),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (:Account)-[:Transfers]->(account:Account)\n" +
				"RETURN account, COUNT(*) AS `num_incoming_transfers`\n" +
				"GROUP BY account\n\n" +
				"NEXT\n\n" +
				"MATCH (account:Account)<-[:Owns]-(owner:Person)\n" +
				"RETURN account.id AS `account_id`, owner.name AS `owner_name`, num_incoming_transfers",
		},
		{
			input: Graph("FinGraph").
				Match(
					Path().
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (p:Person)-[o:Owns]->(a:Account)\n" +
				"FILTER `p.Id` <> ?\n" +
				"RETURN p.name, a.Id AS `account_id`",
		},
		{
			input: Graph("FinGraph").
				Match(
					Path().
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (p:Person)-[o:Owns]->(a:Account)\n" +
				"FOR element IN [\"all\",\"some\"] WITH OFFSET\n" +
				"RETURN p.Id, element AS `alert_type`, offset\n" +
				"ORDER BY p.Id, element, offset",
		},
		{
			input: Graph("FinGraph").
				Match(
					Path().
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (source:Account)-[e:Transfers]->(destination:Account)\n" +
				"LET a = source\n" +
				"RETURN a.id AS `a_id`",
		},
		{
			input: Graph("FinGraph").
				Match(
					Path().
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (source:Account)-[e:Transfers]->(destination:Account)\n" +
				"ORDER BY source.Id\n" +
				"LIMIT 3\n" +
				"RETURN source.Id, source.nick_name",
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (p:Person)\n" +
				"OFFSET 2\n" +
				"RETURN p.name, p.id",
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (p:Person)\n" +
				"RETURN p.name, p.id\n" +
				"LIMIT 1\n" +
				"OFFSET 1",
		},
		{
			input: Graph("FinGraph").
				Match(
					Path().
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (src:Account)-[transfer:Transfers]->(dst:Account)\n" +
				"WITH DISTINCT dst\n" +
				"RETURN dst.id AS `destination_id`",
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
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (p:Person)\n" +
				"RETURN p.name, 1 AS `group_id`\n" +
				"UNION ALL\n" +
				"MATCH (p:Person)\n" +
				"RETURN 2 AS `group_id`, p.name",
		},
		{
			input: Graph("FinGraph").
				Match(
					Path().
						From(Node().Variable("p").Labels("Person").Property("id", 1)).
						Via(Edge().Labels("Owns").RightDirection()).
						To(Node().Variable("a").Labels("Account")),
				).
				MatchWithHint(
					Hint{"JOIN_METHOD": "APPLY_JOIN"},
					Path().
						From(Node().Variable("a").Labels("Account")).
						Via(Edge().Variable("e").Labels("Transfers").RightDirection()).
						To(Node().Variable("oa").Labels("Account")),
				).
				Return(
					Column("oa.id"),
				),
			wantArgs: []any{1},
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (p:Person {id: ?})-[:Owns]->(a:Account)\n" +
				"MATCH @{JOIN_METHOD=APPLY_JOIN} (a:Account)-[e:Transfers]->(oa:Account)\n" +
				"RETURN oa.id",
		},
		{
			input: sql.Select("n.name", "n.id").From(
				GraphTable(
					Graph("FinGraph").
						Match(
							Node().Variable("n").Labels("Person"),
						).
						Return(
							Column("n"),
						),
				).As("PersonNames"),
			),
			wantQuery: "SELECT `n.name`, `n.id` FROM GRAPH_TABLE(\n" +
				"\t`FinGraph`\n" +
				"\tMATCH (n:Person)\n" +
				"\tRETURN n\n" +
				") AS `PersonNames`",
		},
		{
			input: Graph("FinGraph").
				Match(
					Node().Variable("p").LabelExpr(OrL(L("Singer"), AndL(NotL(L("Writer")), NotL(L("Producer"))))),
				).
				Return(
					Column("p.id"),
				),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (p:Singer|(!Writer&!Producer))\n" +
				"RETURN p.id",
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
