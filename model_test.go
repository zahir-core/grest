package grest

import (
	"encoding/json"
	"testing"
)

func TestModelGetQuerySchema(t *testing.T) {
	expected := expectedSchemaStr()
	result := ""
	a := &Article{}
	resultByte, err := json.MarshalIndent(a.GetSchema(), "", "  ")
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("json.MarshalIndent(a.GetSchema(), \"\", \"  \") [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

type Article struct {
	Model
	ID          NullUUID     `json:"id"           db:"a.id"`
	Title       NullString   `json:"title"        db:"a.title"`
	Content     NullString   `json:"content"      db:"a.content"`
	AuthorID    NullUUID     `json:"author.id"    db:"a.author_id"`
	AuthorName  NullString   `json:"author.name"  db:"u.name"`
	AuthorEmail NullString   `json:"author.email" db:"u.email"`
	Categories  []Category   `json:"categories"   db:"article.id={id}&is_active=true"` // {id} will be replaced to parent id, parsed using String{}.GetVars
	Detail      NullJSON     `json:"detail"       db:"a.detail"`
	TotalReview NullFloat64  `json:"total_review" db:"coalesce(tr.total_review,0)"`
	IsActive    NullBool     `json:"is_active"    db:"a.is_active"`
	IsHidden    NullBool     `json:"is_hidden"    db:"a.is_hidden,hide"`
	CreatedAt   NullDateTime `json:"created_at"   db:"a.created_at"`
	UpdatedAt   NullDateTime `json:"updated_at"   db:"a.updated_at"`
	DeletedAt   NullDateTime `json:"deleted_at"   db:"a.deleted_at"`
}

func (Article) TableVersion() string {
	return "22.02.080822"
}

func (Article) TableName() string {
	return "articles"
}

func (Article) TableAliasName() string {
	return "u"
}

func (m *Article) GetFields() map[string]map[string]any {
	m.SetFields(m)
	return m.Fields
}

func (m *Article) GetRelations() map[string]map[string]any {
	m.AddRelation("left", "users", "u", []map[string]any{{"column1": "u.id", "operator": "=", "column2": "a.author_id"}})
	totalReview := &TotalReview{}
	m.AddRelation("left", totalReview.GetSchema(), "tr", []map[string]any{{"column1": "tr.id", "operator": "=", "column2": "a.id"}})
	return m.Relations
}

func (m *Article) GetFilters() []map[string]any {
	m.AddFilter(map[string]any{"column1": "a.deleted_at", "operator": "=", "value": nil})
	return m.Filters
}

func (m *Article) GetSorts() []map[string]any {
	m.AddSort(map[string]any{"column": "a.created_at", "direction": "desc"})
	return m.Sorts
}

func (m *Article) GetSchema() map[string]any {
	return m.SetSchema(m)
}

func (m *Article) GetOpenAPISchema() map[string]any {
	return m.SetOpenAPISchema(m)
}

type Category struct {
	Model
	ID          NullUUID     `json:"id"           db:"c.id"`
	Code        NullString   `json:"code"         db:"c.code"`
	Name        NullString   `json:"name"         db:"c.name"`
	IsActive    NullBool     `json:"is_active"    db:"c.is_active"`
	AuthorID    NullUUID     `json:"author.id"    db:"c.author_id"`
	AuthorName  NullString   `json:"author.name"  db:"u.name"`
	AuthorEmail NullString   `json:"author.email" db:"u.email"`
	CreatedAt   NullDateTime `json:"created.time" db:"c.created_at"`
	UpdatedAt   NullDateTime `json:"updated.time" db:"c.updated_at"`
	DeletedAt   NullDateTime `json:"deleted.time" db:"c.deleted_at"`
	ArticleID   NullUUID     `json:"article.id"   db:"ac.article_id,hide"`
}

func (Category) TableName() string {
	return "categories"
}

func (Category) TableVersion() string {
	return "22.02.210949"
}

func (Category) TableAliasName() string {
	return "c"
}

func (m *Category) GetFields() map[string]map[string]any {
	m.SetFields(m)
	return m.Fields
}

func (m *Category) GetRelations() map[string]map[string]any {
	m.AddRelation("inner", "articles_categories", "ac", []map[string]any{{"column1": "ac.category_id", "operator": "=", "column2": "c.id"}})
	m.AddRelation("left", "users", "u", []map[string]any{{"column1": "u.id", "operator": "=", "column2": "c.author_id"}})
	return m.Relations
}

func (m *Category) GetFilters() []map[string]any {
	m.AddFilter(map[string]any{"column1": "c.deleted_at", "operator": "=", "value": nil})
	return m.Filters
}

func (m *Category) GetSorts() []map[string]any {
	m.AddSort(map[string]any{"column": "c.code", "direction": "asc"})
	return m.Sorts
}

func (m *Category) GetSchema() map[string]any {
	return m.SetSchema(m)
}

func (m *Category) GetOpenAPISchema() map[string]any {
	return m.SetOpenAPISchema(m)
}

type TotalReview struct {
	Model
	ArticleID   NullUUID  `json:"id"           db:"r.article_id,group"`
	TotalReview NullInt64 `json:"total_review" db:"count(r.article_id)"`
}

func (TotalReview) TableName() string {
	return "reviews"
}

func (TotalReview) TableVersion() string {
	return "22.02.210949"
}

func (TotalReview) TableAliasName() string {
	return "r"
}

func (m *TotalReview) GetFields() map[string]map[string]any {
	m.SetFields(m)
	return m.Fields
}

func (m *TotalReview) GetSchema() map[string]any {
	return m.SetSchema(m)
}

func (m *TotalReview) GetOpenAPISchema() map[string]any {
	return m.SetOpenAPISchema(m)
}

func expectedSchemaStr() string {
	return `{
  "arrayFields": {
    "categories": {
      "filter": "article.id={id}\u0026is_active=true",
      "schema": {
        "arrayFields": null,
        "fields": {
          "article.id": {
            "as": "article.id",
            "db": "ac.article_id",
            "isGroup": false,
            "isHide": true,
            "type": "NullUUID"
          },
          "author.email": {
            "as": "author.email",
            "db": "u.email",
            "isGroup": false,
            "isHide": false,
            "type": "NullString"
          },
          "author.id": {
            "as": "author.id",
            "db": "c.author_id",
            "isGroup": false,
            "isHide": false,
            "type": "NullUUID"
          },
          "author.name": {
            "as": "author.name",
            "db": "u.name",
            "isGroup": false,
            "isHide": false,
            "type": "NullString"
          },
          "code": {
            "as": "code",
            "db": "c.code",
            "isGroup": false,
            "isHide": false,
            "type": "NullString"
          },
          "created.time": {
            "as": "created.time",
            "db": "c.created_at",
            "isGroup": false,
            "isHide": false,
            "type": "NullDateTime"
          },
          "deleted.time": {
            "as": "deleted.time",
            "db": "c.deleted_at",
            "isGroup": false,
            "isHide": false,
            "type": "NullDateTime"
          },
          "id": {
            "as": "id",
            "db": "c.id",
            "isGroup": false,
            "isHide": false,
            "type": "NullUUID"
          },
          "is_active": {
            "as": "is_active",
            "db": "c.is_active",
            "isGroup": false,
            "isHide": false,
            "type": "NullBool"
          },
          "name": {
            "as": "name",
            "db": "c.name",
            "isGroup": false,
            "isHide": false,
            "type": "NullString"
          },
          "updated.time": {
            "as": "updated.time",
            "db": "c.updated_at",
            "isGroup": false,
            "isHide": false,
            "type": "NullDateTime"
          }
        },
        "filters": [
          {
            "column1": "c.deleted_at",
            "operator": "=",
            "value": null
          }
        ],
        "groups": null,
        "isFlat": false,
        "relations": {
          "ac": {
            "conditions": [
              {
                "column1": "ac.category_id",
                "column2": "c.id",
                "operator": "="
              }
            ],
            "tableAliasName": "ac",
            "tableName": "articles_categories",
            "type": "inner"
          },
          "u": {
            "conditions": [
              {
                "column1": "u.id",
                "column2": "c.author_id",
                "operator": "="
              }
            ],
            "tableAliasName": "u",
            "tableName": "users",
            "type": "left"
          }
        },
        "sorts": [
          {
            "column": "c.code",
            "direction": "asc"
          }
        ]
      }
    }
  },
  "fields": {
    "author.email": {
      "as": "author.email",
      "db": "u.email",
      "isGroup": false,
      "isHide": false,
      "type": "NullString"
    },
    "author.id": {
      "as": "author.id",
      "db": "a.author_id",
      "isGroup": false,
      "isHide": false,
      "type": "NullUUID"
    },
    "author.name": {
      "as": "author.name",
      "db": "u.name",
      "isGroup": false,
      "isHide": false,
      "type": "NullString"
    },
    "content": {
      "as": "content",
      "db": "a.content",
      "isGroup": false,
      "isHide": false,
      "type": "NullString"
    },
    "created_at": {
      "as": "created_at",
      "db": "a.created_at",
      "isGroup": false,
      "isHide": false,
      "type": "NullDateTime"
    },
    "deleted_at": {
      "as": "deleted_at",
      "db": "a.deleted_at",
      "isGroup": false,
      "isHide": false,
      "type": "NullDateTime"
    },
    "detail": {
      "as": "detail",
      "db": "a.detail",
      "isGroup": false,
      "isHide": false,
      "type": "NullJSON"
    },
    "id": {
      "as": "id",
      "db": "a.id",
      "isGroup": false,
      "isHide": false,
      "type": "NullUUID"
    },
    "is_active": {
      "as": "is_active",
      "db": "a.is_active",
      "isGroup": false,
      "isHide": false,
      "type": "NullBool"
    },
    "is_hidden": {
      "as": "is_hidden",
      "db": "a.is_hidden",
      "isGroup": false,
      "isHide": true,
      "type": "NullBool"
    },
    "title": {
      "as": "title",
      "db": "a.title",
      "isGroup": false,
      "isHide": false,
      "type": "NullString"
    },
    "total_review": {
      "as": "total_review",
      "db": "coalesce(tr.total_review,0)",
      "isGroup": false,
      "isHide": false,
      "type": "NullFloat64"
    },
    "updated_at": {
      "as": "updated_at",
      "db": "a.updated_at",
      "isGroup": false,
      "isHide": false,
      "type": "NullDateTime"
    }
  },
  "filters": [
    {
      "column1": "a.deleted_at",
      "operator": "=",
      "value": null
    }
  ],
  "groups": null,
  "isFlat": false,
  "relations": {
    "tr": {
      "conditions": [
        {
          "column1": "tr.id",
          "column2": "a.id",
          "operator": "="
        }
      ],
      "tableAliasName": "tr",
      "tableName": {
        "arrayFields": null,
        "fields": {
          "id": {
            "as": "id",
            "db": "r.article_id",
            "isGroup": true,
            "isHide": false,
            "type": "NullUUID"
          },
          "total_review": {
            "as": "total_review",
            "db": "count(r.article_id)",
            "isGroup": false,
            "isHide": false,
            "type": "NullInt64"
          }
        },
        "filters": null,
        "groups": {
          "id": "r.article_id"
        },
        "isFlat": false,
        "relations": null,
        "sorts": null
      },
      "tableSchema": {
        "arrayFields": null,
        "fields": {
          "id": {
            "as": "id",
            "db": "r.article_id",
            "isGroup": true,
            "isHide": false,
            "type": "NullUUID"
          },
          "total_review": {
            "as": "total_review",
            "db": "count(r.article_id)",
            "isGroup": false,
            "isHide": false,
            "type": "NullInt64"
          }
        },
        "filters": null,
        "groups": {
          "id": "r.article_id"
        },
        "isFlat": false,
        "relations": null,
        "sorts": null
      },
      "type": "left"
    },
    "u": {
      "conditions": [
        {
          "column1": "u.id",
          "column2": "a.author_id",
          "operator": "="
        }
      ],
      "tableAliasName": "u",
      "tableName": "users",
      "type": "left"
    }
  },
  "sorts": [
    {
      "column": "a.created_at",
      "direction": "desc"
    }
  ]
}`
}
