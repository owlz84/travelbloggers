package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	router.Handler(http.MethodGet, "/blogs/view/:blogid", dynamic.ThenFunc(app.blogView))
	router.Handler(http.MethodGet, "/posts/view/:postid", dynamic.ThenFunc(app.postView))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
	router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.accountView))
	router.Handler(http.MethodGet, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdate))
	router.Handler(http.MethodPost, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdatePost))
	router.Handler(http.MethodGet, "/blogs/create", protected.ThenFunc(app.blogCreate))
	router.Handler(http.MethodPost, "/blogs/create", protected.ThenFunc(app.blogCreatePost))
	router.Handler(http.MethodPost, "/blogs/delete", protected.ThenFunc(app.blogDeletePost))
	router.Handler(http.MethodGet, "/posts/create", protected.ThenFunc(app.postPublish))
	router.Handler(http.MethodPost, "/posts/create", protected.ThenFunc(app.postPublishPost))
	router.Handler(http.MethodGet, "/posts/edit/:id", protected.ThenFunc(app.postEdit))
	router.Handler(http.MethodPost, "/posts/edit/:id", protected.ThenFunc(app.postEditPost))
	router.Handler(http.MethodGet, "/images/upload", protected.ThenFunc(app.uploadImage))
	router.Handler(http.MethodPost, "/images/upload", protected.ThenFunc(app.uploadImagePost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
