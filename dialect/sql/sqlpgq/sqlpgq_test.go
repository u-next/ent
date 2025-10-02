package sqlpgq

import (
	"fmt"
	"strconv"
	"testing"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlhint"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		input     sql.Querier
		wantQuery string
		wantArgs  []any
	}{
		{
			input: func() *GraphQuery {
				account := N().Named("account").Labels("Account")
				owner := N().Named("owner").Labels("Person")
				return Graph("FinGraph").
					Match(
						From(N().Labels("Account")).Via(E().Labels("Transfers").RightDirection()).To(account),
					).
					Return(
						Column(account.F()),
						Column(sql.Count("*")).As("num_incoming_transfers"),
					).
					GroupBy(
						Column(account.F()),
					).
					Next().
					Match(
						To(account).Via(E().Labels("Owns").LeftDirection()).From(owner),
					).
					Return(
						Column(account.F("id")).As("account_id"),
						Column(owner.F("name")).As("owner_name"),
						Column("`num_incoming_transfers`"),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (:`Account`)-[:`Transfers`]->(`account`:`Account`)\n" +
				"RETURN `account`, COUNT(*) AS `num_incoming_transfers`\n" +
				"GROUP BY `account`\n\n" +
				"NEXT\n\n" +
				"MATCH (`account`:`Account`)<-[:`Owns`]-(`owner`:`Person`)\n" +
				"RETURN `account`.`id` AS `account_id`, `owner`.`name` AS `owner_name`, `num_incoming_transfers`",
		},
		{
			input: func() *GraphQuery {
				p := N().Named("p").Labels("Person")
				a := N().Named("a").Labels("Account")
				o := E().Named("o").Labels("Owns")
				return Graph("FinGraph").
					Match(
						From(p).Via(o.RightDirection()).To(a),
					).
					Filter(
						sql.NEQ(p.F("Id"), "1"),
					).
					Return(
						Column(p.F("name")),
						Column(a.F("Id")).As("account_id"),
					)
			}(),
			wantArgs: []any{"1"},
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person`)-[`o`:`Owns`]->(`a`:`Account`)\n" +
				"FILTER `p`.`Id` <> ?\n" +
				"RETURN `p`.`name`, `a`.`Id` AS `account_id`",
		},
		{
			input: func() *GraphQuery {
				p := N().Named("p").Labels("Person")
				a := N().Named("a").Labels("Account")
				o := E().Named("o").Labels("Owns")
				iter := For("element").In(sql.Expr("[\"all\",\"some\"]")).WithOffset()
				return Graph("FinGraph").
					Match(
						From(p).Via(o.RightDirection()).To(a),
					).
					For(iter).
					Return(
						Column(p.F("Id")),
						Column(iter.Elem()).As("alert_type"),
						Column(iter.Offset()),
					).
					OrderBy(
						Column(p.F("Id")),
						Column(iter.Elem()),
						Column(iter.Offset()),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person`)-[`o`:`Owns`]->(`a`:`Account`)\n" +
				"FOR `element` IN [\"all\",\"some\"] WITH OFFSET\n" +
				"RETURN `p`.`Id`, `element` AS `alert_type`, `offset`\n" +
				"ORDER BY `p`.`Id`, `element`, `offset`",
		},
		{
			input: func() *GraphQuery {
				source := N().Named("source").Labels("Account")
				destination := N().Named("destination").Labels("Account")
				e := E().Named("e").Labels("Transfers")
				a := Assign("a", sql.Expr(source.F()))
				return Graph("FinGraph").
					Match(
						From(source).Via(e.RightDirection()).To(destination),
					).
					Let(a).
					Return(
						Column(a.F("id")).As("a_id"),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`source`:`Account`)-[`e`:`Transfers`]->(`destination`:`Account`)\n" +
				"LET `a` = `source`\n" +
				"RETURN `a`.`id` AS `a_id`",
		},
		{
			input: func() *GraphQuery {
				source := N().Named("source").Labels("Account")
				destination := N().Named("destination").Labels("Account")
				e := E().Named("e").Labels("Transfers")
				return Graph("FinGraph").
					Match(
						From(source).Via(e.RightDirection()).To(destination),
					).
					OrderBy(
						Column(source.F("Id")),
					).
					Limit(3).
					Return(
						Column(source.F("Id")),
						Column(source.F("nick_name")),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`source`:`Account`)-[`e`:`Transfers`]->(`destination`:`Account`)\n" +
				"ORDER BY `source`.`Id`\n" +
				"LIMIT 3\n" +
				"RETURN `source`.`Id`, `source`.`nick_name`",
		},
		{
			input: func() *GraphQuery {
				p := N().Named("p").Labels("Person")
				return Graph("FinGraph").
					Match(p).
					Offset(2).
					Return(
						Column(p.F("name")),
						Column(p.F("id")),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person`)\n" +
				"OFFSET 2\n" +
				"RETURN `p`.`name`, `p`.`id`",
		},
		{
			input: func() *GraphQuery {
				p := N().Named("p").Labels("Person")
				return Graph("FinGraph").
					Match(p).
					Return(
						Column(p.F("name")),
						Column(p.F("id")),
					).
					Limit(1).
					Offset(1)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person`)\n" +
				"RETURN `p`.`name`, `p`.`id`\n" +
				"LIMIT 1\n" +
				"OFFSET 1",
		},
		{
			input: func() *GraphQuery {
				src := N().Named("src").Labels("Account")
				dst := N().Named("dst").Labels("Account")
				transfer := E().Named("transfer").Labels("Transfers")
				return Graph("FinGraph").
					Match(
						From(src).Via(transfer.RightDirection()).To(dst),
					).
					WithDistinct(
						Column(dst.F()),
					).
					Return(
						Column(dst.F("id")).As("destination_id"),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`src`:`Account`)-[`transfer`:`Transfers`]->(`dst`:`Account`)\n" +
				"WITH DISTINCT `dst`\n" +
				"RETURN `dst`.`id` AS `destination_id`",
		},
		{
			input: func() *GraphQuery {
				p := N().Named("p").Labels("Person")
				return Graph("FinGraph").
					Match(p).
					Return(
						Column(p.F("name")),
						Const(1).As("group_id"),
					).
					UnionAll().
					Match(p).
					Return(
						Const(2).As("group_id"),
						Column(p.F("name")),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person`)\n" +
				"RETURN `p`.`name`, ? AS `group_id`\n" +
				"UNION ALL\n" +
				"MATCH (`p`:`Person`)\n" +
				"RETURN ? AS `group_id`, `p`.`name`",
			wantArgs: []any{1, 2},
		},
		{
			input: func() *GraphQuery {
				p := N().Named("p").Labels("Person").Property("id", sql.Expr("?", 1))
				a := N().Named("a").Labels("Account")
				e := E().Named("e").Labels("Transfers")
				oa := N().Named("oa").Labels("Account")
				return Graph("FinGraph").
					Match(
						From(p).Via(E().Labels("Owns").RightDirection()).To(a),
					).
					MatchWithHint(
						sqlhint.JoinHint{sqlhint.JoinMethod: sqlhint.JoinMethodApplyJoin},
						From(a).Via(e.RightDirection()).To(oa),
					).
					Return(
						Column(oa.F("id")),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person` {id: ?})-[:`Owns`]->(`a`:`Account`)\n" +
				"MATCH @{JOIN_METHOD=APPLY_JOIN} (`a`:`Account`)-[`e`:`Transfers`]->(`oa`:`Account`)\n" +
				"RETURN `oa`.`id`",
			wantArgs: []any{1},
		},
		{
			input: func() *sql.Selector {
				n := N().Named("n").Labels("Person")
				return sql.Select(
					n.F("name"),
					n.F("id"),
				).From(
					GraphTable(Graph("FinGraph").
						Match(n).
						Return(
							Column(n.F()),
						),
					).As("PersonNames"),
				)
			}(),
			wantQuery: "SELECT `n`.`name`, `n`.`id` FROM GRAPH_TABLE (\n" +
				"  `FinGraph`\n" +
				"  MATCH (`n`:`Person`)\n" +
				"  RETURN `n`\n" +
				") AS `PersonNames`",
		},
		{
			input: func() *GraphQuery {
				p := N().
					Named("p").
					LabelExpr(
						OrL(
							NodeTable("Singer").L(),
							AndL(
								NotL(NodeTable("Writer").L()),
								NotL(NodeTable("Producer").L())),
						),
					)
				return Graph("FinGraph").
					Match(p).
					Return(
						Column(p.F("id")),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Singer`|(!`Writer`&!`Producer`))\n" +
				"RETURN `p`.`id`",
		},
		{
			input: func() *GraphQuery {
				src := N().Named("src").Labels("Account")
				dst := N().Named("dst").Labels("Account")
				transfer := E().Labels("Transfers")
				subpath := From(N().Labels("Account")).Via(transfer.RightDirection()).To(N().Named("mid").Labels("Account").Property("is_blocked", sql.Expr("?", true)))
				lower := 1
				upper := 2
				return Graph("FinGraph").
					Match(
						From(src).Path(subpath.Bounded(&lower, &upper)).Via(transfer.RightDirection()).To(dst),
					).
					Return(
						Column(src.F("id")).As("src_account_id"),
						Column(dst.F("id")).As("dst_account_id"),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`src`:`Account`)((:`Account`)-[:`Transfers`]->(`mid`:`Account` {is_blocked: ?})){1, 2}-[:`Transfers`]->(`dst`:`Account`)\n" +
				"RETURN `src`.`id` AS `src_account_id`, `dst`.`id` AS `dst_account_id`",
			wantArgs: []any{true},
		},
		{
			input: func() *GraphQuery {
				src := N().Named("src").Labels("Account")
				mid := N().Named("mid").Labels("Account")
				dst := N().Named("dst").Labels("Account")
				t1 := E().Named("t1").Labels("Transfers")
				t2 := E().Named("t2").Labels("Transfers")
				p := From(src).Via(t1.RightDirection()).To(mid).Variable("p")
				q := From(mid).Via(t2.RightDirection()).To(dst).Variable("q")
				fullPath := Assign("full_path", Concat(p, q))
				return Graph("FinGraph").
					Match(p, q).
					Let(fullPath).
					Return(
						Column(fmt.Sprintf("TO_JSON(%s)", fullPath.F())).As("results"),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH `p` = (`src`:`Account`)-[`t1`:`Transfers`]->(`mid`:`Account`), `q` = (`mid`:`Account`)-[`t2`:`Transfers`]->(`dst`:`Account`)\n" +
				"LET `full_path` = `p` || `q`\n" +
				"RETURN TO_JSON(`full_path`) AS `results`",
		},
		{
			input: func() *GraphQuery {
				p := N().Named("p").Labels("Person").Property("Name", sql.Expr("?", "Lee"))
				a := N().Named("a").Labels("Account")
				o := E().Named("o").Labels("Owns")
				match := Match(
					From(p).Via(o.RightDirection()).To(a),
				)
				return Graph("FinGraph").
					Return(
						ColumnExpr(ExistsQuery(match)).As("results"),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"RETURN EXISTS {\n" +
				"  MATCH (`p`:`Person` {Name: ?})-[`o`:`Owns`]->(`a`:`Account`)\n" +
				"} AS `results`",
			wantArgs: []any{"Lee"},
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
