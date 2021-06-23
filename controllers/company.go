package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/joncalhoun/qson"

	"groupware-gin/models"
	"groupware-gin/responses"
)

/*
 * GET /companies
 *
 * Find some companies
 */

type findParams struct {
	Search string `json:"search" valid:"optional"`
	SortBy string `json:"sort_by" valid:"optional,in(name|since)"`
	Limit  *int   `json:"limit" valid:"optional,range(5|100)"`
}

func (s *Server) FindCompanies(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// validae URL query
	var params findParams
	if r.URL.RawQuery != "" { // hack: qson fails on empty string
		err := qson.Unmarshal(&params, r.URL.RawQuery)
		if err != nil {
			responses.ERROR(w, http.StatusFailedDependency, err)
			return
		}
		params.Search = strings.TrimSpace(params.Search)
		result, err := govalidator.ValidateStruct(params)
		if err != nil {
			responses.ERROR(w, http.StatusFailedDependency, err)
			return
		}
		if !result {
			responses.ERROR(w, http.StatusFailedDependency, errors.New("validation failed"))
			return
		}
	}

	// perform DB query
	query := make([]string, 0)
	query = append(query, "FOR x IN companies")
	bindVars := gin.H{}
	if params.Search != "" {
		query = append(query, "FILTER CONTAINS(x.name, @search)")
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
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	defer cursor.Close()

	// make a result
	companies := []models.Company{}
	var doc models.Company
	for {
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			responses.ERROR(w, http.StatusFailedDependency, err)
			return
		}
		companies = append(companies, doc)
	}
	responses.JSON(w, http.StatusOK, companies)
}

/*
 * GET /companies/{key}
 *
 * Show a company
 */

func (s *Server) ShowCompany(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}

	// validate params
	vars := mux.Vars(r)
	found, err := companies.DocumentExists(ctx, vars["key"])
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	if !found {
		err := errors.New("this company does not exist")
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	// make a result
	var doc models.Company
	_, err = companies.ReadDocument(ctx, vars["key"], &doc)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	responses.JSON(w, http.StatusOK, doc)
}

/*
 * POST /companies
 *
 * Store a company
 */

type storeParams struct {
	Name  string `json:"name" valid:"required,notnull"`
	Since string `json:"since" valid:"required,rfc3339"`
}

func (s *Server) StoreCompany(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// validate payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var params storeParams
	err := dec.Decode(&params)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	params.Name = govalidator.Trim(params.Name, "")
	res, err := govalidator.ValidateStruct(params)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	if !res {
		responses.ERROR(w, http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// create a document
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	var doc models.Company
	otherCtx := driver.WithReturnNew(ctx, &doc)
	anotherCtx := driver.WithKeepNull(otherCtx, false)
	_, err = companies.CreateDocument(anotherCtx, params)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	responses.JSON(w, http.StatusOK, doc)
}

/*
 * PUT /companies/{key}
 *
 * Update a company
 */

func (s *Server) validateParams(w http.ResponseWriter, r *http.Request) (string, error) {
	ctx := context.Background()
	vars := mux.Vars(r)
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		return "", err
	}
	found, err := companies.DocumentExists(ctx, vars["key"])
	if err != nil {
		return vars["key"], err
	}
	if !found {
		return vars["key"], errors.New("does not exist")
	}
	return vars["key"], nil
}

type updateParams struct {
	Name  string `json:"name" validate:"optional,notnull"`
	Since string `json:"since" validate:"optional,rfc3339"`
}

func (s *Server) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// validate params
	key, err := s.validateParams(w, r)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}

	// validate payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var params updateParams
	err = dec.Decode(&params)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	params.Name = govalidator.Trim(params.Name, "")
	result, err := govalidator.ValidateStruct(params)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	if !result {
		responses.ERROR(w, http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// update a document
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	_, err = companies.UpdateDocument(ctx, key, params)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}

	// make a result
	// payload contains some fields optionally
	// we must fetch all fields from db, before we can see them
	var doc models.Company
	_, err = companies.ReadDocument(ctx, key, &doc)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	responses.JSON(w, http.StatusOK, doc)
}

/*
 * DELETE /companies/{key}
 *
 * Delete a company
 */

type deleteParams struct {
	Forever bool `json:"forever" valid:"optional,bool"`
}

func (s *Server) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	ctx := context.Background()

	// validate params
	key, err := s.validateParams(w, r)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}

	// validate payload
	params := deleteParams{}
	err = dec.Decode(&params)
	if err != nil && err != io.EOF {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	res, err := govalidator.ValidateStruct(params)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	if !res {
		responses.ERROR(w, http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// perform an action
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	if params.Forever {
		// delete a document permanently
		_, err = companies.RemoveDocument(ctx, key)
		if err != nil {
			responses.ERROR(w, http.StatusFailedDependency, err)
			return
		}
		responses.JSON(w, http.StatusNoContent, "")
	} else {
		// delete a document temporarily
		var doc models.Company
		otherCtx := driver.WithReturnNew(ctx, &doc)
		_, err = companies.UpdateDocument(otherCtx, key, gin.H{
			"deleted_at": time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			responses.ERROR(w, http.StatusFailedDependency, err)
			return
		}
		responses.JSON(w, http.StatusOK, doc)
	}
}

/*
 * PATCH /companies/{key}
 *
 * Restorer a company that was deleted temporarily
 */

func (s *Server) RestoreCompany(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// validate params
	key, err := s.validateParams(w, r)
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}

	// update a document
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	otherCtx := driver.WithKeepNull(ctx, false)
	var doc models.Company
	anotherCtx := driver.WithReturnNew(otherCtx, &doc)
	_, err = companies.UpdateDocument(anotherCtx, key, gin.H{
		"deleted_at": nil,
	})
	if err != nil {
		responses.ERROR(w, http.StatusFailedDependency, err)
		return
	}
	responses.JSON(w, http.StatusOK, doc)
}
