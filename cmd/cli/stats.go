package cli

import (
	"fmt"
	"log"
	"os"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"

	"gorm.io/driver/sqlite" 
	"gorm.io/gorm"
)

var shortCodeFlag string

var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Affiche les statistiques (nombre de clics) pour un lien court.",
	Long: `Cette commande permet de récupérer et d'afficher le nombre total de clics
pour une URL courte spécifique en utilisant son code.

Exemple:
  url-shortener stats --code="xyz123"`,
	Run: func(cmd *cobra.Command, args []string) {
		if shortCodeFlag == "" {
            fmt.Fprintln(os.Stderr, "ERREUR: Le flag --code est requis.")
            os.Exit(1)
        }


		cfg := cmd2.Cfg
        if cfg == nil {
            fmt.Fprintln(os.Stderr, "ERREUR: Impossible de charger la configuration globale.")
            os.Exit(1)
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

		linkRepo := repository.NewLinkRepository(db)
        linkService := services.NewLinkService(linkRepo)

		link, totalClicks, err := linkService.GetLinkStats(shortCodeFlag)
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                fmt.Fprintf(os.Stderr, "Aucun lien trouvé pour le code court: %s\n", shortCodeFlag)
            } else {
                fmt.Fprintf(os.Stderr, "Erreur lors de la récupération des statistiques: %v\n", err)
            }
            os.Exit(1)
        }


		fmt.Printf("Statistiques pour le code court: %s\n", link.Shortcode)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("Total de clics: %d\n", totalClicks)
	},
}

func init() {
	StatsCmd.Flags().StringVar(&shortCodeFlag, "code", "", "Code court du lien à analyser")
	
	StatsCmd.MarkFlagRequired("code")

	cmd2.RootCmd.AddCommand(StatsCmd)
}
