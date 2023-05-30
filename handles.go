package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-pg/pg/v10"
)

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Started CreateAccount")
	creds := &User{}
	jsErr := json.NewDecoder(r.Body).Decode(creds)
	if jsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := Open()
	defer db.Close()
	ctx := context.Background()

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}

	queryErr := db.Model(creds).Column("username").Where("?=?", pg.Ident("username"), creds.Username).Select()
	if queryErr != nil {
		hashedPaswd, _ := passwordHash(creds.PasswordHash)
		dbCred := &User{
			Username:     creds.Username,
			PasswordHash: string(hashedPaswd),
		}

		_, insErr := db.Model(dbCred).Insert()
		if insErr != nil {
			w.Write([]byte("unfortunately we couldn't create your account. Please try again."))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(201)
		return
	}

	w.Write([]byte("account already exists. please login"))
	w.WriteHeader(http.StatusOK)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	creds := &User{}
	json.NewDecoder(r.Body).Decode(creds)

	if creds.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	Password := creds.PasswordHash
	if Password == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	dbCred := &User{}
	dbConn := Open()
	defer dbConn.Close()

	ctx := context.Background()

	if err := dbConn.Ping(ctx); err != nil {
		panic(err)
	}

	//scan into dbcred from database
	SelectErr := dbConn.Model(dbCred).Column("passwordhash", "id").Where("username=?", creds.Username).Select(&dbCred.PasswordHash, &dbCred.ID)
	if SelectErr != nil {
		w.Write([]byte("please enter a valid username"))
		return
	}

	pswdValErr := validateHash(Password, dbCred.PasswordHash)
	if pswdValErr != nil {
		w.Write([]byte("incorrect password"))
		return
	}
	JWTtoken, err := GenerateJWT(dbCred.ID, creds.Username)
	if err != nil {
		w.Write([]byte("could not authenticate user")) //why here?
		return
	}

	w.Header().Add("Authorization", "Bearer"+JWTtoken)
}

func addBook(w http.ResponseWriter, r *http.Request) {
	dbConn := Open()
	defer dbConn.Close()
	userIDStr := r.Context().Value(uuid).(string)
	userID, _ := strconv.Atoi(userIDStr)

	switch {
	case strings.Contains(r.URL.Path, "/read"):
		book := &UserRequest{}
		readErr := json.NewDecoder(r.Body).Decode(book)
		if readErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if book.Bookname == "" || book.Bookname == " " {
			w.WriteHeader(http.StatusNoContent)
		}

		bookInfo := &Book{
			UserID:   userID,
			Bookname: book.Bookname,
		}
		fmt.Println(bookInfo)
		_, insErr := dbConn.Model(bookInfo).Insert()
		if insErr != nil {
			w.Write([]byte("sorry! we were unable to add your read book" + insErr.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	case strings.Contains(r.URL.Path, "/unread"):
		book := &UserRequest{}
		readErr := json.NewDecoder(r.Body).Decode(book)
		if readErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if book.Bookname == "" || book.Bookname == " " {
			w.WriteHeader(http.StatusNoContent)
		}

		bookInfo := &Book{
			UserID:   userID,
			Bookname: book.Bookname,
		}
		_, insErr := dbConn.Model(bookInfo).Insert()
		if insErr != nil {
			w.Write([]byte("sorry! we were unable to add your unread book"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	dbConn := Open()
	defer dbConn.Close()
	userIDStr := r.Context().Value(uuid).(string)
	userID, _ := strconv.Atoi(userIDStr)

	switch {
	case strings.Contains(r.URL.Path, "/read"):
		var books []Book
		queryErr := dbConn.Model(&books).Column("bookname").Where("?=?", pg.Ident("user_id"), userID).Select()
		if queryErr != nil {
			w.Write([]byte("unable to list read books"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := json.NewEncoder(w).Encode(books)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusAccepted)
	case strings.Contains(r.URL.Path, "/unread"):
		var books []Book
		queryErr := dbConn.Model(&books).Column("bookname").Where("?=?", pg.Ident("user_id"), userID).Select()
		if queryErr != nil {
			w.Write([]byte("unable to list unread books"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err := json.NewEncoder(w).Encode(books)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
func removeBook(w http.ResponseWriter, r *http.Request) {
	dbConn := Open()
	defer dbConn.Close()
	userIDStr := r.Context().Value(uuid).(string)
	userID, _ := strconv.Atoi(userIDStr)

	switch {
	case strings.Contains(r.URL.Path, "/read"):
		book := &UserRequest{}
		readErr := json.NewDecoder(r.Body).Decode(book)
		if readErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if book.Bookname == "" || book.Bookname == " " {
			w.WriteHeader(http.StatusNoContent)
		}

		bookInfoModel := &Book{}
		_, delErr := dbConn.Model(bookInfoModel).
			Where("?=?", pg.Ident("bookname"), book.Bookname).
			Where("?=?", pg.Ident("user_id"), userID).
			Limit(1).Delete()
		if delErr != nil {
			w.Write([]byte("This book is not available in your read list"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	case strings.Contains(r.URL.Path, "/unread"):
		book := &UserRequest{}
		readErr := json.NewDecoder(r.Body).Decode(book)
		if readErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if book.Bookname == "" || book.Bookname == " " {
			w.WriteHeader(http.StatusNoContent)
		}

		unreadBookInfo := &Book{
			UserID:   userID,
			Bookname: book.Bookname,
		}
		_, delErr := dbConn.Model(unreadBookInfo).
			Where("?=?", pg.Ident("bookname"), book.Bookname).
			Where("?=?", pg.Ident("user_id"), userID).
			Limit(1).Delete()
		if delErr != nil {
			w.Write([]byte("This book is not available in your unread list"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
