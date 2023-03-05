package main

import (
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	posts, err := app.posts.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, post := range posts {
		blog, err := app.blogs.Get(post.BlogID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		post.BlogName = blog.Name
		post.UserID = blog.OwnerID
		user, err := app.users.Get(post.UserID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		post.UserName = user.Name
	}
	//app.posts.RenderPostsAsHTML(&posts)
	data := app.newTemplateData(r)
	data.Posts = posts
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/home.tmpl",
	)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/about.tmpl",
	)
	return
}
