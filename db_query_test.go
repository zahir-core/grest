package grest

import (
	"net/url"
	"regexp"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestDBQueryGeneral(t *testing.T) {

	db, mock, err := NewMockDB()
	if err != nil {
		t.Fatalf("Error occured : [%v]", err.Error())
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT 
		"a"."id" AS "id",
		"a"."title" AS "title",
		"a"."content" AS "content",
		"a"."author_id" AS "author.id",
		"u"."name" AS "author.name",
		"u"."email" AS "author.email",
		"a"."detail" AS "detail",
		"a"."is_active" AS "is_active",
		"a"."created_at" AS "created_at",
		"a"."updated_at" AS "updated_at",
		"a"."deleted_at" AS "deleted_at"
	FROM
		"articles" AS "u"
		LEFT JOIN "users" AS "u" ON "u"."id" = "a"."author_id"
	WHERE
		"a"."deleted_at" IS NULL
		AND (lower("u"."name") LIKE $1 OR "a"."is_active"=$2)
		AND (lower("a"."title") LIKE $3 OR lower("a"."content") LIKE $4 OR lower("u"."name") LIKE $5)
	ORDER BY
		"a"."updated_at" DESC
	LIMIT 10`)).
		// WithArgs("%foo%", 1, "%bar%", "%bar%", "%bar%"). // todo: check args with %
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"title",
			"content",
			"author.id",
			"author.name",
			"author.email",
			"detail",
			"is_active",
			"created_at",
			"updated_at",
			"deleted_at",
		}))

	q := url.Values{}
	// q.Add("author.email.$ilike", "user@email.com%")
	q.Add("$group", "id,title,$sum=total_review")
	q.Add("$or", "author.name.$ilike=foo||is_active=true")
	q.Add("$search", "title,content,author.name=bar")
	Find(db, &Article{}, q)
}
