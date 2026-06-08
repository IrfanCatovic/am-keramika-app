package database

var DB *gorm.DB //Global database connection

func ConnectDB() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
	os.Getenv("DB_HOST"), 
	os.Getenv("DB_PORT"), 
	os.Getenv("DB_USER"), 
	os.Getenv("DB_PASSWORD"), 
	os.Getenv("DB_NAME"),
)

database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err != nil {
	log.Fatal("Neuspjela konekcija na bazu: ", err)
}

DB = database
log.Println("Uspješna konekcija na bazu")
}

