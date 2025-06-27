package cli

import (
	"fmt"
	"log"
	"net/url"
	"os"


	"github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/config"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var longURLFlag string

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie et affiche le code court généré.

Exemple:
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cmd *cobra.Command, args []string) {
		if longURLFlag == "" {
			log.Fatal("FATAL: Le flag --url est requis.")
		}

		if _, err := url.ParseRequestURI(longURLFlag); err != nil {
			log.Printf("FATAL: URL invalide: %v", err)
			os.Exit(1)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("FATAL: Échec du chargement de la configuration: %v", err)
		}

		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec de la connexion à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}

		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Erreur lors de la fermeture de la base de données: %v", err)
			}
		}()

		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		link, err := linkService.CreateLink(longURLFlag)
		if err != nil {
			log.Printf("FATAL: Échec de la création du lien: %v", err)
			os.Exit(1)
		}

		fullShortURL := fmt.Sprintf("%s/%s", cfg.Server.BaseURL, link.Shortcode)
		fmt.Printf("URL courte créée avec succès:\n")
		fmt.Printf("Code: %s\n", link.Shortcode)
		fmt.Printf("URL complète: %s\n", fullShortURL)
	},
}

func init() {
	CreateCmd.Flags().StringVar(&longURLFlag, "url", "", "URL longue à raccourcir")

	CreateCmd.MarkFlagRequired("url")

	cmd.RootCmd.AddCommand(CreateCmd)
}
