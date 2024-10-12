package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/samber/lo"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("path parameter required")
	}

	absPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	iconDownloader, err := NewIconDownloader()
	if err != nil {
		log.Fatalln(err)
	}

	collections, err := iconDownloader.GetCollections(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	selectedCollections := make([]Collection, 0)

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[Collection]().Title("Select icon collections").
				Options(
					lo.Map(collections, func(v Collection, _ int) huh.Option[Collection] {
						return huh.NewOption(fmt.Sprintf("%s (%d)", v.Name, v.Total), v)
					})...,
				).
				Value(&selectedCollections).Height(15),
		),
	).Run()
	if err != nil {
		log.Fatalln(err)
	}

	allSelectedIcons := make(map[string][]string)

	for _, v := range selectedCollections {
		selectedIcons := make([]string, 0)

		icons, err := iconDownloader.GetIconList(ctx, v.ID)
		if err != nil {
			log.Fatalln(err)
		}

		err2 := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().Title(v.Name).Description(fmt.Sprintf("id: %s, total: %d", v.ID, v.Total)).
					Options(
						lo.Map(icons, func(iv string, _ int) huh.Option[string] {
							return huh.NewOption(iv, iv)
						})...,
					).
					Value(&selectedIcons).Height(15),
			),
		).Run()
		if err2 != nil {
			log.Fatalln(err)
		}

		allSelectedIcons[v.ID] = selectedIcons
	}

	iconCount := 0
	for _, v := range allSelectedIcons {
		iconCount += len(v)
	}

	if iconCount == 0 {
		log.Fatalln("No icons selected")
	}

	isConfirmed := false

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title(fmt.Sprintf("Are you sure you want to download %d icons?", iconCount)).
				Value(&isConfirmed),
		),
	).Run()
	if err != nil {
		log.Fatalln(err)
	}

	if !isConfirmed {
		return
	}

	var width, height, color string

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Width").Description("optional").Value(&width),
			huh.NewInput().Title("Height").Description("optional").Value(&height),
			huh.NewInput().Title("Color").Description("optional").Value(&color),
		),
	).Run()
	if err != nil {
		log.Fatalln(err)
	}

	ui := NewUIModel(iconCount)

	program := tea.NewProgram(ui)

	go func() {
		_, err = program.Run()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	startTime := time.Now()
	for collection, icons := range allSelectedIcons {
		for _, chunk := range lo.Chunk(icons, 10) {
			err3 := iconDownloader.GetIcons(ctx, absPath, collection, chunk, width, height, color, program)
			if err3 != nil {
				log.Fatalln(err3)
			}
		}
	}
	program.Quit()
	program.Wait()

	fmt.Printf("\n%d icons downloaded! Took: %s\n", iconCount, time.Since(startTime).String())
}
