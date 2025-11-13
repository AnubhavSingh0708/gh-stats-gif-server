package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	// We need to import jpeg and png to allow image.Decode to work with them
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/valyala/fasthttp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// GitHubUser holds the stats we care about from the GitHub API /users endpoint
type GitHubUser struct {
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	PublicRepos int    `json:"public_repos"`
	ReposURL    string `json:"repos_url"`
}

// GitHubRepo is a single repo, used for counting stars
type GitHubRepo struct {
	StargazersCount int `json:"stargazers_count"`
}

// StatsData holds all the final info we want to draw
type StatsData struct {
	Name        string
	Avatar      image.Image
	Followers   int
	Following   int
	PublicRepos int
	TotalStars  int
}

// Theme defines the colors for the image
type Theme struct {
	BGColor    color.Color
	TextColor  color.Color
	TitleColor color.Color
	StatsColor color.Color
}

// themes holds our predefined color palettes
var themes = map[string]Theme{
	"light": {
		BGColor:    color.White,
		TextColor:  color.Black,
		TitleColor: color.RGBA{R: 40, G: 120, B: 200, A: 255},
		StatsColor: color.RGBA{R: 50, G: 50, B: 50, A: 255},
	},
	"dark": {
		BGColor:    color.RGBA{R: 30, G: 30, B: 30, A: 255},
		TextColor:  color.RGBA{R: 230, G: 230, B: 230, A: 255},
		TitleColor: color.RGBA{R: 100, G: 180, B: 255, A: 255},
		StatsColor: color.RGBA{R: 200, G: 200, B: 200, A: 255},
	},
	// You can add more themes here! e.g., "a"
	"a": {
		BGColor:    color.RGBA{R: 255, G: 240, B: 240, A: 255}, // Light pink
		TextColor:  color.RGBA{R: 80, G: 20, B: 20, A: 255},    // Dark red
		TitleColor: color.RGBA{R: 200, G: 50, B: 50, A: 255},   // Red
		StatsColor: color.RGBA{R: 120, G: 40, B: 40, A: 255},   // Darker red
	},
}

// Font cache for common font sizes
var fontCache = make(map[int]font.Face)

// requestHandler is the main entry point for all server requests
func requestHandler(ctx *fasthttp.RequestCtx) {
	// 1. Parse query parameters
	githubIDBytes := ctx.QueryArgs().Peek("id")
	themeNameBytes := ctx.QueryArgs().Peek("theme")

	if len(githubIDBytes) == 0 {
		ctx.Error("Missing 'id' query parameter", fasthttp.StatusBadRequest)
		return
	}

	githubID := string(githubIDBytes)
	themeName := string(themeNameBytes)

	// Select theme, default to "light"
	theme, ok := themes[themeName]
	if !ok {
		theme = themes["light"]
	}

	// 2. Fetch all GitHub stats
	stats, err := fetchStatsData(githubID)
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to get GitHub stats for %s: %v", githubID, err), fasthttp.StatusNotFound)
		return
	}

	// 3. Create the stats image
	imgBuf, err := createStatsImage(stats, theme)
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to create image: %v", err), fasthttp.StatusInternalServerError)
		return
	}

	// 4. Serve the image
	ctx.SetContentType("image/png")
	// Set caching headers. This is IMPORTANT to avoid hitting API rate limits.
	// Cache for 1 hour in browser (max-age) and on CDNs/proxies (s-maxage)
	ctx.Response.Header.Set("Cache-Control", "public, max-age=3600, s-maxage=3600")
	ctx.Write(imgBuf.Bytes())
}

// fetchStatsData orchestrates all the API calls
func fetchStatsData(username string) (*StatsData, error) {
	// 1. Get primary user data
	user, err := getUserData(username)
	if err != nil {
		return nil, fmt.Errorf("could not get user data: %w", err)
	}

	// 2. Get the avatar image
	avatar, err := getAvatar(user.AvatarURL)
	if err != nil {
		return nil, fmt.Errorf("could not get avatar: %w", err)
	}

	// 3. Get total stars
	totalStars, err := getTotalStars(user.ReposURL)
	if err != nil {
		// Don't fail the whole request, just set stars to 0
		log.Printf("Warning: could not get stars for %s: %v", username, err)
		totalStars = 0
	}

	// Use username as Name if 'name' field is null
	displayName := user.Name
	if displayName == "" {
		displayName = username
	}

	// 4. Assemble final data
	stats := &StatsData{
		Name:        displayName,
		Avatar:      avatar,
		Followers:   user.Followers,
		Following:   user.Following,
		PublicRepos: user.PublicRepos,
		TotalStars:  totalStars,
	}

	return stats, nil
}

// getUserData fetches primary user data from the GitHub API
func getUserData(username string) (*GitHubUser, error) {
	apiURL := "https://api.github.com/users/" + username

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user GitHubUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// getAvatar fetches the user's profile picture
func getAvatar(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avatar URL returned status: %s", resp.Status)
	}

	// image.Decode can handle JPEG, PNG, etc.
	// as long as the correct format packages are imported
	avatar, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return avatar, nil
}

