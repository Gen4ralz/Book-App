package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"vue-api/internal/data"

	"github.com/go-chi/chi/v5"
	"github.com/mozillazg/go-slugify"
)

var staticPath = "./static/"

type jsonResponse struct {
	Error bool		`json:"error"`
	Message string 	`json:"message"`
	Data interface{}`json:"data,omitempty"`
}

type envelope map[string]interface{}

func (app *application) Login(res http.ResponseWriter,req *http.Request) {
	type credentials struct {
		UserName string		`json:"email"`
		PassWord string		`json:"password"`
	}

	var creds credentials
	var payload jsonResponse

	err := app.readJSON(res,req, &creds)
	if err != nil {
		app.errorLog.Println(err)
		payload.Error = true
		payload.Message = "invalid json supplied, or json missing entirely"
		_ = app.writeJSON(res, http.StatusBadRequest, payload)
	}

	// authenticate
	// app.infoLog.Println(creds.UserName, creds.PassWord)

	// look up the user by email
	user, err := app.models.User.GetByEmail(creds.UserName)
	if err != nil {
		app.errorJSON(res, errors.New("invalid username/password"))
		return
	}

	// validate the user's password
	validPassword, err := user.PasswordMatches(creds.PassWord)
	if err != nil || !validPassword {
		app.errorJSON(res, errors.New("invalid username/password"))
		return
	}

	// make sure user is active
	if user.Active == 0 {
		app.errorJSON(res, errors.New("user is not active"))
		return
	}

	// we have a valid user, so generate a token
	token, err := app.models.Token.GenerateToken(user.ID, 24*time.Hour) 
	if err != nil {
		app.errorJSON(res, err)
		return
	}
	// save it to the database
	err = app.models.Token.Insert(*token, *user)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	// send back a response
	payload = jsonResponse{
		Error: false,
		Message: "logged in",
		Data: envelope{"token": token, "user": user},
	}

	// out,err := json.MarshalIndent(payload,"","\t")
	err = app.writeJSON(res, http.StatusOK, payload)
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) Logout(res http.ResponseWriter, req *http.Request) {
	var requestPayload struct{
		Token string	`json:"token"`
	}

	err := app.readJSON(res, req, &requestPayload)
	if err != nil {
		app.errorJSON(res, errors.New("invalid json"))
		return
	}
	err = app.models.Token.DeleteByToken(requestPayload.Token)
	if err != nil {
		app.errorJSON(res, errors.New("invalid json"))
		return
	}

	payload := jsonResponse{
		Error: false,
		Message: "logged out",
	}

	_ = app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) AllUsers(res http.ResponseWriter, req *http.Request){
	var users data.User
	all,err := users.GetAll()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	payload := jsonResponse{
		Error: false,
		Message: "success",
		Data: envelope{"users":all},
	}
	app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) EditUser(res http.ResponseWriter,req *http.Request) {
	var user data.User

	err := app.readJSON(res, req, &user)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	if user.ID == 0 {
		// add user
		if _, err := app.models.User.Insert(user); err != nil {
			app.errorJSON(res, err)
			return
		}
	} else {
		// editing user
		u, err := app.models.User.GetByID(user.ID)
		if err != nil {
			app.errorJSON(res, err)
			return
		}

		u.Email = user.Email
		u.FirstName = user.FirstName
		u.LastName = user.LastName
		u.Active = user.Active

		if err := u.Update(); err != nil {
			app.errorJSON(res, err)
			return
		}

		// if password != string, update password
		if user.Password != "" {
			err := u.ResetPassword(user.Password)
			if err != nil {
				app.errorJSON(res, err)
				return
			}
		}
	}

	payload := jsonResponse{
		Error: false,
		Message: "Changes saved",
	}

	_ = app.writeJSON(res, http.StatusAccepted, payload)
}

func (app *application) GetUser(res http.ResponseWriter, req *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	user,err := app.models.User.GetByID(userID)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	_ = app.writeJSON(res, http.StatusOK, user)
}

