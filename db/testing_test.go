package db

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"grest.dev/grest"
)

func TestMockDBCreate(t *testing.T) {
	db, mock, err := NewMockDB()
	if err != nil {
		t.Fatalf("Error occured : [%v]", err.Error())
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	contact := Contact{
		ID:   uuid.NewString(),
		Name: "Test 1",
	}

	mock.ExpectBegin()
	mock.ExpectExec(
		regexp.QuoteMeta(`INSERT INTO "contacts" ("id","name") VALUES ($1,$2)`)).
		WithArgs(contact.ID, contact.Name).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err = db.Create(&contact).Error; err != nil {
		t.Errorf("Failed to insert contact, got error: %v", err)
		t.FailNow()
	}
}

func TestMockDBFind(t *testing.T) {
	db, mock, err := NewMockDB()
	if err != nil {
		t.Fatalf("Error occured : [%v]", err.Error())
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	contact := Contact{}
	id := uuid.NewString()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
	"id", 
	"name" 
FROM 
	"contacts" 
WHERE 
	"id" = $1`)).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	if err = db.Select(`"id", "name"`).Where(`"id" = ?`, id).Find(&contact).Error; err != nil {
		t.Errorf("Failed to get contact, got error: %v", err)
		t.FailNow()
	}
}

type Contact struct {
	ID   string
	Name string
}

type Article struct {
	Model
	ID          grest.NullUUID     `json:"id"           db:"a.id"`
	Title       grest.NullString   `json:"title"        db:"a.title"`
	Content     grest.NullString   `json:"content"      db:"a.content"`
	AuthorID    grest.NullUUID     `json:"author.id"    db:"a.author_id"`
	AuthorName  grest.NullString   `json:"author.name"  db:"u.name"`
	AuthorEmail grest.NullString   `json:"author.email" db:"u.email"`
	Categories  []Category         `json:"categories"   db:"ac.article_id=id"`
	Detail      grest.NullJSON     `json:"detail"       db:"a.detail"`
	IsActive    grest.NullBool     `json:"is_active"    db:"a.is_active"`
	IsHidden    grest.NullBool     `json:"is_hidden"    db:"a.is_hidden,hide"`
	CreatedAt   grest.NullDateTime `json:"created_at"   db:"a.created_at"`
	UpdatedAt   grest.NullDateTime `json:"updated_at"   db:"a.updated_at"`
	DeletedAt   grest.NullDateTime `json:"deleted_at"   db:"a.deleted_at"`
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

func (a *Article) SetRelation() {
	a.Relation = append(a.Relation, NewRelation("left", "users", "u", []Filter{{Column: "u.id", Column2: "a.author_id"}}))
}

func (a *Article) SetFilter() {
	a.Filter = append(a.Filter, NewFilter("a.deleted_at", "=", nil))
}

func (a *Article) SetSort() {
	a.Sort = append(a.Sort, NewSort("a.updated_at", "desc"))
}

type Category struct {
	Model
	ID          grest.NullUUID     `json:"id"           db:"c.id"`
	Code        grest.NullString   `json:"code"         db:"c.code"`
	Name        grest.NullString   `json:"name"         db:"c.name"`
	IsActive    grest.NullBool     `json:"is_active"    db:"c.is_active"`
	AuthorID    grest.NullUUID     `json:"author.id"    db:"c.author_id"`
	AuthorName  grest.NullString   `json:"author.name"  db:"u.name"`
	AuthorEmail grest.NullString   `json:"author.email" db:"u.email"`
	CreatedAt   grest.NullDateTime `json:"created.time" db:"c.created_at"`
	UpdatedAt   grest.NullDateTime `json:"updated.time" db:"c.updated_at"`
	DeletedAt   grest.NullDateTime `json:"deleted.time" db:"c.deleted_at"`
	ArticleID   grest.NullUUID     `json:"-"            db:"ac.article_id"`
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

func (c *Category) SetRelation() {
	c.Relation = append(c.Relation, NewRelation("inner", "articles_categories", "ac", []Filter{{Column: "ac.category_id", Column2: "c.id"}}))
	c.Relation = append(c.Relation, NewRelation("left", "users", "u", []Filter{{Column: "u.id", Column2: "c.author_id"}}))
}
