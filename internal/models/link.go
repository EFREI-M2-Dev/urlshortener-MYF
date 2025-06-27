package models

import "time"




type Link struct {
    ID        uint      `gorm:"primaryKey"`                          
    Shortcode string    `gorm:"size:10;uniqueIndex;not null"`        
    LongURL   string    `gorm:"not null"`                           
    CreatedAt time.Time
}