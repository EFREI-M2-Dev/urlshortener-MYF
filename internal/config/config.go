package config

import (
	"log" // Pour logger les informations ou erreurs de chargement de config

	"github.com/spf13/viper" // La bibliothèque pour la gestion de configuration
)


type Config struct {
	Server struct {
		Port     int    `mapstructure:"port"`
		BaseURL  string `mapstructure:"base_url"`
	} `mapstructure:"server"`

	Database struct {
		Name string `mapstructure:"name"`
	} `mapstructure:"database"`

	Analytics struct {
		BufferSize int `mapstructure:"buffer_size"`
		WorkerCount int `mapstructure:"worker_count"`
	} `mapstructure:"analytics"`

	Monitor struct {
		IntervalMinutes int `mapstructure:"interval_minutes"`
	} `mapstructure:"monitor"`

	Workers struct {
		ClickEventsBufferSize int `mapstructure:"click_events_buffer_size"`
	} `mapstructure:"workers"`
}

// LoadConfig charge la configuration de l'application en utilisant Viper.
// Elle recherche un fichier 'config.yaml' dans le dossier 'configs/'.
// Elle définit également des valeurs par défaut si le fichier de config est absent ou incomplet.
func LoadConfig() (*Config, error) {


	viper.AddConfigPath("./configs") // Chemin relatif au répertoire d'exécution


	viper.SetConfigName("config") // Nom du fichier de config sans l'extension

	viper.SetConfigType("yaml")

	

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.base_url", "http://localhost:8080")
	viper.SetDefault("database.name", "urlshortener.db")
	viper.SetDefault("analytics.buffer_size", 1000)
	viper.SetDefault("analytics.worker_count", 4)
	viper.SetDefault("monitor.interval_minutes", 5)


	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Configuration file not found or error reading it: %v. Using default values.", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Printf("Error unmarshalling configuration: %v", err)
		return nil, err
	}

	// Log  pour vérifier la config chargée
	log.Printf("Configuration loaded: Server Port=%d, DB Name=%s, Analytics Buffer=%d, Monitor Interval=%dmin",
		cfg.Server.Port, cfg.Database.Name, cfg.Analytics.BufferSize, cfg.Monitor.IntervalMinutes)

	return &cfg, nil // Retourne la configuration chargée
}
