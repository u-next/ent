package sqlpgq

import (
	"strconv"
	"testing"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlhint"
	"entgo.io/ent/dialect/sql/sqljson"
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
						account.F(),
						sql.As(sql.Count("`*`"), "num_incoming_transfers"),
					).
					GroupBy(account.F()).
					Next().
					Match(
						To(account).Via(E().Labels("Owns").LeftDirection()).From(owner),
					).
					ReturnAs(account.F("id"), "account_id").
					ReturnAs(owner.F("name"), "owner_name").
					Return("`num_incoming_transfers`")
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (:`Account`)-[:`Transfers`]->(`account`:`Account`)\n" +
				"RETURN `account`, COUNT(`*`) AS `num_incoming_transfers`\n" +
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
					Return(p.F("name")).
					ReturnAs(a.F("Id"), "account_id")
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
				iter := Element("element").In(sql.Expr("?", []string{"all", "some"})).WithOffset()
				return Graph("FinGraph").
					Match(
						From(p).Via(o.RightDirection()).To(a),
					).
					For(iter).
					Return(p.F("Id")).
					ReturnAs(iter.Elem(), "alert_type").
					Return(iter.Offset()).
					OrderBy(
						p.F("Id"),
						iter.Elem(),
						iter.Offset(),
					)
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person`)-[`o`:`Owns`]->(`a`:`Account`)\n" +
				"FOR `element` IN ? WITH OFFSET\n" +
				"RETURN `p`.`Id`, `element` AS `alert_type`, `offset`\n" +
				"ORDER BY `p`.`Id`, `element`, `offset`",
			wantArgs: []any{[]string{"all", "some"}},
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
					ReturnAs(a.F("id"), "a_id")
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
						source.F("Id"),
					).
					Limit(3).
					Return(
						source.F("Id"),
						source.F("nick_name"),
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
						p.F("name"),
						p.F("id"),
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
						p.F("name"),
						p.F("id"),
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
						dst.F(),
					).
					ReturnAs(dst.F("id"), "destination_id")
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
					Return(p.F("name")).
					ReturnAs("1", "group_id").
					UnionAll().
					Match(p).
					ReturnAs("2", "group_id").
					Return(p.F("name"))
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`p`:`Person`)\n" +
				"RETURN `p`.`name`, 1 AS `group_id`\n" +
				"UNION ALL\n" +
				"MATCH (`p`:`Person`)\n" +
				"RETURN 2 AS `group_id`, `p`.`name`",
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
						oa.F("id"),
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
						Return(n.F()),
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
						p.F("id"),
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
						sql.As(src.F("id"), "src_account_id"),
						sql.As(dst.F("id"), "dst_account_id"),
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
					ReturnAs("TO_JSON(`full_path`)", "results")
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
					ReturnExprAs(ExistsQuery(match), "results")
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"RETURN EXISTS {\n" +
				"  MATCH (`p`:`Person` {Name: ?})-[`o`:`Owns`]->(`a`:`Account`)\n" +
				"} AS `results`",
			wantArgs: []any{"Lee"},
		},
		{
			input: func() *GraphQuery {
				src := N().Named("src").Labels("Account")
				mid := N().Named("mid").Labels("Account")
				dst := N().Named("dst").Labels("Account")
				t1 := E().Named("t1").Labels("Transfers")
				t2 := E().Named("t2").Labels("Transfers")
				p := Assign("p", sql.ExprFunc(Paths(src.F(), t1.F(), mid.F(), t2.F(), dst.F())))
				return Graph("FinGraph").
					Match(
						From(src).Via(t1.RightDirection()).To(mid).Via(t2.RightDirection()).To(dst),
					).
					Let(p).
					ReturnExprAs(sqljson.ValuePath("TO_JSON(`p`)[0]", sqljson.Path("labels")), "element_a").
					ReturnExprAs(sqljson.ValuePath("TO_JSON(`p`)[1]", sqljson.Path("labels")), "element_b").
					ReturnExprAs(sqljson.ValuePath("TO_JSON(`p`)[2]", sqljson.Path("labels")), "element_c")
			}(),
			wantQuery: "GRAPH `FinGraph`\n" +
				"MATCH (`src`:`Account`)-[`t1`:`Transfers`]->(`mid`:`Account`)-[`t2`:`Transfers`]->(`dst`:`Account`)\n" +
				"LET `p` = PATH(`src`, `t1`, `mid`, `t2`, `dst`)\n" +
				"RETURN JSON_QUERY(TO_JSON(`p`)[0], '$.labels') AS `element_a`, JSON_QUERY(TO_JSON(`p`)[1], '$.labels') AS `element_b`, JSON_QUERY(TO_JSON(`p`)[2], '$.labels') AS `element_c`",
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
