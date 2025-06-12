package scheduler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"flash-sale-service/internal/database"
	"flash-sale-service/internal/models"
	redisClient "flash-sale-service/internal/redis"
)

type Scheduler struct {
	db    *database.DB
	redis *redisClient.Client
}

// NewScheduler creates a new scheduler instance
func NewScheduler(db *database.DB, redis *redisClient.Client) *Scheduler {
	return &Scheduler{
		db:    db,
		redis: redis,
	}
}

// generateSaleID generates a unique sale ID
func generateSaleID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	timestamp := time.Now().Unix()
	return fmt.Sprintf("sale_%d_%s", timestamp, hex.EncodeToString(bytes)), nil
}

// generateItemID generates a unique item ID
func generateItemID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("item_%s", hex.EncodeToString(bytes)), nil
}

// Item name templates for variety
var itemNameTemplates = []string{
	"Premium %s Collection",
	"Limited Edition %s",
	"Exclusive %s Series",
	"Designer %s Set",
	"Luxury %s Bundle",
	"Special %s Edition",
	"Artisan %s Collection",
	"Elite %s Package",
	"Signature %s Line",
	"Master %s Series",
}

var itemCategories = []string{
	"Smartphone", "Laptop", "Headphones", "Watch", "Camera", "Tablet", "Speaker", "Gaming Console",
	"Fitness Tracker", "Smart TV", "Keyboard", "Mouse", "Monitor", "Printer", "Router", "Drone",
	"VR Headset", "Smart Home Hub", "Wireless Charger", "Power Bank", "Bluetooth Earbuds", "Webcam",
	"External Drive", "Graphics Card", "Processor", "Memory Card", "USB Cable", "Phone Case",
	"Screen Protector", "Car Charger", "Desk Lamp", "Office Chair", "Backpack", "Wallet",
	"Sunglasses", "Perfume", "Skincare Set", "Makeup Kit", "Hair Dryer", "Electric Toothbrush",
	"Coffee Maker", "Blender", "Air Fryer", "Vacuum Cleaner", "Humidifier", "Air Purifier",
	"Yoga Mat", "Dumbbells", "Resistance Bands", "Protein Powder", "Water Bottle", "Running Shoes",
}

var colorVariants = []string{
	"Black", "White", "Silver", "Gold", "Rose Gold", "Blue", "Red", "Green", "Purple", "Pink",
	"Gray", "Bronze", "Copper", "Titanium", "Space Gray", "Midnight", "Starlight", "Ocean Blue",
	"Forest Green", "Sunset Orange", "Deep Purple", "Coral", "Mint", "Lavender", "Crimson",
}

// generateItemName generates a random item name
func generateItemName() (string, error) {
	// Select random template
	templateIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(itemNameTemplates))))
	if err != nil {
		return "", err
	}
	template := itemNameTemplates[templateIndex.Int64()]

	// Select random category
	categoryIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(itemCategories))))
	if err != nil {
		return "", err
	}
	category := itemCategories[categoryIndex.Int64()]

	// Select random color
	colorIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(colorVariants))))
	if err != nil {
		return "", err
	}
	color := colorVariants[colorIndex.Int64()]

	// Combine category and color
	categoryWithColor := fmt.Sprintf("%s %s", color, category)

	return fmt.Sprintf(template, categoryWithColor), nil
}

// generateImageURL generates a placeholder image URL
func generateImageURL(itemID string) string {
	// Use a placeholder image service with item-specific parameters
	width := 400
	height := 400
	
	// Generate a seed based on item ID for consistent images
	seed := 0
	for _, char := range itemID {
		seed += int(char)
	}
	
	return fmt.Sprintf("https://picsum.photos/seed/%d/%d/%d", seed, width, height)
}

// generateItems generates the specified number of items for a sale
func (s *Scheduler) generateItems(saleID string, count int) ([]models.Item, error) {
	items := make([]models.Item, count)
	
	for i := 0; i < count; i++ {
		itemID, err := generateItemID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate item ID: %w", err)
		}

		itemName, err := generateItemName()
		if err != nil {
			return nil, fmt.Errorf("failed to generate item name: %w", err)
		}

		imageURL := generateImageURL(itemID)

		items[i] = models.Item{
			ItemID:   itemID,
			SaleID:   saleID,
			Name:     itemName,
			ImageURL: imageURL,
		}
	}

	return items, nil
}

