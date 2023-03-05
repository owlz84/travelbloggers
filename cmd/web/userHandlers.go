package main

import (
	"errors"
	"net/http"
	"time"
	"travelbloggers/internal/models"
	"travelbloggers/internal/validator"
)

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type accountViewForm struct {
	Name   string         `form:"name"`
	Email  string         `form:"email"`
	Joined time.Time      `form:"joined"`
	Blogs  []*models.Blog `form:"-"`
}

type accountPasswordUpdateForm struct {
	CurrentPassword         string `form:"currentPassword"`
	NewPassword             string `form:"newPassword"`
	NewPasswordConfirmation string `form:"newPasswordConfirmation"`
	validator.Validator     `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/signup.tmpl",
	)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using our helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(
			w,
			http.StatusUnprocessableEntity,
			data,
			"./ui/html/pages/signup.tmpl",
		)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(
				w,
				http.StatusUnprocessableEntity,
				data,
				"./ui/html/pages/signup.tmpl",
			)
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/login.tmpl",
	)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(
			w,
			http.StatusUnprocessableEntity,
			data,
			"./ui/html/pages/login.tmpl",
		)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(
				w,
				http.StatusUnprocessableEntity,
				data,
				"./ui/html/pages/login.tmpl",
			)
		} else {
			app.serverError(w, err)
		}
		return
	}
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	redirectPathLogin := app.sessionManager.GetString(r.Context(), "redirectPathLogin")
	if redirectPathLogin != "" {
		http.Redirect(w, r, redirectPathLogin, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/post/create", http.StatusSeeOther)

}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(id)
	if user == nil {

		http.Redirect(w, r, "/post/create", http.StatusSeeOther)
	}
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
	}
	blogs, err := app.blogs.GetByOwner(id)
	if err != nil {
		app.serverError(w, err)
	}
	data := app.newTemplateData(r)
	data.Form = accountViewForm{
		Name:   user.Name,
		Email:  user.Email,
		Joined: user.Created,
		Blogs:  blogs,
	}
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/account.tmpl",
	)
}

func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountPasswordUpdateForm{}
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/pages/password.tmpl",
	)
}

func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	form := accountPasswordUpdateForm{}
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirmation), "newPasswordConfirmation", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "New password must be at least 8 characters long")
	form.CheckField(validator.IsEqual(form.NewPassword, form.NewPasswordConfirmation), "newPasswordConfirmation", "New passwords must match")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(
			w,
			http.StatusUnprocessableEntity,
			data,
			"./ui/html/pages/password.tmpl",
		)
		return
	}
	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	err = app.users.PasswordUpdate(id, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Current password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(
				w,
				http.StatusUnprocessableEntity,
				data,
				"./ui/html/pages/password.tmpl",
			)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your password was changed successfully!")
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}