// getTotalStars fetches all repos (up to 100) and sums their stars
func getTotalStars(reposURL string) (int, error) {
	// Get first 100 repos. For a user with > 100, we'd need to handle pagination
	// This is a good-enough approximation for this app.
	apiURL := reposURL + "?per_page=100"

	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("repos API returned status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var repos []GitHubRepo
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return 0, err
	}

	totalStars := 0
	for _, repo := range repos {
		totalStars += repo.StargazersCount
	}

	return totalStars, nil
}

// resizeImage resizes an image to the specified width and height by sampling pixels
// This uses nearest-neighbor sampling to preserve the image quality
func resizeImage(src image.Image, width, height int) image.Image {
	// Get the source image bounds
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Max.X - srcBounds.Min.X
	srcHeight := srcBounds.Max.Y - srcBounds.Min.Y

	// Create a new RGBA image with the target dimensions
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Calculate the scaling factors
	xRatio := float64(srcWidth) / float64(width)
	yRatio := float64(srcHeight) / float64(height)

	// Sample pixels from the source image and fill the destination
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate the source coordinates by sampling proportionally
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)

			// Ensure we stay within bounds
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}

			// Get the pixel from the source and set it in the destination
			r, g, b, a := src.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY).RGBA()
			dst.SetRGBA(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
		}
	}

	return dst
}

// getFontFace returns a font face for the given size, trying system fonts
func getFontFace(size int) font.Face {
	if face, ok := fontCache[size]; ok {
		return face
	}

	// Try common system font paths - prefer Inconsolata regular (not bold)
	fontPaths := []string{
		"/usr/share/fonts/truetype/inconsolata/Inconsolata.ttf",
		"/usr/share/fonts/truetype/inconsolata/Inconsolata-Regular.ttf",
		"/usr/share/fonts/opentype/inconsolata/Inconsolata-Regular.otf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationMono-Regular.ttf",
		"/System/Library/Fonts/Menlo.ttc",
		"C:\\Windows\\Fonts\\consola.ttf",
	}

	for _, path := range fontPaths {
		fontData, err := ioutil.ReadFile(path)
		if err != nil {
			continue
		}

		parsedFont, err := opentype.Parse(fontData)
		if err != nil {
			continue
		}

		face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
			Size: float64(size),
			DPI:  72,
		})
		if err != nil {
			continue
		}

		fontCache[size] = face
		return face
	}

	// Fallback: return a simple monospace face if no system font found
	log.Printf("Warning: Could not find system font, using default rendering")
	return font.Face(nil)
}

// createStatsImage draws the stats onto a new image and returns the PNG bytes
func createStatsImage(stats *StatsData, theme Theme) (*bytes.Buffer, error) {
	// Define image dimensions
	width := 1200
	height := 800

	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the background
	draw.Draw(img, img.Bounds(), &image.Uniform{C: theme.BGColor}, image.Point{}, draw.Src)

	// --- Draw Avatar ---
	avatarSize := 160
	// Create a destination rectangle for the avatar
	avatarDestRect := image.Rect(53, 53, 53+avatarSize, 53+avatarSize)
	// Resize the avatar by sampling pixels proportionally from the source image
	resizedAvatar := resizeImage(stats.Avatar, avatarSize, avatarSize)
	// Draw the resized avatar into the destination rectangle
	draw.Draw(img, avatarDestRect, resizedAvatar, image.Point{0, 0}, draw.Src)

	// --- Draw Text ---
	// NOTE: Using system TrueType fonts for larger sizes

	// Draw the Name, vertically centered with the avatar
	addLabel(img, 53+avatarSize+53, 147, stats.Name, theme.TitleColor, 72)

	// --- Draw the stats in two columns ---
	yPos := 293 // Start below the avatar
	statSpacing := 67
	xPos1 := 53
	xPos2 := 640 // Start of second column

	addLabel(img, xPos1, yPos, fmt.Sprintf("Followers: %d", stats.Followers), theme.StatsColor, 48)
	addLabel(img, xPos2, yPos, fmt.Sprintf("Following: %d", stats.Following), theme.StatsColor, 48)

	yPos += statSpacing
	addLabel(img, xPos1, yPos, fmt.Sprintf("Public Repos: %d", stats.PublicRepos), theme.StatsColor, 48)
	addLabel(img, xPos2, yPos, fmt.Sprintf("Total Stars: %d", stats.TotalStars), theme.StatsColor, 48)

	// --- Encode to PNG ---
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// addLabel is a helper function to draw text on the image with specified font size
func addLabel(img draw.Image, x, y int, label string, clr color.Color, fontSize int) {
	face := getFontFace(fontSize)
	if face == nil {
		log.Printf("Warning: Could not load font, skipping text: %s", label)
		return
	}

	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{C: clr},
		Face: face,
		Dot:  point,
	}
	d.DrawString(label)
}

// main starts the fasthttp server
func main() {
	port := ":8800"
	log.Printf("Starting GitHub stats server on %s...", port)

	// Start the server
	if err := fasthttp.ListenAndServe(port, requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %v", err)
	}
}
