package database

import (
	"bufio"
	"caslette-server/models"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	// Auto migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Diamond{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.UserPermission{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed default roles and permissions
	seedDefaultData(db)

	// Check if this is a fresh installation (no users exist)
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)

	if userCount == 0 {
		log.Println("No users found in database. Creating default admin user...")
		createDefaultAdmin(db)
	} else {
		// Check if admin user exists, if not create one
		var adminUser models.User
		result := db.Where("username = ?", "admin").First(&adminUser)
		if result.Error == gorm.ErrRecordNotFound {
			log.Println("Admin user not found. Creating default admin user...")
			createDefaultAdmin(db)
		}
	}

	log.Println("Database migration completed successfully")
}

func createDefaultAdmin(db *gorm.DB) {
	// Hash default password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create default admin user
	adminUser := models.User{
		Username:  "admin",
		Email:     "admin@caslette.com",
		Password:  string(hashedPassword),
		FirstName: "Admin",
		LastName:  "User",
		IsActive:  true,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Printf("Warning: Failed to create default admin user: %v", err)
		return
	}

	// Assign admin role
	var adminRole models.Role
	db.Where("name = ?", "admin").First(&adminRole)
	db.Model(&adminUser).Association("Roles").Append(&adminRole)

	// Create initial diamond balance (10000 for admin)
	diamond := models.Diamond{
		UserID:      adminUser.ID,
		Amount:      10000,
		Balance:     10000,
		Type:        "bonus",
		Description: "Admin welcome bonus",
		Metadata:    "{}",
	}
	db.Create(&diamond)

	log.Printf("✅ Default admin user created successfully!")
	log.Printf("📧 Username: admin")
	log.Printf("🔑 Password: admin123")
	log.Printf("💎 Starting diamond balance: 10,000")
}

func seedDefaultData(db *gorm.DB) {
	// Create default permissions
	permissions := []models.Permission{
		{Name: "user.create", Description: "Create users", Resource: "users", Action: "create"},
		{Name: "user.read", Description: "Read users", Resource: "users", Action: "read"},
		{Name: "user.update", Description: "Update users", Resource: "users", Action: "update"},
		{Name: "user.delete", Description: "Delete users", Resource: "users", Action: "delete"},
		{Name: "role.create", Description: "Create roles", Resource: "roles", Action: "create"},
		{Name: "role.read", Description: "Read roles", Resource: "roles", Action: "read"},
		{Name: "role.update", Description: "Update roles", Resource: "roles", Action: "update"},
		{Name: "role.delete", Description: "Delete roles", Resource: "roles", Action: "delete"},
		{Name: "diamond.read", Description: "Read diamond transactions", Resource: "diamonds", Action: "read"},
		{Name: "diamond.credit", Description: "Credit diamonds to users", Resource: "diamonds", Action: "credit"},
		{Name: "diamond.debit", Description: "Debit diamonds from users", Resource: "diamonds", Action: "debit"},
		{Name: "admin.access", Description: "Access admin dashboard", Resource: "admin", Action: "access"},
		{Name: "poker.table.create", Description: "Create poker tables", Resource: "poker", Action: "table_create"},
		{Name: "poker.table.delete", Description: "Delete poker tables", Resource: "poker", Action: "table_delete"},
	}

	for _, permission := range permissions {
		db.FirstOrCreate(&permission, models.Permission{Name: permission.Name})
	}

	// Create default roles
	adminRole := models.Role{Name: "admin", Description: "Administrator with full access"}
	db.FirstOrCreate(&adminRole, models.Role{Name: "admin"})

	userRole := models.Role{Name: "user", Description: "Regular user"}
	db.FirstOrCreate(&userRole, models.Role{Name: "user"})

	moderatorRole := models.Role{Name: "moderator", Description: "Moderator with limited admin access"}
	db.FirstOrCreate(&moderatorRole, models.Role{Name: "moderator"})

	// Assign permissions to admin role (all permissions)
	var allPermissions []models.Permission
	db.Find(&allPermissions)
	db.Model(&adminRole).Association("Permissions").Replace(allPermissions)

	// Assign basic permissions to user role
	var userPermissions []models.Permission
	db.Where("name IN ?", []string{"user.read", "diamond.read"}).Find(&userPermissions)
	db.Model(&userRole).Association("Permissions").Replace(userPermissions)

	// Assign moderate permissions to moderator role
	var moderatorPermissions []models.Permission
	db.Where("name IN ?", []string{
		"user.read", "user.update", "diamond.read", "diamond.credit", "diamond.debit", "admin.access", "poker.table.create",
	}).Find(&moderatorPermissions)
	db.Model(&moderatorRole).Association("Permissions").Replace(moderatorPermissions)

	log.Println("Default roles and permissions seeded successfully")
}

func createSuperuser(db *gorm.DB) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🎰 Welcome to Caslette Casino!")
	fmt.Println("Let's create your first administrator account.")
	fmt.Println("========================================")

	// Get username
	fmt.Print("Enter admin username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		username = "admin"
		fmt.Printf("Using default username: %s\n", username)
	}

	// Get email
	fmt.Print("Enter admin email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if email == "" {
		email = "admin@caslette.com"
		fmt.Printf("Using default email: %s\n", email)
	}

	// Get password
	fmt.Print("Enter admin password: ")
	password := getPassword()

	if password == "" {
		password = "admin"
		fmt.Println("Using default password: admin")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create superuser
	superuser := models.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: "Super",
		LastName:  "Admin",
		IsActive:  true,
	}

	if err := db.Create(&superuser).Error; err != nil {
		log.Fatal("Failed to create superuser:", err)
	}

	// Assign admin role
	var adminRole models.Role
	db.Where("name = ?", "admin").First(&adminRole)
	db.Model(&superuser).Association("Roles").Append(&adminRole)

	// Create initial diamond balance (10000 for admin)
	diamond := models.Diamond{
		UserID:      superuser.ID,
		Amount:      10000,
		Balance:     10000,
		Type:        "bonus",
		Description: "Admin welcome bonus",
		Metadata:    "{}",
	}
	db.Create(&diamond)

	fmt.Printf("\n✅ Superuser '%s' created successfully!\n", username)
	fmt.Printf("📧 Email: %s\n", email)
	fmt.Printf("💎 Starting diamond balance: 10,000\n")
	fmt.Println("🚀 You can now access the admin dashboard at http://localhost:5177/")
	fmt.Println("========================================\n")
}

func getPassword() string {
	fmt.Print("(Password will be visible - press Enter for default 'admin'): ")
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(password)
}
