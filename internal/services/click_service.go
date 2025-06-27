package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository" // Importe le package repository
)

// Erreurs personnalisées pour le service
var (
	ErrInvalidClick     = errors.New("click data is invalid")
	ErrInvalidLinkID    = errors.New("link ID must be greater than 0")
	ErrInvalidTimestamp = errors.New("timestamp cannot be in the future")
	ErrEmptyUserAgent   = errors.New("user agent cannot be empty")
	ErrEmptyIPAddress   = errors.New("IP address cannot be empty")
)

// TODO : créer la struct
// ClickService est une structure qui fournit des méthodes pour la logique métier des clics.
// Elle est juste composer de clickRepo qui est de type ClickRepository
type ClickService struct {
	clickRepo repository.ClickRepository
}

// NewClickService crée et retourne une nouvelle instance de ClickService.
// C'est la fonction recommandée pour obtenir un service, assurant que toutes ses dépendances sont injectées.
func NewClickService(clickRepo repository.ClickRepository) *ClickService {
	return &ClickService{
		clickRepo: clickRepo,
	}
}

// RecordClick enregistre un nouvel événement de clic dans la base de données.
// Cette méthode est appelée par le worker asynchrone.
func (s *ClickService) RecordClick(click *models.Click) error {
	if click == nil {
		return fmt.Errorf("click service error: %w", ErrInvalidClick)
	}

	// Validation du LinkID
	if click.LinkID == 0 {
		return fmt.Errorf("click service error: %w", ErrInvalidLinkID)
	}

	// Validation du timestamp (ne peut pas être dans le futur)
	if click.Timestamp.After(time.Now()) {
		return fmt.Errorf("click service error: %w", ErrInvalidTimestamp)
	}

	// Validation du UserAgent
	if click.UserAgent == "" {
		return fmt.Errorf("click service error: %w", ErrEmptyUserAgent)
	}

	// Validation de l'IP
	if click.IPAddress == "" {
		return fmt.Errorf("click service error: %w", ErrEmptyIPAddress)
	}

	if err := s.clickRepo.CreateClick(click); err != nil {
		return fmt.Errorf("failed to record click for LinkID %d: %w", click.LinkID, err)
	}

	return nil
}

// GetClicksCountByLinkID récupère le nombre total de clics pour un LinkID donné.
// Cette méthode pourrait être utilisée par le LinkService pour les statistiques, ou directement par l'API stats.
func (s *ClickService) GetClicksCountByLinkID(linkID uint) (int, error) {
	// Validation du LinkID
	if linkID == 0 {
		return 0, fmt.Errorf("click service error: %w", ErrInvalidLinkID)
	}

	// (CountclicksByLinkID) pour compter les clics par LinkID.
	count, err := s.clickRepo.CountClicksByLinkID(linkID)
	if err != nil {
		return 0, fmt.Errorf("failed to count clicks for LinkID %d: %w", linkID, err)
	}

	return count, nil
}
