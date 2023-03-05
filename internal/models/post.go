package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Post struct {
	ID       int
	BlogID   int
	BlogName string
	UserID   int
	UserName string
	Title    string
	Content  string
	Country  string
	Location string
	DateFrom time.Time
	DateTo   time.Time
	Created  time.Time
}

type PostModelInterface interface {
	Get(id int) (*Post, error)
	Insert(blogID int, title, content, country, location string, dateFrom, dateTo time.Time) error
	Update(postID int, title, content, country, location string, dateFrom, dateTo time.Time) error
	GetByBlog(blogID int) ([]*Post, error)
	Latest() ([]*Post, error)
}

func (p *PostModel) GetByBlog(blogID int) ([]*Post, error) {
	posts := []*Post{}
	stmt := `SELECT id, blog_id, title, content, country, location, date_from, date_to, created
		FROM posts
		WHERE blog_id = ?
		ORDER BY date_to DESC`
	rows, err := p.DB.Query(stmt, blogID)
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(
			&post.ID, &post.BlogID, &post.Title,
			&post.Content, &post.Country, &post.Location,
			&post.DateFrom, &post.DateTo, &post.Created,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNoRecord
			} else {
				return nil, err
			}
		}
		posts = append(posts, post)
	}
	return posts, nil
}

type PostModel struct {
	DB *sql.DB
}

func (p *PostModel) Insert(blogID int, title, content, country, location string, dateFrom, dateTo time.Time) error {
	stmt := `INSERT INTO posts (blog_id, title, content, country, location, date_from, date_to, created)
    VALUES(?, ?, ?, ?, ?, ?, ?, UTC_TIMESTAMP())`

	content = strings.ReplaceAll(content, "\r", "")

	_, err := p.DB.Exec(stmt, blogID, title, content, country, location, dateFrom, dateTo)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostModel) Get(id int) (*Post, error) {
	post := &Post{}
	stmt := `SELECT id, blog_id, title, content, country, location, date_from, date_to, created
		FROM posts
		WHERE id = ?`
	err := p.DB.QueryRow(stmt, id).Scan(
		&post.ID, &post.BlogID, &post.Title,
		&post.Content, &post.Country, &post.Location,
		&post.DateFrom, &post.DateTo, &post.Created,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return post, nil
}

func (p *PostModel) Update(postID int, title, content, country, location string, dateFrom, dateTo time.Time) error {
	stmt := `UPDATE posts
SET title=?, content=?, country=?, location=?, date_from=?, date_to=?
WHERE id=?`

	content = strings.ReplaceAll(content, "\r", "")

	_, err := p.DB.Exec(stmt, title, content, country, location, dateFrom, dateTo, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRecord
		} else {
			return err
		}
	}
	return nil
}

func (p *PostModel) Latest() ([]*Post, error) {
	posts := []*Post{}
	numEntries := 10
	stmt := `SELECT id, blog_id, title, country, location, created
		FROM posts
		ORDER BY created DESC
		LIMIT ?`
	rows, err := p.DB.Query(stmt, numEntries)
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(
			&post.ID, &post.BlogID, &post.Title,
			&post.Country, &post.Location, &post.Created,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNoRecord
			} else {
				return nil, err
			}
		}
		posts = append(posts, post)
	}
	return posts, nil
}
