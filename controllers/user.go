package controllers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/joncalhoun/qson"

	"groupware-gin/models"
)

/*
 * GET /users
 *
 * Find some users
 */

type FindUsersParams struct {
	Search string `json:"search" valid:"optional"`
	SortBy string `json:"sort_by" valid:"optional,in(name|email)"`
	Limit  *int   `json:"limit" valid:"optional,range(5|100)"`
}

func (s *Server) FindUsers(c *gin.Context) {
	ctx := context.Background()

	// validae URL query
	var params FindUsersParams
	if c.Request.URL.RawQuery != "" { // hack: qson fails on empty string
		err := qson.Unmarshal(&params, c.Request.URL.RawQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		params.Search = strings.TrimSpace(params.Search)
		result, err := govalidator.ValidateStruct(params)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		if !result {
			c.JSON(http.StatusBadRequest, errors.New("validation failed"))
			return
		}
	}

	// perform DB query
	query := make([]string, 0)
	query = append(query, "FOR x IN users")
	bindVars := gin.H{}
	if params.Search != "" {
		query = append(query, "FILTER CONTAINS(x.name, @search) || CONTAINS(x.email, @search)")
		bindVars["search"] = params.Search
	}
	if params.SortBy != "" {
		query = append(query, "SORT x."+params.SortBy+" ASC")
	}
	if params.Limit != nil {
		query = append(query, "LIMIT 0, @limit")
		bindVars["limit"] = params.Limit
	}
	query = append(query, "RETURN x")
	cursor, err := s.DB.Query(ctx, strings.Join(query, " "), bindVars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer cursor.Close()

	// make a result
	users := []models.User{}
	var doc models.User
	for {
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		users = append(users, doc)
	}
	c.JSON(http.StatusOK, users)
}

/*
 * GET /users/:key
 *
 * Show a user
 */

func (s *Server) ShowUser(c *gin.Context) {
	ctx := context.Background()
	users, err := s.DB.Collection(ctx, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// validate params
	key := c.Param("key")
	found, err := users.DocumentExists(ctx, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, errors.New("this user does not exist"))
		return
	}

	// make a result
	var doc models.User
	_, err = users.ReadDocument(ctx, key, &doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}

/*
 * POST /users
 *
 * Store a user
 */

type StoreUserParams struct {
	Name                 string `json:"name" valid:"required,notnull"`
	Email                string `json:"email" valid:"required,email"`
	Password             string `json:"password" valid:"required,length(6|64),store_confirmed"`
	PasswordConfirmation string `json:"password_confirmation" valid:"required"`
}

func (s *Server) StoreUser(c *gin.Context) {
	ctx := context.Background()

	// validate payload
	params := StoreUserParams{
		Name:                 govalidator.Trim(c.Request.FormValue("name"), ""), // default trim removes space
		Email:                c.Request.FormValue("email"),
		Password:             c.Request.FormValue("password"),
		PasswordConfirmation: c.Request.FormValue("password_confirmation"),
	}
	res, err := govalidator.ValidateStruct(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if !res {
		c.JSON(http.StatusBadRequest, errors.New("validation failed"))
		return
	}
	if params.Password != c.Request.FormValue("password_confirmation") {
		c.JSON(http.StatusBadRequest, errors.New("password not matched"))
		return
	}

	// create a document
	found, err := s.HasCollection("users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var users driver.Collection
	if found {
		col, err := s.DB.Collection(ctx, "users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		users = col
	} else {
		options := &driver.CreateCollectionOptions{
			Type: driver.CollectionTypeDocument,
		}
		col, err := s.DB.CreateCollection(ctx, "users", options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		users = col
	}
	hasher := md5.New()
	now := time.Now().UTC()
	meta, err := users.CreateDocument(ctx, gin.H{
		"name":       params.Name,
		"email":      params.Email,
		"password":   hex.EncodeToString(hasher.Sum([]byte(params.Password))),
		"created_at": now,
		"updated_at": now,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// accept the uploaded file
	fileName, err := AcceptFile(c, "avatar", "storage/users/"+meta.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var doc models.User
	otherCtx := driver.WithReturnNew(ctx, &doc)
	_, err = users.UpdateDocument(otherCtx, meta.Key, gin.H{
		"avatar": "users/" + meta.Key + "/" + fileName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}

/*
 * PATCH /users/:key
 *
 * Update a user
 */

func (s *Server) validateUserParams(c *gin.Context) (string, error) {
	ctx := context.Background()
	key := c.Param("key")
	users, err := s.DB.Collection(ctx, "users")
	if err != nil {
		return "", err
	}
	found, err := users.DocumentExists(ctx, key)
	if err != nil {
		return key, err
	}
	if !found {
		return key, errors.New("does not exist")
	}
	return key, nil
}

type UpdateUserParams struct {
	Name                 string `json:"name,omitempty" validate:"optional,notnull"`
	Email                string `json:"email,omitempty" valid:"optional,email"`
	Password             string `json:"password,omitempty" valid:"optional,length(6|64),update_confirmed"`
	PasswordConfirmation string `json:"password_confirmation,omitempty" valid:"optional"`
}

func (s *Server) UpdateUser(c *gin.Context) {
	ctx := context.Background()

	// validate params
	key, err := s.validateUserParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// validate payload
	var params UpdateUserParams
	name := c.Request.FormValue("name")
	if name != "" {
		params.Name = govalidator.Trim(name, "") // default trim removes space
	}
	email := c.Request.FormValue("email")
	if email != "" {
		params.Email = email
	}
	password := c.Request.FormValue("password")
	if password != "" {
		params.Password = password
	}
	passwordConfirmation := c.Request.FormValue("password_confirmation")
	if passwordConfirmation != "" {
		params.PasswordConfirmation = passwordConfirmation
	}
	result, err := govalidator.ValidateStruct(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if !result {
		c.JSON(http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// accept the uploaded file
	fileName, err := AcceptFile(c, "avatar", "storage/users/"+key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// update a document
	users, err := s.DB.Collection(ctx, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	data := gin.H{
		"updated_at": time.Now().UTC(),
	}
	if params.Name != "" {
		data["name"] = params.Name
	}
	if params.Email != "" {
		data["email"] = params.Email
	}
	if params.Password != "" {
		hasher := md5.New()
		data["password"] = hex.EncodeToString(hasher.Sum([]byte(params.Password)))
	}
	var doc models.User
	if fileName != "" {
		_, err := users.ReadDocument(ctx, key, &doc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		os.Remove("storage/" + doc.Avatar)
		data["avatar"] = "users/" + key + "/" + fileName
	}
	otherCtx := driver.WithReturnNew(ctx, &doc)
	_, err = users.UpdateDocument(otherCtx, key, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}

/*
 * DELETE /users/:key
 *
 * Delete a user
 */

type DeleteUserParams struct {
	Mode string `json:"mode" valid:"required,in(erase|trash|restore)"`
}

func (s *Server) DeleteUser(c *gin.Context) {
	ctx := context.Background()

	// validate params
	key, err := s.validateUserParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// validate payload
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	var params DeleteUserParams
	err = dec.Decode(&params)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	res, err := govalidator.ValidateStruct(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if !res {
		c.JSON(http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// perform an action
	users, err := s.DB.Collection(ctx, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if params.Mode == "erase" {
		// delete a document permanently
		os.RemoveAll("storage/users/" + key)
		_, err = users.RemoveDocument(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusNoContent, "")
	} else if params.Mode == "trash" {
		// delete a document temporarily
		var doc models.User
		otherCtx := driver.WithReturnNew(ctx, &doc)
		_, err = users.UpdateDocument(otherCtx, key, gin.H{
			"deleted_at": time.Now().UTC(),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, doc)
	} else if params.Mode == "restore" {
		// restore a document that was deleted temprarily
		otherCtx := driver.WithKeepNull(ctx, false) // don't keep empty field
		var doc models.User
		anotherCtx := driver.WithReturnNew(otherCtx, &doc)
		_, err = users.UpdateDocument(anotherCtx, key, gin.H{
			"deleted_at": nil,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, doc)
	}
}
