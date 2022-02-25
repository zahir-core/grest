package db

import (
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestQuery(t *testing.T) {
	db, err := NewMockDB()
	if err != nil {
		t.Fatalf("Error occured : [%v]", err.Error())
	}
	articles := []Article{}
	q := url.Values{}
	q.Add("$or", "author.name.$ilike=john||is_active=true")
	q.Add("detail.path.to.detail.$like", "some detail")
	q.Add("$sort", "author.name,-detail.path.to.detail:i,title:i,-updated_at")
	q.Add("$select", "$max:title")
	q.Add("$search", "title,content,author.name=john")
	Find(db, &articles, q)
}

func NewMockDB() (*gorm.DB, error) {
	sqlDB, _, _ := sqlmock.New()

	c := gorm.Config{}
	c.PrepareStmt = false
	c.DryRun = true
	c.SkipDefaultTransaction = true
	c.AllowGlobalUpdate = false
	c.Logger = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		Colorful: true,
		LogLevel: logger.Info,
	})

	return gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &c)
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
	IsActive    NullBool     `json:"is_active"    db:"a.is_active"`
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

func (c *Category) SetRelation() {
	c.Relation = append(c.Relation, NewRelation("inner", "articles_categories", "ac", []Filter{{Column: "ac.category_id", Column2: "c.id"}}))
	c.Relation = append(c.Relation, NewRelation("left", "users", "u", []Filter{{Column: "u.id", Column2: "c.author_id"}}))
}
