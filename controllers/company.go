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
	"github.com/joncalhoun/qson"

	"groupware-gin/models"
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

func (s *Server) FindCompanies(c *gin.Context) {
	ctx := context.Background()

	// validae URL query
	var params findParams
	if c.Request.URL.RawQuery != "" { // hack: qson fails on empty string
		err := qson.Unmarshal(&params, c.Request.URL.RawQuery)
		if err != nil {
			c.JSON(http.StatusFailedDependency, err)
			return
		}
		params.Search = strings.TrimSpace(params.Search)
		result, err := govalidator.ValidateStruct(params)
		if err != nil {
			c.JSON(http.StatusFailedDependency, err)
			return
		}
		if !result {
			c.JSON(http.StatusFailedDependency, errors.New("validation failed"))
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
		c.JSON(http.StatusFailedDependency, err)
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
			c.JSON(http.StatusFailedDependency, err)
			return
		}
		companies = append(companies, doc)
	}
	c.JSON(http.StatusOK, companies)
}

/*
 * GET /companies/{key}
 *
 * Show a company
 */

func (s *Server) ShowCompany(c *gin.Context) {
	ctx := context.Background()
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}

	// validate params
	key := c.Param("key")
	found, err := companies.DocumentExists(ctx, key)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, errors.New("this company does not exist"))
		return
	}

	// make a result
	var doc models.Company
	_, err = companies.ReadDocument(ctx, key, &doc)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	c.JSON(http.StatusOK, doc)
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

func (s *Server) StoreCompany(c *gin.Context) {
	ctx := context.Background()

	// validate payload
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	var params storeParams
	err := dec.Decode(&params)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	params.Name = govalidator.Trim(params.Name, "")
	res, err := govalidator.ValidateStruct(params)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	if !res {
		c.JSON(http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// create a document
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	var doc models.Company
	otherCtx := driver.WithReturnNew(ctx, &doc)
	anotherCtx := driver.WithKeepNull(otherCtx, false)
	_, err = companies.CreateDocument(anotherCtx, params)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}

/*
 * PUT /companies/{key}
 *
 * Update a company
 */

func (s *Server) validateParams(c *gin.Context) (string, error) {
	ctx := context.Background()
	key := c.Param("key")
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		return "", err
	}
	found, err := companies.DocumentExists(ctx, key)
	if err != nil {
		return key, err
	}
	if !found {
		return key, errors.New("does not exist")
	}
	return key, nil
}

type updateParams struct {
	Name  string `json:"name" validate:"optional,notnull"`
	Since string `json:"since" validate:"optional,rfc3339"`
}

func (s *Server) UpdateCompany(c *gin.Context) {
	ctx := context.Background()

	// validate params
	key, err := s.validateParams(c)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}

	// validate payload
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	var params updateParams
	err = dec.Decode(&params)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	params.Name = govalidator.Trim(params.Name, "")
	result, err := govalidator.ValidateStruct(params)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	if !result {
		c.JSON(http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// update a document
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	_, err = companies.UpdateDocument(ctx, key, params)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}

	// make a result
	// payload contains some fields optionally
	// we must fetch all fields from db, before we can see them
	var doc models.Company
	_, err = companies.ReadDocument(ctx, key, &doc)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}

/*
 * DELETE /companies/{key}
 *
 * Delete a company
 */

type deleteParams struct {
	Forever bool `json:"forever" valid:"optional,bool"`
}

func (s *Server) DeleteCompany(c *gin.Context) {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	ctx := context.Background()

	// validate params
	key, err := s.validateParams(c)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}

	// validate payload
	params := deleteParams{}
	err = dec.Decode(&params)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	res, err := govalidator.ValidateStruct(params)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	if !res {
		c.JSON(http.StatusBadRequest, errors.New("validation failed"))
		return
	}

	// perform an action
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	if params.Forever {
		// delete a document permanently
		_, err = companies.RemoveDocument(ctx, key)
		if err != nil {
			c.JSON(http.StatusFailedDependency, err)
			return
		}
		c.JSON(http.StatusNoContent, "")
	} else {
		// delete a document temporarily
		var doc models.Company
		otherCtx := driver.WithReturnNew(ctx, &doc)
		_, err = companies.UpdateDocument(otherCtx, key, gin.H{
			"deleted_at": time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			c.JSON(http.StatusFailedDependency, err)
			return
		}
		c.JSON(http.StatusOK, doc)
	}
}

/*
 * PATCH /companies/{key}
 *
 * Restorer a company that was deleted temporarily
 */

func (s *Server) RestoreCompany(c *gin.Context) {
	ctx := context.Background()

	// validate params
	key, err := s.validateParams(c)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}

	// update a document
	companies, err := s.DB.Collection(ctx, "companies")
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	otherCtx := driver.WithKeepNull(ctx, false)
	var doc models.Company
	anotherCtx := driver.WithReturnNew(otherCtx, &doc)
	_, err = companies.UpdateDocument(anotherCtx, key, gin.H{
		"deleted_at": nil,
	})
	if err != nil {
		c.JSON(http.StatusFailedDependency, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}
