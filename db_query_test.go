package grest

import (
	"database/sql/driver"
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
			"a"."author_id" AS "author.id",
			"a"."content" AS "content",
			"a"."created_at" AS "created_at",
			"a"."deleted_at" AS "deleted_at",
			"a"."detail" AS "detail",
			"a"."id" AS "id",
			"a"."is_active" AS "is_active",
			"a"."is_hidden" AS "is_hidden",
			"a"."title" AS "title",
			"a"."updated_at" AS "updated_at",
			"u"."email" AS "author.email",
			"u"."name" AS "author.name",
			coalesce(tr.total_review,0) AS "total_review"
		FROM 
			"articles" AS "a"
			LEFT JOIN "users" AS "u" ON "u"."id"="a"."author_id"
			LEFT JOIN (
				SELECT
					"r"."article_id" AS "id",
					count(r.article_id) AS "total_review"
				FROM "reviews" AS "r"
				GROUP BY
					"r"."article_id"
			) AS "tr" ON "tr"."id"="a"."id"
		WHERE
			"a"."deleted_at" IS NULL
			AND json_extract_path_text("detail"::json,'foo','bar') LIKE $1
		ORDER BY 
			"a"."created_at" DESC
		LIMIT 10`)).
		WithArgs(driver.Value("%baz%")).
		WillReturnRows(sqlmock.NewRows([]string{
			"author.id",
			"content",
			"created_at",
			"deleted_at",
			"detail",
			"id",
			"is_active",
			"is_hidden",
			"title",
			"updated_at",
			"author.email",
			"author.name",
			"total_review",
		}))

	q := url.Values{}
	// q.Add("author.email.$ilike", "user@email.com%")
	// q.Add("$select", "id,title,$sum:total_review")
	// q.Add("$or", "author.name.$ilike=foo||is_active=true")
	// q.Add("$search", "title,content,author.name=bar")
	q.Add("detail.foo.bar.$like", "baz")
	Find(db, &Article{}, q)
}
