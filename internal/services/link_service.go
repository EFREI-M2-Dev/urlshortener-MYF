package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm" // Nécessaire pour la gestion spécifique de gorm.ErrRecordNotFound

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository" // Importe le package repository
)

// Définition du jeu de caractères pour la génération des codes courts.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"


type LinkService struct {
	linkRepo repository.LinkRepository // Référence vers le repository de liens
}

// NewLinkService crée et retourne une nouvelle instance de LinkService.
func NewLinkService(linkRepo repository.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}


const shortCodeLength = 6 // Longueur du code court
func GenerateShortCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than 0")
	}

	code := make([]byte, length)
	for i := range code {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("error generating random index: %w", err)
		}
		code[i] = charset[index.Int64()]
	}

	return string(code), nil
}

// CreateLink crée un nouveau lien raccourci.
// Il génère un code court unique, puis persiste le lien dans la base de données.
func (s *LinkService) CreateLink(longURL string) (*models.Link, error) {
	// TODO 1: Implémenter la logique de retry pour générer un code court unique.
	// Essayez de générer un code, vérifiez s'il existe déjà en base, et retentez si une collision est trouvée.
	// Limitez le nombre de tentatives pour éviter une boucle infinie.



	const maxRetries = 5
	var shortCode string

	for i := 0; i < maxRetries; i++ {

		code, err := GenerateShortCode(shortCodeLength)
		if err != nil {
			return nil, fmt.Errorf("error generating short code: %w", err)
		}

	

		_, err = s.linkRepo.GetLinkByShortCode(code)

		if err != nil {
			// Si l'erreur est 'record not found' de GORM, cela signifie que le code est unique.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				shortCode = code // Le code est unique, on peut l'utiliser
				break            // Sort de la boucle de retry
			}
			// Si c'est une autre erreur de base de données, retourne l'erreur.
			return nil, fmt.Errorf("database error checking short code uniqueness: %w", err)
		}

		// Si aucune erreur (le code a été trouvé), cela signifie une collision.
		log.Printf("Short code '%s' already exists, retrying generation (%d/%d)...", code, i+1, maxRetries)
		// La boucle continuera pour générer un nouveau code.
	}


	if shortCode == "" {
		return nil, errors.New("failed to generate a unique short code after maximum retries")
	}

	link := &models.Link{
		LongURL:   longURL,
		Shortcode: shortCode,
		CreatedAt: time.Now(),
	}

	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, fmt.Errorf("error creating link in repository: %w", err)
	}

	return link, nil
}

// GetLinkByShortCode récupère un lien via son code court.
// Il délègue l'opération de recherche au repository.
func (s *LinkService) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	
	link, err := s.linkRepo.GetLinkByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("link with shortcode '%s' not found", shortCode)
		}
		return nil, fmt.Errorf("error retrieving link: %w", err)
	}
	return link, nil
}

// GetLinkStats récupère les statistiques pour un lien donné (nombre total de clics).
// Il interagit avec le LinkRepository pour obtenir le lien, puis avec le ClickRepository
func (s *LinkService) GetLinkStats(shortCode string) (*models.Link, int, error) {

	link, err := s.linkRepo.GetLinkByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, fmt.Errorf("link with shortcode '%s' not found", shortCode)
		}
		return nil, 0, fmt.Errorf("error retrieving link: %w", err)
	}


	clickCount, err := s.linkRepo.CountClicksByLinkID(link.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting clicks for link ID %d: %w", link.ID, err)
	}

	return link, clickCount, nil
}