// createNewSale creates a new flash sale with items
func (s *Scheduler) createNewSale() error {
	log.Println("Creating new flash sale...")

	// Generate sale ID
	saleID, err := generateSaleID()
	if err != nil {
		return fmt.Errorf("failed to generate sale ID: %w", err)
	}

	// Calculate sale times
	now := time.Now()
	startTime := now.Truncate(time.Hour)
	endTime := startTime.Add(time.Hour)

	// Create sale record
	sale := &models.Sale{
		SaleID:     saleID,
		StartTime:  startTime,
		EndTime:    endTime,
		TotalItems: models.ItemsPerSale,
		ItemsSold:  0,
		Status:     models.SaleStatusActive,
	}

	// Save sale to database
	if err := s.db.CreateSale(sale); err != nil {
		return fmt.Errorf("failed to create sale in database: %w", err)
	}

	// Generate items
	items, err := s.generateItems(saleID, models.ItemsPerSale)
	if err != nil {
		return fmt.Errorf("failed to generate items: %w", err)
	}

	// Save items to database
	if err := s.db.CreateItems(items); err != nil {
		return fmt.Errorf("failed to create items in database: %w", err)
	}

	// Initialize sale in Redis
	if err := s.redis.InitializeSale(saleID, startTime, endTime); err != nil {
		return fmt.Errorf("failed to initialize sale in Redis: %w", err)
	}

	log.Printf("Successfully created sale %s with %d items", saleID, len(items))
	return nil
}

// cleanupExpiredSales marks expired sales as completed
func (s *Scheduler) cleanupExpiredSales() error {
	// This would typically update sales that have passed their end time
	// For now, we'll implement a simple cleanup of expired checkout sessions
	count, err := s.redis.CleanupExpiredCheckouts()
	if err != nil {
		return fmt.Errorf("failed to cleanup expired checkouts: %w", err)
	}

	if count > 0 {
		log.Printf("Cleaned up %d expired checkout sessions", count)
	}

	return nil
}

// waitUntilNextHour waits until the next hour boundary
func waitUntilNextHour() time.Duration {
	now := time.Now()
	nextHour := now.Truncate(time.Hour).Add(time.Hour)
	return nextHour.Sub(now)
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	log.Println("Starting flash sale scheduler...")

	// Create initial sale if none exists
	activeSale, err := s.db.GetActiveSale()
	if err != nil {
		return fmt.Errorf("failed to check for active sale: %w", err)
	}

	if activeSale == nil {
		log.Println("No active sale found, creating initial sale...")
		if err := s.createNewSale(); err != nil {
			return fmt.Errorf("failed to create initial sale: %w", err)
		}
	} else {
		log.Printf("Found active sale: %s", activeSale.SaleID)
	}

	// Wait until the next hour boundary for the first scheduled sale
	waitDuration := waitUntilNextHour()
	log.Printf("Waiting %v until next scheduled sale creation", waitDuration)

	select {
	case <-time.After(waitDuration):
		// Continue to main loop
	case <-ctx.Done():
		log.Println("Scheduler stopped before first scheduled sale")
		return ctx.Err()
	}

	// Main scheduler loop - create new sale every hour
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	// Cleanup ticker - run every 15 minutes
	cleanupTicker := time.NewTicker(15 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.createNewSale(); err != nil {
				log.Printf("Failed to create new sale: %v", err)
				// Continue running even if one sale creation fails
			}

		case <-cleanupTicker.C:
			if err := s.cleanupExpiredSales(); err != nil {
				log.Printf("Failed to cleanup expired sales: %v", err)
				// Continue running even if cleanup fails
			}

		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return ctx.Err()
		}
	}
}

func StartScheduler(db *sql.DB) {
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for {
            <-ticker.C
            createNewSale(db)
        }
    }()
}

func createNewSale(db *sql.DB) {
    startTime := time.Now().Truncate(time.Hour)
    endTime := startTime.Add(1 * time.Hour)

    _, err := db.Exec(`
        INSERT INTO sales (start_time, end_time, total_items)
        VALUES ($1, $2, $3)
    `, startTime, endTime, 10000)

    if err != nil {
        log.Printf("Error creating new sale: %v", err)
        return
    }

    log.Printf("New sale created from %v to %v", startTime, endTime)

    // Generate 10,000 items for this sale
    generateItems(db, startTime)
}

func generateItems(db *sql.DB, saleStartTime time.Time) {
    // Implementation for generating 10,000 unique items
    // This is a placeholder and should be implemented with actual logic
    log.Println("Generating 10,000 items for the new sale")
}
