package models

import (
	"database/sql"
	"errors"
)

type Blog struct {
	ID      int
	OwnerID int
	Name    string
}

type BlogModelInterface interface {
	Insert(ownerID int, name string) (int, error)
	Get(id int) (*Blog, error)
	GetByOwner(id int) ([]*Blog, error)
}

type BlogModel struct {
	DB *sql.DB
}

func (b *BlogModel) Insert(ownerID int, name string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlogModel) Get(id int) (*Blog, error) {
	blog := &Blog{}
	stmt := "SELECT id, owner_id, name FROM blogs WHERE id = ?"
	err := b.DB.QueryRow(stmt, id).Scan(&blog.ID, &blog.OwnerID, &blog.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return blog, nil
}

func (b *BlogModel) GetByOwner(id int) ([]*Blog, error) {
	blogs := []*Blog{}
	stmt := `SELECT id, owner_id, name
		FROM blogs
		WHERE owner_id = ?`
	rows, err := b.DB.Query(stmt, id)
	for rows.Next() {
		blog := &Blog{}
		err = rows.Scan(&blog.ID, &blog.OwnerID, &blog.Name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNoRecord
			} else {
				return nil, err
			}
		}
		blogs = append(blogs, blog)
	}
	return blogs, nil

}
