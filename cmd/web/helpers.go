package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/tz"
	"html/template"
	"net/http"
	"runtime/debug"
	"time"
)

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
	}
}

func (app *application) render(w http.ResponseWriter, status int, data *templateData, files ...string) {

	ts := template.New("").Funcs(functions)
	ts, err := ts.ParseFiles(files...) //template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	buf := new(bytes.Buffer)

	w.WriteHeader(status)

	// Execute the template set and write the response body. Again, if there
	// is any error we call the serverError() helper.
	err = ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	if err != nil {
		return
	}
}

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) decodePostForm(r *http.Request, dst any) error {

	app.formDecoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	err := r.ParseForm()
	if err != nil {
		return err
	}
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		return err
	}

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

func (app *application) getCountries() []string {
	var countries []string
	for _, country := range tz.GetCountries() {
		countries = append(countries, country.Name)
	}
	return countries
}
