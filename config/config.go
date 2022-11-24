package config

type Config struct {
	Mode          string
	ApiListen     string
	DatabaseURL   string
	SecretKey     string
	RedisAddr     string
	RedisPassword string
}
