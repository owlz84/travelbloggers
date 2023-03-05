package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"travelbloggers/internal/validator"
)

type blogCreateForm struct {
	ID                  int
	OwnerID             int
	Name                string `form:"name"`
	validator.Validator `form:"-"`
}

type blogViewForm struct {
	BlogID         int
	BlogName       string
	OwnerID        int
	OwnerName      string
	ViewerIsAuthor bool
}

func (app *application) blogCreate(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	currentUserID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	data.Form = blogCreateForm{
		OwnerID: currentUserID,
		Name:    "",
	}

	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/createBlog.tmpl",
	)

	return
}

func (app *application) blogCreatePost(w http.ResponseWriter, r *http.Request) {
	var form blogCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "'Name' field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, data, "./ui/html/pages/createBlog.tmpl")
		return
	}

	currentUserID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	id, err := app.blogs.Insert(currentUserID, form.Name)

	app.sessionManager.Put(r.Context(), "flash", "Blog successfully published!")
	http.Redirect(w, r, fmt.Sprintf("/blogs/view/%d", id), http.StatusSeeOther)

}

func (app *application) blogView(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	blogId, err := strconv.Atoi(params.ByName("blogid"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	data := app.newTemplateData(r)
	blog, err := app.blogs.Get(blogId)
	user, err := app.users.Get(blog.OwnerID)
	currentUserID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	form := blogViewForm{
		blog.ID,
		blog.Name,
		user.ID,
		user.Name,
		currentUserID == user.ID,
	}
	data.Form = form

	posts, err := app.posts.GetByBlog(blogId)
	//app.posts.RenderPostsAsHTML(&posts)

	data.Posts = posts
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/viewBlog.tmpl",
	)
}

func (app *application) blogDeletePost(w http.ResponseWriter, r *http.Request) {

}
