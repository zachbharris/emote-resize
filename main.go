package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

// EmoteSize represents a target emote size with platform and dimensions
type EmoteSize struct {
	Platform string
	Name     string
	Width    int
	Height   int
}

// Define all emote sizes
var emoteSizes = []EmoteSize{
	// Discord emote sizes
	{"Discord", "Small", 28, 28},
	{"Discord", "Medium", 32, 32},
	{"Discord", "Large", 48, 48},
	{"Discord", "Animated", 128, 128},

	// Twitch emote sizes
	{"Twitch", "1.0", 28, 28},
	{"Twitch", "2.0", 56, 56},
	{"Twitch", "3.0", 112, 112},

	// 7TV emote sizes
	{"7TV", "1x", 32, 32},
	{"7TV", "2x", 64, 64},
	{"7TV", "3x", 96, 96},
	{"7TV", "4x", 128, 128},
}

type App struct {
	window       fyne.Window
	selectedFile string
	statusLabel  *widget.Label
	convertBtn   *widget.Button
}

func main() {
	myApp := app.New()
	myApp.SetIcon(nil)
	
	w := myApp.NewWindow("Emote Size Converter")
	w.Resize(fyne.NewSize(500, 300))
	w.CenterOnScreen()

	converter := &App{
		window:      w,
		statusLabel: widget.NewLabel("No file selected"),
		convertBtn:  widget.NewButton("Convert & Save Bundle", nil),
	}

	converter.convertBtn.Disable()
	converter.setupUI()
	
	w.ShowAndRun()
}

func (a *App) setupUI() {
	title := widget.NewCard("Emote Size Converter", "", 
		widget.NewLabel("Convert images to Discord, Twitch, and 7TV emote sizes"))

	selectBtn := widget.NewButton("Select Image File", a.selectFile)
	selectBtn.Importance = widget.MediumImportance

	a.convertBtn.OnTapped = a.convertAndSave
	a.convertBtn.Importance = widget.HighImportance

	// Create info about supported formats
	formatInfo := widget.NewCard("Supported Formats", "", 
		widget.NewLabel("JPEG, PNG, GIF, WebP, WebM"))

	// Create size info
	sizeInfo := widget.NewRichTextFromMarkdown(`**Target Sizes:**
- Discord: 28x28, 32x32, 48x48, 128x128
- Twitch: 28x28, 56x56, 112x112  
- 7TV: 32x32, 64x64, 96x96, 128x128`)

	buttonContainer := container.NewHBox(selectBtn, a.convertBtn)
	
	content := container.NewVBox(
		title,
		container.NewHBox(formatInfo, sizeInfo),
		widget.NewSeparator(),
		buttonContainer,
		a.statusLabel,
	)

	a.window.SetContent(container.NewPadded(content))
}

func (a *App) selectFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			a.showError("Error opening file", err)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// Check file extension
		uri := reader.URI()
		ext := strings.ToLower(filepath.Ext(uri.Path()))
		
		validExts := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".gif":  true,
			".webp": true,
			".webm": true,
		}

		if !validExts[ext] {
			a.showError("Invalid file type", fmt.Errorf("please select a JPEG, PNG, GIF, WebP, or WebM file"))
			return
		}

		a.selectedFile = uri.Path()
		filename := filepath.Base(a.selectedFile)
		a.statusLabel.SetText(fmt.Sprintf("Selected: %s", filename))
		a.convertBtn.Enable()

	}, a.window)
}

func (a *App) convertAndSave() {
	if a.selectedFile == "" {
		return
	}

	a.convertBtn.Disable()
	a.statusLabel.SetText("Converting images...")

	go func() {
		err := a.processImage()
		if err != nil {
			a.showError("Conversion failed", err)
			a.convertBtn.Enable()
			return
		}

		a.statusLabel.SetText("Conversion completed successfully!")
		a.convertBtn.Enable()
		
		// Show success dialog
		dialog.ShowInformation("Success", 
			"All emote sizes have been created and saved!", a.window)
	}()
}

func (a *App) processImage() error {
	// Open and decode the image
	file, err := os.Open(a.selectedFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode based on file extension
	var img image.Image
	ext := strings.ToLower(filepath.Ext(a.selectedFile))
	
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	case ".gif":
		img, err = gif.Decode(file)
	case ".webp":
		img, err = webp.Decode(file)
	case ".webm":
		// WebM is video format, but we'll try to decode as image
		img, _, err = image.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}

	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Get base filename without extension
	baseFilename := strings.TrimSuffix(filepath.Base(a.selectedFile), filepath.Ext(a.selectedFile))
	outputDir := filepath.Dir(a.selectedFile)

	// Create output directory for the bundle
	bundleDir := filepath.Join(outputDir, baseFilename+"_emote_bundle")
	err = os.MkdirAll(bundleDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Convert to all sizes
	for _, size := range emoteSizes {
		// Resize image maintaining aspect ratio, then crop to exact size
		resized := imaging.Fill(img, size.Width, size.Height, imaging.Center, imaging.Lanczos)
		
		// Create filename
		filename := fmt.Sprintf("%s-%s-%s-%dx%d.png", 
			baseFilename, size.Platform, size.Name, size.Width, size.Height)
		outputPath := filepath.Join(bundleDir, filename)

		// Save as PNG to preserve transparency
		err = imaging.Save(resized, outputPath)
		if err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}
	}

	return nil
}

func (a *App) showError(title string, err error) {
	dialog.ShowError(err, a.window)
	a.statusLabel.SetText(fmt.Sprintf("Error: %s", err.Error()))
}
