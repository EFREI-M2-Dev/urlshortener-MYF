package cli

import (
	"fmt"
	"log"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Exécute les migrations de la base de données pour créer ou mettre à jour les tables.",
	Long: `Cette commande se connecte à la base de données configurée (SQLite)
et exécute les migrations automatiques de GORM pour créer les tables 'links' et 'clicks'
basées sur les modèles Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := cmd2.Cfg
        if cfg == nil {
            log.Fatalf("FATAL: Impossible de charger la configuration globale.")
        }

		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
        if err != nil {
            log.Fatalf("FATAL: Impossible d'ouvrir la base de données: %v", err)
        }

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}
		
		defer sqlDB.Close()

		err = db.AutoMigrate(&models.Link{}, &models.Click{})
        if err != nil {
            log.Fatalf("FATAL: Erreur lors de la migration: %v", err)
        }

		fmt.Println("Migrations de la base de données exécutées avec succès.")
	},
}

func init() {
	cmd2.RootCmd.AddCommand(MigrateCmd)
}
