package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

//go:embed static templates
var content embed.FS

// Flag structure for levels
type Flags struct {
	Levels map[string]string `yaml:"levels"`
}

var flags Flags

func main() {
	// Load flags from YAML
	if err := loadFlags(); err != nil {
		panic(fmt.Sprintf("Failed to load flags: %v", err))
	}

	// Initialize Gin
	router := gin.Default()

	// Load templates from embed.FS
	router.SetHTMLTemplate(loadTemplates())

	// Serve static files
	router.StaticFS("/static", http.FS(content))

	// Routes
	router.GET("/", homeHandler)
	router.GET("/level/:id", levelHandler)
	router.POST("/level/:id", validateFlag)

	// Start the server
	fmt.Println("Server running at http://localhost:8080")
	router.Run(":8080")
}

// Load flags from YAML
func loadFlags() error {
	data, err := content.ReadFile("static/flags.yaml")
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &flags)
}

// Load HTML templates from embed.FS
func loadTemplates() *template.Template {
	tmpl := template.New("")
	template.Must(tmpl.ParseFS(content, "templates/*.html"))
	return tmpl
}

// Handlers

// Home page
func homeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "CTF Game",
		"message": "Welcome to the CTF Game! Start your journey by solving the challenges.",
	})
}

// Render the level page
func levelHandler(c *gin.Context) {
	levelID := c.Param("id")

	// Check if the level exists
	if _, exists := flags.Levels[levelID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": "Level not found"})
		return
	}

	c.HTML(http.StatusOK, "level.html", gin.H{
		"level": levelID,
	})
}

// Validate flag submission
func validateFlag(c *gin.Context) {
	levelID := c.Param("id")
	submittedFlag := c.PostForm("flag")

	// Check if the level exists
	correctFlag, exists := flags.Levels[levelID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": "Level not found"})
		return
	}

	// Validate the submitted flag
	if submittedFlag == correctFlag {
		nextLevel := strconv.Itoa(getNextLevel(levelID))
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Correct flag! Proceed to the next level.",
			"next":    fmt.Sprintf("/level/%s", nextLevel),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Incorrect flag. Try again!",
		})
	}
}

// Get the next level ID
func getNextLevel(currentLevel string) int {
	level, err := strconv.Atoi(currentLevel)
	if err != nil {
		return 0 // Return 0 if conversion fails
	}
	return level + 1
}
