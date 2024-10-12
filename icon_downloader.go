package main

import (
	"cmp"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

//go:embed tmpl.svg
var svgTemplate string

type IconDownloader struct {
	client *http.Client
	tmpl   *template.Template
}

func NewIconDownloader() (*IconDownloader, error) {
	tmpl, err := template.New("default").Parse(svgTemplate)
	if err != nil {
		return nil, fmt.Errorf("template: parse: %w", err)
	}

	return &IconDownloader{
		client: &http.Client{},
		tmpl:   tmpl,
	}, nil
}

func (s *IconDownloader) GetCollections(ctx context.Context) ([]Collection, error) {
	uri := "https://api.iconify.design/collections"

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("http: new request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: do: %w", err)
	}
	defer resp.Body.Close()

	collections := make(map[string]Collection)

	err = json.NewDecoder(resp.Body).Decode(&collections)
	if err != nil {
		return nil, fmt.Errorf("json: decode: %w", err)
	}

	list := lo.MapToSlice(collections, func(k string, v Collection) Collection {
		v.ID = k
		return v
	})

	slices.SortFunc(list, func(a, b Collection) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return list, nil
}

func (s *IconDownloader) GetIconList(ctx context.Context, collectionID string) ([]string, error) {
	uri := fmt.Sprintf("https://api.iconify.design/collection?prefix=%s&chars=true&aliases=true", collectionID)

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("http: new request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: do: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Categories    map[string][]string `json:"categories"`
		Uncategorized []string            `json:"uncategorized"`
	}

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("json: decode: %w", err)
	}

	icons := make([]string, 0)

	for _, v := range body.Categories {
		for i := range v {
			if !slices.Contains(icons, v[i]) {
				icons = append(icons, v[i])
			}
		}
	}

	for _, v := range body.Uncategorized {
		if !slices.Contains(icons, v) {
			icons = append(icons, v)
		}
	}

	slices.Sort(icons)

	return icons, nil
}

func (s *IconDownloader) GetIcons(ctx context.Context, basePath, collectionID string, iconIDs []string, width, height, color string, program *tea.Program) error {
	uri := fmt.Sprintf("https://api.iconify.design/%s.json?icons=%s", collectionID, strings.Join(iconIDs, ","))

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return fmt.Errorf("http: new request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("client: do: %w", err)
	}
	defer resp.Body.Close()

	var body IconResp

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return fmt.Errorf("json: decode: %w", err)
	}

	for k, v := range body.Icons {
		err := os.MkdirAll(filepath.Join(basePath, collectionID), os.ModePerm)
		if err != nil {
			return fmt.Errorf("os: mkdirall: %w", err)
		}

		file, err := os.Create(filepath.Join(basePath, collectionID, fmt.Sprintf("%s.svg", k)))
		if err != nil {
			return fmt.Errorf("os: create: %w", err)
		}

		b := v.Body

		if color != "" {
			b = strings.ReplaceAll(b, "currentColor", color)
		}

		canvasWidth := body.Width
		canvasHeight := body.Height

		if v.Height != 0 {
			canvasHeight = v.Height
		}

		if v.Width != 0 {
			canvasWidth = v.Width
		}

		err = s.tmpl.ExecuteTemplate(file, "default", map[string]any{
			"Width":        width,
			"Height":       height,
			"CanvasWidth":  canvasWidth,
			"CanvasHeight": canvasHeight,
			"Body":         b,
		})
		if err != nil {
			return fmt.Errorf("template: execute: %w", err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("file: close: %w", err)
		}

		program.Send(DownloadEvent{
			Collection: collectionID,
			Icon:       k,
		})
	}

	return nil
}
