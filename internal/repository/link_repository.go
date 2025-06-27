package repository

import (
	"log"

	"github.com/axellelanca/urlshortener/internal/models"
	"gorm.io/gorm"
)

// LinkRepository est une interface qui définit les méthodes d'accès aux données
// pour les opérations CRUD sur les liens.
type LinkRepository interface {
	CreateLink(link *models.Link) error
	GetLinkByShortCode(shortCode string) (*models.Link, error)
	GetAllLinks() ([]models.Link, error)
	CountClicksByLinkID(linkID uint) (int, error)
}

// TODO :  GormLinkRepository est l'implémentation de LinkRepository utilisant GORM.
type GormLinkRepository struct {
	db *gorm.DB
}

// NewLinkRepository crée et retourne une nouvelle instance de GormLinkRepository.
// Cette fonction retourne *GormLinkRepository, qui implémente l'interface LinkRepository.
func NewLinkRepository(db *gorm.DB) *GormLinkRepository {
	// TODO
	return &GormLinkRepository{db: db}
}

// CreateLink insère un nouveau lien dans la base de données.
func (r *GormLinkRepository) CreateLink(link *models.Link) error {
	// TODO 1: Utiliser GORM pour créer un nouvel enregistrement (link) dans la table des liens.
	if err := r.db.Create(link).Error; err != nil {
		log.Printf("Erreur lors de la création du lien: %v", err)
	} else {
		log.Printf("Lien créé avec succès: %s", link.Shortcode)
	}
	return nil
}

// GetLinkByShortCode récupère un lien de la base de données en utilisant son shortCode.
// Il renvoie gorm.ErrRecordNotFound si aucun lien n'est trouvé avec ce shortCode.
func (r *GormLinkRepository) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	var link models.Link
	// TODO 2: Utiliser GORM pour trouver un lien par son ShortCode.
	// La méthode First de GORM recherche le premier enregistrement correspondant et le mappe à 'link'.
	if err := r.db.Where("shortcode = ?", shortCode).First(&link).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Aucun lien trouvé pour le shortCode: %s", shortCode)
			return nil, err 
		}
		log.Printf("Erreur lors de la récupération du lien: %v", err)
		return nil, err 
	}
	log.Printf("Lien récupéré avec succès: %s", link.Shortcode)
	return &link, nil 
}

// GetAllLinks récupère tous les liens de la base de données.
// Cette méthode est utilisée par le moniteur d'URLs.
func (r *GormLinkRepository) GetAllLinks() ([]models.Link, error) {
	var links []models.Link
	// TODO 3: Utiliser GORM pour récupérer tous les liens.
	if err := r.db.Find(&links).Error; err != nil {
		log.Printf("Erreur lors de la récupération des liens: %v", err)
		return nil, err 
	}
	return links, nil
}

// CountClicksByLinkID compte le nombre total de clics pour un ID de lien donné.
func (r *GormLinkRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64 // GORM retourne un int64 pour les comptes
	// TODO 4: Utiliser GORM pour compter les enregistrements dans la table 'clicks'
	// où 'LinkID' correspond à l'ID du lien donné.

	if err := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count).Error; err != nil {
		log.Printf("Erreur lors du comptage des clics pour le lien ID %d: %v", linkID, err)
		return 0, err
	}

	return int(count), nil
}
