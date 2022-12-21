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
	Categories  []Category   `json:"categories"   db:"ac.article_id=id"`
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
	m.AddRelation("left", "users", "u", []map[string]any{{"column_1": "u.id", "operator": "=", "column_2": "a.author_id"}})
	totalReview := &TotalReview{}
	m.AddRelation("left", totalReview.GetSchema(), "tr", []map[string]any{{"column_1": "tr.id", "operator": "=", "column_2": "a.id"}})
	return m.Relations
}

func (m *Article) GetFilters() []map[string]any {
	m.AddFilter(map[string]any{"column_1": "a.deleted_at", "operator": "=", "value": nil})
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
	ArticleID   NullUUID     `json:"-"            db:"ac.article_id"`
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
	m.AddRelation("inner", "articles_categories", "ac", []map[string]any{{"column_1": "ac.category_id", "operator": "=", "column_2": "c.id"}})
	m.AddRelation("left", "users", "u", []map[string]any{{"column_1": "u.id", "operator": "=", "column_2": "c.author_id"}})
	return m.Relations
}

func (m *Category) GetFilters() []map[string]any {
	m.AddFilter(map[string]any{"column_1": "c.deleted_at", "operator": "=", "value": nil})
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
  "array_fields": {
    "categories": {
      "filter": "ac.article_id=id",
      "schema": {
        "array_fields": null,
        "fields": {
          "author.email": {
            "as": "author.email",
            "db": "u.email",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullString",
            "validate": ""
          },
          "author.id": {
            "as": "author.id",
            "db": "c.author_id",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullUUID",
            "validate": ""
          },
          "author.name": {
            "as": "author.name",
            "db": "u.name",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullString",
            "validate": ""
          },
          "code": {
            "as": "code",
            "db": "c.code",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullString",
            "validate": ""
          },
          "created.time": {
            "as": "created.time",
            "db": "c.created_at",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullDateTime",
            "validate": ""
          },
          "deleted.time": {
            "as": "deleted.time",
            "db": "c.deleted_at",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullDateTime",
            "validate": ""
          },
          "id": {
            "as": "id",
            "db": "c.id",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullUUID",
            "validate": ""
          },
          "is_active": {
            "as": "is_active",
            "db": "c.is_active",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullBool",
            "validate": ""
          },
          "name": {
            "as": "name",
            "db": "c.name",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullString",
            "validate": ""
          },
          "updated.time": {
            "as": "updated.time",
            "db": "c.updated_at",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullDateTime",
            "validate": ""
          }
        },
        "filters": [
          {
            "column_1": "c.deleted_at",
            "operator": "=",
            "value": null
          }
        ],
        "groups": null,
        "is_flat": false,
        "relations": {
          "ac": {
            "conditions": [
              {
                "column_1": "ac.category_id",
                "column_2": "c.id",
                "operator": "="
              }
            ],
            "table_alias_name": "ac",
            "table_name": "articles_categories",
            "type": "inner"
          },
          "u": {
            "conditions": [
              {
                "column_1": "u.id",
                "column_2": "c.author_id",
                "operator": "="
              }
            ],
            "table_alias_name": "u",
            "table_name": "users",
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
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullString",
      "validate": ""
    },
    "author.id": {
      "as": "author.id",
      "db": "a.author_id",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullUUID",
      "validate": ""
    },
    "author.name": {
      "as": "author.name",
      "db": "u.name",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullString",
      "validate": ""
    },
    "categories": {
      "as": "categories",
      "db": "ac.article_id=id",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": true,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "",
      "validate": ""
    },
    "content": {
      "as": "content",
      "db": "a.content",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullString",
      "validate": ""
    },
    "created_at": {
      "as": "created_at",
      "db": "a.created_at",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullDateTime",
      "validate": ""
    },
    "deleted_at": {
      "as": "deleted_at",
      "db": "a.deleted_at",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullDateTime",
      "validate": ""
    },
    "detail": {
      "as": "detail",
      "db": "a.detail",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullJSON",
      "validate": ""
    },
    "id": {
      "as": "id",
      "db": "a.id",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullUUID",
      "validate": ""
    },
    "is_active": {
      "as": "is_active",
      "db": "a.is_active",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullBool",
      "validate": ""
    },
    "is_hidden": {
      "as": "is_hidden",
      "db": "a.is_hidden",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": true,
      "note": "",
      "title": "",
      "type": "NullBool",
      "validate": ""
    },
    "title": {
      "as": "title",
      "db": "a.title",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullString",
      "validate": ""
    },
    "total_review": {
      "as": "total_review",
      "db": "coalesce(tr.total_review",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullFloat64",
      "validate": ""
    },
    "updated_at": {
      "as": "updated_at",
      "db": "a.updated_at",
      "default": "",
      "example": "",
      "gorm": "",
      "is_array": false,
      "is_group": false,
      "is_hide": false,
      "note": "",
      "title": "",
      "type": "NullDateTime",
      "validate": ""
    }
  },
  "filters": [
    {
      "column_1": "a.deleted_at",
      "operator": "=",
      "value": null
    }
  ],
  "groups": null,
  "is_flat": false,
  "relations": {
    "tr": {
      "conditions": [
        {
          "column_1": "tr.id",
          "column_2": "a.id",
          "operator": "="
        }
      ],
      "table_alias_name": "tr",
      "table_name": {
        "array_fields": null,
        "fields": {
          "id": {
            "as": "id",
            "db": "r.article_id",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": true,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullUUID",
            "validate": ""
          },
          "total_review": {
            "as": "total_review",
            "db": "count(r.article_id)",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullInt64",
            "validate": ""
          }
        },
        "filters": null,
        "groups": {
          "id": "r.article_id"
        },
        "is_flat": false,
        "relations": null,
        "sorts": null
      },
      "table_schema": {
        "array_fields": null,
        "fields": {
          "id": {
            "as": "id",
            "db": "r.article_id",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": true,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullUUID",
            "validate": ""
          },
          "total_review": {
            "as": "total_review",
            "db": "count(r.article_id)",
            "default": "",
            "example": "",
            "gorm": "",
            "is_array": false,
            "is_group": false,
            "is_hide": false,
            "note": "",
            "title": "",
            "type": "NullInt64",
            "validate": ""
          }
        },
        "filters": null,
        "groups": {
          "id": "r.article_id"
        },
        "is_flat": false,
        "relations": null,
        "sorts": null
      },
      "type": "left"
    },
    "u": {
      "conditions": [
        {
          "column_1": "u.id",
          "column_2": "a.author_id",
          "operator": "="
        }
      ],
      "table_alias_name": "u",
      "table_name": "users",
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