func (app *application) DeleteUser(res http.ResponseWriter, req *http.Request) {
	var requestPayload struct {
		ID int `json:"id"`
	}

	err := app.readJSON(res, req, &requestPayload)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	err = app.models.User.DeleteByID(requestPayload.ID)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	payload := jsonResponse{
		Error: false,
		Message: "User deleted",
	}

	_ = app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) LogUserOutAndSetInactive(res http.ResponseWriter, req *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	user, err := app.models.User.GetByID(userID)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	user.Active = 0
	err = user.Update()
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	// delete tokens for user
	err = app.models.Token.DeleteTokensForUser(userID)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	payload := jsonResponse {
		Error: false,
		Message: "user logged out and set to inactive",
	}

	_ = app.writeJSON(res, http.StatusAccepted, payload)
}

func (app *application) ValidateToken(res http.ResponseWriter, req *http.Request) {
	var requestPayload struct {
		Token string `json:"token"`
	}

	err := app.readJSON(res, req, &requestPayload)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	valid := false
	valid, _ = app.models.Token.ValidToken(requestPayload.Token)

	payload := jsonResponse {
		Error: false,
		Data: valid,
	}

	_ = app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) AllBooks(res http.ResponseWriter, req *http.Request) {
	books, err := app.models.Book.GetAll()
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	payload := jsonResponse {
		Error: false,
		Message: "success",
		Data: envelope{"books": books},
	}

	app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) OneBook(res http.ResponseWriter, req *http.Request) {
	slug := chi.URLParam(req, "slug")

	book, err := app.models.Book.GetOneBySlug(slug)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	payload := jsonResponse {
		Error: false,
		Data: book,
	}

	app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) AuthorsAll(res http.ResponseWriter, req *http.Request) {
	all, err := app.models.Author.All()
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	type selectData struct{
		Value int		`json:"value"`
		Text string		`json:"text"`
	}

	var results []selectData

	for _, x := range all {
		author := selectData{
			Value: x.ID,
			Text: x.AuthorName,
		}

		results = append(results, author)
	}

	payload := jsonResponse {
		Error: false,
		Data: results,
	}
	app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) EditBook(res http.ResponseWriter, req *http.Request) {
	var requestPayload struct {
		ID 				int		`json:"id"`
		Title 			string	`json:"title"`
		AuthorID 		int		`json:"author_id"`
		PublicationYear int		`json:"publication_year"`
		Description		string	`json:"description"`
		CoverBase64		string	`json:"cover"`
		GenresIDs		[]int	`json:"genre_ids"`
	}

	err := app.readJSON(res, req, &requestPayload)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	book := data.Book{
		ID: requestPayload.ID,
		Title: requestPayload.Title,
		AuthorID: requestPayload.AuthorID,
		PublicationYear: requestPayload.PublicationYear,
		Description: requestPayload.Description,
		Slug: slugify.Slugify(requestPayload.Title),
		GenreIDs: requestPayload.GenresIDs,
	}

	if len(requestPayload.CoverBase64) > 0 {
		// we have a cover
		decoded, err := base64.StdEncoding.DecodeString(requestPayload.CoverBase64)
		if err != nil {
			app.errorJSON(res, err)
			return
		}

		// write image to /static/covers
		if err := os.WriteFile(fmt.Sprintf("%s/covers/%s.jpg", staticPath, book.Slug), decoded, 0666); err != nil {
			app.errorJSON(res, err)
			return
		}
	}

	if book.ID == 0 {
		// adding a book
		_,err := app.models.Book.Insert(book)
		if err != nil {
			app.errorJSON(res, err)
			return
		}
	} else {
		// update a book
		err := book.Update()
		if err != nil {
			app.errorJSON(res, err)
			return
		}
	}

	payload := jsonResponse {
		Error: false,
		Message: "Changes saved",
	}

	app.writeJSON(res, http.StatusAccepted, payload)
}

func (app *application) BookByID(res http.ResponseWriter, req *http.Request) {
	bookID, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	book, err := app.models.Book.GetOneById(bookID)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	payload := jsonResponse{
		Error: false,
		Data: book,
	}

	app.writeJSON(res, http.StatusOK, payload)
}

func (app *application) DeleteBook(res http.ResponseWriter, req *http.Request) {
	var requestPayload struct {
		ID int	`json:"id"`
	}

	err := app.readJSON(res, req, &requestPayload)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	err = app.models.Book.DeleteByID(requestPayload.ID)
	if err != nil {
		app.errorJSON(res, err)
		return
	}

	payload := jsonResponse{
		Error: false,
		Message: "Book deleted",
	}

	app.writeJSON(res, http.StatusOK, payload)
}