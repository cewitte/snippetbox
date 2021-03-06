package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

// Update the signature for the routes() method so that it returns a http.Handler instead of a *http.ServeMux.
func (app *application) routes() http.Handler {
	// Create a middleware chain containing our "standard" middleware which will be used for every request our application receives.
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Create a new middleware chain containing the middleware specific to our dynamic application routes. For now, this chain will only contain the session middleware, but we'll add more to it later.
	// Use the nosurf middleware on all our "dynamic" routes
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	mux := pat.New()
	// Update these routes to use the new dynamic middleware chain followed by the appropriate handler function.
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	// Add the requireAuthentication middleware to the chain.
	mux.Get("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet)) // Moved down.

	// Add the five new routes for user authentication.
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	// Add the requiredAuthentication middleware to the chain.
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))

	// Create a file server which servers files out of the "./ui/static" directory.
	// Note that the path given to the http.Dir function is relative to the project directory root
	// The neutered file system prevents directory listing as described here: https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static")})
	// fileServer := http.FileServer(http.Dir("./ui/static"))

	// Use the mux.Handle() function to register the file server as the handler for all URL paths that start with "/static/". For matching paths, we strip the "/static" prefix before the request reaches the file server.
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// Return the "standard" middleware chain followed by the servemix.
	return standardMiddleware.Then(mux)
}
