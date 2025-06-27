package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm" // Pour gérer gorm.ErrRecordNotFound
)



var ClickEventsChannel chan *models.ClickEvent

// SetupRoutes configure toutes les routes de l'API Gin et injecte les dépendances nécessaires
func SetupRoutes(router *gin.Engine, linkService *services.LinkService, clickService *services.ClickService) {
	cfg := cmd2.Cfg
	if cfg == nil {
		log.Fatal("Configuration non chargée. Veuillez vérifier la configuration.")
	}
	
	// Le channel est initialisé ici.
	if ClickEventsChannel == nil {
		ClickEventsChannel = make(chan *models.ClickEvent, cfg.Analytics.BufferSize)
	}

	router.GET("/health", HealthCheckHandler)


	router.POST("/api/v1/links", CreateShortLinkHandler(linkService))
	router.GET("/api/v1/links/:shortCode/stats", GetLinkStatsHandler(linkService))

	// Route de Redirection (au niveau racine pour les short codes)
	router.GET("/:shortCode", RedirectHandler(linkService, clickService))
}

// HealthCheckHandler gère la route /health pour vérifier l'état du service.
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CreateLinkRequest représente le corps de la requête JSON pour la création d'un lien.
type CreateLinkRequest struct {
	LongURL string `json:"long_url" binding:"required,url"` // 'binding:required' pour validation, 'url' pour format URL
}

// CreateShortLinkHandler gère la création d'une URL courte.
func CreateShortLinkHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateLinkRequest
		
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatal("Configuration non chargée. Veuillez vérifier la configuration.")
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			// Si la validation échoue, retourne une erreur 400 Bad Request avec le message
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
			return
		}


		link, err := linkService.CreateLink(req.LongURL)
		if err != nil {
			log.Printf("Error creating link for %s: %v", req.LongURL, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Retourne le code court et l'URL longue dans la réponse JSON.
		c.JSON(http.StatusCreated, gin.H{
			"short_code":     link.Shortcode,
			"long_url":       link.LongURL,
			"full_short_url": cfg.Server.BaseURL + link.Shortcode, 
		})
	}
}

// RedirectHandler gère la redirection d'une URL courte vers l'URL longue et l'enregistrement asynchrone des clics.
func RedirectHandler(linkService *services.LinkService, clickService *services.ClickService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupère le shortCode de l'URL avec c.Param
		shortCode := c.Param("shortCode")


		link, err := linkService.GetLinkByShortCode(shortCode)

		if err != nil {
			// Si le lien n'est pas trouvé, retourner HTTP 404 Not Found.
			// Utiliser errors.Is et l'erreur Gorm
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
				return
			}
			// Gérer d'autres erreurs potentielles de la base de données ou du service
			log.Printf("Error retrieving link for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		clickEvent := &models.ClickEvent{
			LinkID:    link.ID,
			Timestamp: time.Now(),
			UserAgent: c.Request.UserAgent(),
			IPAddress: c.ClientIP(),
		}

		

		select {
		case ClickEventsChannel <- clickEvent:
			// Si l'envoi est réussi, on continue
			click := &models.Click{
				LinkID:    clickEvent.LinkID,
				Timestamp: clickEvent.Timestamp,
				UserAgent: clickEvent.UserAgent,
				IPAddress: clickEvent.IPAddress,
			}
			clickService.RecordClick(click)
		default:
			log.Printf("Warning: ClickEventsChannel is full, dropping click event for %s.", shortCode)
		}

		c.Redirect(http.StatusFound, link.LongURL)
		log.Printf("Redirecting short code %s to long URL %s", shortCode, link.LongURL)
	}
}

// GetLinkStatsHandler gère la récupération des statistiques pour un lien spécifique.
func GetLinkStatsHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortCode := c.Param("shortCode")



		link, err := linkService.GetLinkByShortCode(shortCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
				return
			}
			log.Printf("Error retrieving link stats for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		_, totalClicks, err := linkService.GetLinkStats(shortCode)
		if err != nil {
			log.Printf("Error retrieving total clicks for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Retourne les statistiques dans la réponse JSON.
		c.JSON(http.StatusOK, gin.H{
			"short_code":   link.Shortcode,
			"long_url":     link.LongURL,
			"total_clicks": totalClicks,
		})
	}
}
