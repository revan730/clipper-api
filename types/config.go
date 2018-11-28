package types

// Config represents configuration for application
type Config struct {
	// Port to listen for requests
	Port          int
	DBAddr        string
	DB            string
	DBUser        string
	DBPassword    string
	RedisAddr     string
	RedisPassword string
	AdminLogin    string
	AdminPassword string
	JWTSecret     string
	RabbitAddress string
}
