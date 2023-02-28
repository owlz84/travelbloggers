package main

import (
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/julienschmidt/httprouter"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	Name   string    `form:"name"`
	Email  string    `form:"email"`
	Joined time.Time `form:"joined"`
}

type accountPasswordUpdateForm struct {
	CurrentPassword         string `form:"currentPassword"`
	NewPassword             string `form:"newPassword"`
	NewPasswordConfirmation string `form:"newPasswordConfirmation"`
	validator.Validator     `form:"-"`
}

type blogViewForm struct {
	BlogID         int
	BlogName       string
	OwnerID        int
	OwnerName      string
	ViewerIsAuthor bool
}
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
	DateFrom            time.Time `form:"datefrom"`
	DateTo              time.Time `form:"dateto"`
	validator.Validator `form:"-"`
}

type ImageDetail struct {
	ImagePath     string
	ThumbnailPath string
}

type imageUploadForm struct {
	Images []*ImageDetail
}

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
		"./ui/html/base.tmpl",
		"./ui/html/pages/home.tmpl",
		"./ui/html/partials/nav.tmpl",
	)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/base.tmpl",
		"./ui/html/pages/about.tmpl",
		"./ui/html/partials/nav.tmpl",
	)
	return
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/base.tmpl",
		"./ui/html/pages/signup.tmpl",
		"./ui/html/partials/nav.tmpl",
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
			"./ui/html/base.tmpl",
			"./ui/html/pages/signup.tmpl",
			"./ui/html/partials/nav.tmpl",
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
				"./ui/html/base.tmpl",
				"./ui/html/pages/signup.tmpl",
				"./ui/html/partials/nav.tmpl",
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
		"./ui/html/base.tmpl",
		"./ui/html/pages/login.tmpl",
		"./ui/html/partials/nav.tmpl",
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
			"./ui/html/base.tmpl",
			"./ui/html/pages/login.tmpl",
			"./ui/html/partials/nav.tmpl",
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
				"./ui/html/base.tmpl",
				"./ui/html/pages/login.tmpl",
				"./ui/html/partials/nav.tmpl",
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
	data := app.newTemplateData(r)
	data.Form = accountViewForm{
		Name:   user.Name,
		Email:  user.Email,
		Joined: user.Created,
	}
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/base.tmpl",
		"./ui/html/pages/account.tmpl",
		"./ui/html/partials/nav.tmpl",
	)
	//fmt.Fprintf(w, "%+v", user)
}

func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountPasswordUpdateForm{}
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/base.tmpl",
		"./ui/html/pages/password.tmpl",
		"./ui/html/partials/nav.tmpl",
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
			"./ui/html/base.tmpl",
			"./ui/html/pages/password.tmpl",
			"./ui/html/partials/nav.tmpl",
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
				"./ui/html/base.tmpl",
				"./ui/html/pages/password.tmpl",
				"./ui/html/partials/nav.tmpl",
			)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your password was changed successfully!")
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
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
		"./ui/html/base.tmpl",
		"./ui/html/pages/createPost.tmpl",
		"./ui/html/partials/nav.tmpl",
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
			"./ui/html/base.tmpl",
			"./ui/html/pages/createPost.tmpl",
			"./ui/html/partials/nav.tmpl",
		)
		return
	}

	err = app.posts.Insert(form.BlogID, form.Title, form.Content, form.Country, form.DateFrom, form.DateTo)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Post successfully published!")

	http.Redirect(w, r, fmt.Sprintf("/blogs/view/%d", form.BlogID), http.StatusSeeOther)
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
		"./ui/html/base.tmpl",
		"./ui/html/pages/blog.tmpl",
		"./ui/html/partials/nav.tmpl",
	)
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
		"./ui/html/base.tmpl",
		"./ui/html/pages/viewPost.tmpl",
		"./ui/html/partials/nav.tmpl",
	)
}

func (app *application) uploadImage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/base.tmpl",
		"./ui/html/pages/uploadImages.tmpl",
		"./ui/html/partials/nav.tmpl",
	)

}

func (app *application) uploadImagePost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(1024 << 20)
	if err != nil {
		app.serverError(w, err)
		return
	} // 32MB is the default used by FormFile
	fhs := r.MultipartForm.File["img"]
	var form imageUploadForm
	for _, fh := range fhs {
		f, err := fh.Open()
		if err != nil {
			app.serverError(w, err)
			return
		}
		imageRoot := "./ui/static/user-img"
		tempFile, err := os.CreateTemp(imageRoot, "upload-*.png")
		if err != nil {
			app.serverError(w, err)
			return
		}
		fileName := tempFile.Name()

		fileBytes, err := io.ReadAll(f)
		if err != nil {
			fmt.Println(err)
		}
		tempFile.Write(fileBytes)
		tempFile.Close()

		var thumbnail image.Image
		img, err := imaging.Open(fileName)
		if err != nil {
			app.serverError(w, err)
			return
		}

		thumbnail = imaging.Thumbnail(img, 256, 256, imaging.CatmullRom)
		dst := imaging.New(256, 256, color.NRGBA{0, 0, 0, 0})
		dst = imaging.Paste(dst, thumbnail, image.Pt(0, 0))
		thumbnailPath := strings.ReplaceAll(fileName, "user-img", "user-img/thumbs")
		err = imaging.Save(dst, thumbnailPath)
		if err != nil {
			app.serverError(w, err)
			return
		}

		image := ImageDetail{
			ImagePath:     strings.ReplaceAll(fileName, "./ui", ""),
			ThumbnailPath: strings.ReplaceAll(thumbnailPath, "./ui", "")}
		form.Images = append(form.Images, &image)

		f.Close()
	}

	data := app.newTemplateData(r)
	data.Form = form
	app.render(
		w,
		http.StatusOK,
		data,
		"./ui/html/base.tmpl",
		"./ui/html/pages/uploadImages.tmpl",
		"./ui/html/partials/nav.tmpl",
	)

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
		"./ui/html/base.tmpl",
		"./ui/html/pages/editPost.tmpl",
		"./ui/html/partials/nav.tmpl",
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
			"./ui/html/base.tmpl",
			"./ui/html/pages/editPost.tmpl",
			"./ui/html/partials/nav.tmpl",
		)
		return
	}

	err = app.posts.Update(form.PostID, form.Title, form.Content, form.Country, form.DateFrom, form.DateTo)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Post successfully updated.")

	http.Redirect(w, r, fmt.Sprintf("/posts/view/%d", form.PostID), http.StatusSeeOther)

}
