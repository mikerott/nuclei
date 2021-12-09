package handlers

import (
	"context"
	"database/sql"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/nuclei/v2/pkg/web/db/dbsql"
	"gopkg.in/yaml.v2"
)

// GetTemplatesResponse is a response for /templates listing
type GetTemplatesResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Folder    string    `json:"folder"`
	Path      string    `json:"path"`
	Createdat time.Time `json:"createdAt"`
	Updatedat time.Time `json:"updatedAt"`
}

// GetTemplates handlers /templates listing route
func (s *Server) GetTemplates(ctx echo.Context) error {
	if folder := ctx.QueryParam("folder"); folder != "" {
		return s.getTemplatesWithFolder(ctx, folder)
	}
	if search := ctx.QueryParam("search"); search != "" {
		return s.getTemplatesWithSearchKey(ctx, search)
	}
	return s.getTemplates(ctx)
}

// getTemplates handles getting templates
func (s *Server) getTemplates(ctx echo.Context) error {
	rows, err := s.db.Queries().GetTemplates(context.Background())
	if err != nil {
		return errors.Wrap(err, "could not get templates by folder")
	}
	response := make([]GetTemplatesResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, GetTemplatesResponse{
			ID:        row.ID,
			Name:      row.Name.String,
			Folder:    row.Folder.String,
			Path:      row.Path,
			Createdat: row.Createdat.Time,
			Updatedat: row.Updatedat.Time,
		})
	}
	return ctx.JSON(200, response)
}

// getTemplatesWithFolder handles getting templates by a folder
func (s *Server) getTemplatesWithFolder(ctx echo.Context, folder string) error {
	rows, err := s.db.Queries().GetTemplatesByFolder(context.Background(), sql.NullString{String: folder, Valid: true})
	if err != nil {
		return errors.Wrap(err, "could not get templates by folder")
	}
	response := make([]GetTemplatesResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, GetTemplatesResponse{
			ID:        row.ID,
			Name:      row.Name.String,
			Folder:    folder,
			Path:      row.Path,
			Createdat: row.Createdat.Time,
			Updatedat: row.Updatedat.Time,
		})
	}
	return ctx.JSON(200, response)
}

// getTemplatesWithSearchKey handles getting templates by a search key for path
func (s *Server) getTemplatesWithSearchKey(ctx echo.Context, searchKey string) error {
	rows, err := s.db.Queries().GetTemplatesBySearchKey(context.Background(), searchKey)
	if err != nil {
		return errors.Wrap(err, "could not get templates by search key")
	}
	response := make([]GetTemplatesResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, GetTemplatesResponse{
			ID:        row.ID,
			Name:      row.Name.String,
			Folder:    row.Folder.String,
			Path:      row.Path,
			Createdat: row.Createdat.Time,
			Updatedat: row.Updatedat.Time,
		})
	}
	return ctx.JSON(200, response)
}

// UpdateTemplateRequest is a request for /templates update
type UpdateTemplateRequest struct {
	Contents string `json:"contents"`
	Path     string `json:"path"`
}

// UpdateTemplate handles /templates updating route
func (s *Server) UpdateTemplate(ctx echo.Context) error {
	var body UpdateTemplateRequest
	if err := jsoniter.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return err
	}
	err := s.db.Queries().UpdateTemplate(context.Background(), dbsql.UpdateTemplateParams{
		Contents:  body.Contents,
		Updatedat: sql.NullTime{Time: time.Now(), Valid: true},
		Path:      body.Path,
	})
	return err
}

// AddTemplateRequest is a request for /templates addition
type AddTemplateRequest struct {
	Contents string `json:"contents"`
	Path     string `json:"path"`
	Folder   string `json:"folder"`
}

type templateNameInfoStructure struct {
	Info struct {
		Name string `yaml:"name"`
	} `yaml:"info"`
}

// AddTemplate handles /templates addition route
func (s *Server) AddTemplate(ctx echo.Context) error {
	var body AddTemplateRequest
	if err := jsoniter.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return err
	}
	var templateNameInfo templateNameInfoStructure
	if err := yaml.NewDecoder(strings.NewReader(body.Contents)).Decode(&templateNameInfo); err != nil {
		return err
	}
	err := s.db.Queries().AddTemplate(context.Background(), dbsql.AddTemplateParams{
		Contents: body.Contents,
		Folder:   sql.NullString{String: body.Folder, Valid: true},
		Path:     body.Path,
		Name:     sql.NullString{String: templateNameInfo.Info.Name, Valid: true},
	})
	return err
}

// DeleteTemplateRequest is a request for /templates deletion
type DeleteTemplateRequest struct {
	Path string `json:"path"`
}

// DeleteTemplate handles /templates deletion route
func (s *Server) DeleteTemplate(ctx echo.Context) error {
	var body DeleteTemplateRequest
	if err := jsoniter.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return err
	}
	err := s.db.Queries().DeleteTemplate(context.Background(), body.Path)
	return err
}

// GetTemplatesRaw handlers /templates content retrieval route
func (s *Server) GetTemplatesRaw(ctx echo.Context) error {
	templatePath := ctx.QueryParam("path")
	if templatePath == "" {
		return errors.New("no path parameter specified for template")
	}
	contents, err := s.db.Queries().GetTemplateContents(context.Background(), templatePath)
	if err != nil {
		return err
	}
	return ctx.String(200, contents)
}
