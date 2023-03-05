package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
	"travelbloggers/internal/models"
	"travelbloggers/internal/validator"
)

type postViewForm struct {
	PostID         int
	BlogID         int
	BlogName       string
	OwnerID        int
	OwnerName      string
	ViewerIsAuthor bool
}

type postPublishForm struct {
	BlogID              int            `form:"blog_id"`
	Title               string         `form:"title"`
	Content             string         `form:"content"`
	Country             string         `form:"country"`
	CountryList         []string       `form:"-"`
	Location            string         `form:"location"`
	DateFrom            time.Time      `form:"datefrom"`
	DateTo              time.Time      `form:"dateto"`
	Blogs               []*models.Blog `form:"-"`
	validator.Validator `form:"-"`
}

type postEditForm struct {
	PostID              int       `form:"post_id"`
	Title               string    `form:"title"`
	Content             string    `form:"content"`
	Country             string    `form:"country"`
	CountryList         []string  `form:"-"`
	Location            string    `form:"location"`
	DateFrom            time.Time `form:"datefrom"`
	DateTo              time.Time `form:"dateto"`
	validator.Validator `form:"-"`
}

func (app *application) postView(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	postId, err := strconv.Atoi(params.ByName("postid"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	data := app.newTemplateData(r)
	post, err := app.posts.Get(postId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	blog, err := app.blogs.Get(post.BlogID)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	user, err := app.users.Get(blog.OwnerID)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	currentUserID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	form := postViewForm{
		post.ID,
		blog.ID,
		blog.Name,
		user.ID,
		user.Name,
		currentUserID == user.ID,
	}
	data.Form = form

	data.Post = post
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/viewPost.tmpl",
	)
}

func (app *application) postPublish(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	form := postPublishForm{}
	form.CountryList = app.getCountries()
	form.DateTo = time.Now()
	form.DateFrom = time.Now()

	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	blogs, err := app.blogs.GetByOwner(userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Blogs = blogs

	data.Form = form
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/createPost.tmpl",
	)
}

func (app *application) postPublishPost(w http.ResponseWriter, r *http.Request) {

	var form postPublishForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CountryList = app.getCountries()

	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	blogs, err := app.blogs.GetByOwner(userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.Blogs = blogs

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Country), "country", "You must select a country.")
	form.CheckField(validator.NotBlank(form.Location), "location", "You must specify a location.")
	form.CheckField(validator.NotEmptyDate(form.DateFrom), "datefrom", "You must select a from date.")
	form.CheckField(validator.NotEmptyDate(form.DateTo), "dateto", "You must select a to date.")
	form.CheckField(validator.TimeBeforeOrEqual(form.DateFrom, form.DateTo), "dateto", "To date must be after from date.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(
			w,
			http.StatusUnprocessableEntity,
			data,
			"./ui/html/pages/createPost.tmpl",
		)
		return
	}

	err = app.posts.Insert(
		form.BlogID,
		form.Title,
		form.Content,
		form.Country,
		form.Location,
		form.DateFrom,
		form.DateTo,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Post successfully published!")

	http.Redirect(w, r, fmt.Sprintf("/blogs/view/%d", form.BlogID), http.StatusSeeOther)
}

func (app *application) postEdit(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	postId, err := strconv.Atoi(params.ByName("id"))
	post, err := app.posts.Get(postId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := postEditForm{
		PostID:      post.ID,
		Title:       post.Title,
		Country:     post.Country,
		Location:    post.Location,
		DateFrom:    post.DateFrom,
		DateTo:      post.DateTo,
		Content:     post.Content,
		CountryList: app.getCountries(),
	}

	data := app.newTemplateData(r)
	data.Form = form
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/editPost.tmpl",
	)

}

func (app *application) postEditPost(w http.ResponseWriter, r *http.Request) {

	var form postEditForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Country), "country", "You must select a country.")
	form.CheckField(validator.NotBlank(form.Location), "location", "You must specify a location.")
	form.CheckField(validator.NotEmptyDate(form.DateFrom), "datefrom", "You must select a from date.")
	form.CheckField(validator.NotEmptyDate(form.DateTo), "dateto", "You must select a to date.")
	form.CheckField(validator.TimeBeforeOrEqual(form.DateFrom, form.DateTo), "dateto", "'To' date must be after (or the same as) 'From' date.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(
			w,
			http.StatusUnprocessableEntity,
			data,
			"./ui/html/pages/editPost.tmpl",
		)
		return
	}

	err = app.posts.Update(
		form.PostID,
		form.Title,
		form.Content,
		form.Country,
		form.Location,
		form.DateFrom,
		form.DateTo,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Post successfully updated.")

	http.Redirect(w, r, fmt.Sprintf("/posts/view/%d", form.PostID), http.StatusSeeOther)

}
