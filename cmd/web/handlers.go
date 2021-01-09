package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cewitte/snippetbox/pkg/forms"
	"github.com/cewitte/snippetbox/pkg/models"
)

// Change the signatura of the home handler so it is defined as a method against *application
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Because Pat matches the "/" path exactly, we can now remove the manual check of r.URL.Path != "/"

	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the new render helper.
	app.render(w, r, "home.page.tmpl", &templateData{
		Snippets: s,
	})
}

// Change the signatura of the showSnippet handler so it is defined as a method against *application
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// Pat doesn't strip the colon from the named capture key, so we need to get the value of :id from the query string instead of ":id"
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w) // use the notFound() helper.
		return
	}

	// Use the SnippetModel object's Get method to retrieve the data for a specific record based on its ID. If no matching record is found, return a 404 Not Found response.
	s, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}

		return
	}

	// Use the new render helper.
	// Pass the flash message to the template.
	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})
}

// Add a new createSnippetForm handler, which for now returns a placeholder response.
func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		// Pass a new empty forms.Form object to the template.
		Form: forms.New(nil),
	})
}

// Change the signatura of the createSnippet handler so it is defined as a method against *application
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	// First we call r.ParseForm() which adds any data in POST request bodies to the r.PostForm map. This also works in the same wat for PUT and PATCH requests. If there are any errors, we use our app.ClientError helper to send a 400 Bad Request response to the user.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Create a new forms.Form struct containing the POSTed data from the form, the use the validation methods to check the content.
	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	// If the form isn't valid, redisplay the template passing in the form.Form object as the data.
	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{
			Form: form,
		})
		return
	}

	// Because the form data (with type url.Values) has been anonymously embedded in the form.Form struct, we can use the Get() method to retrieve the validated value for a particular form field.
	id, err := app.snippets.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the Put() method to add a string value ("Your snippet was saved successfully!") and the corresponding key ("flash") to the session data. Note that if there's no existing session for the current user (or their session has expired) the a new, empty session for them will be automatically create by the session middleware.
	app.session.Put(r, "flash", "Snippet successfully created!")

	// Redirect the user to the relevant page for the snippet
	// Change the redirect to use the new semantic URL style of /snippet/:id
	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Display the user signup form")
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Create a new user...")
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Display the user login form...")
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Authenticate and login the user...")
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Logout the user...")
}

