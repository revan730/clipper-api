package cmd

import (
	"fmt"
	"os"

	"github.com/revan730/diploma-server/src"
	"github.com/spf13/cobra"
)

var (
	logVerbose bool
	serverPort int
	dbAddr     string
	db         string
	dbUser     string
	dbPass     string
	redisAddr  string
	redisPass  string
	adminLogin string
	adminPass  string
	jwtSecret  string
)

var RootCmd = &cobra.Command{
	Use:   "clipper-server",
	Short: "Backend of Clipper CI\\CD service",
}

var serveCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		config := &src.Config{
			Port:          serverPort,
			DBAddr:        dbAddr,
			DB:            db,
			DBUser:        dbUser,
			DBPassword:    dbPass,
			RedisAddr:     redisAddr,
			RedisPassword: redisPass,
			AdminLogin:    adminLogin,
			AdminPassword: adminPass,
			JWTSecret:     jwtSecret,
		}
		logger := src.NewLogger(logVerbose)
		server := src.NewServer(logger, config).Routes()
		server.Run()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 8080,
		"Application TCP port")
	serveCmd.Flags().StringVarP(&dbAddr, "postgresAddr", "a",
		"postgres:5432", "Set PostsgreSQL address")
	serveCmd.Flags().StringVarP(&db, "db", "d",
		"clipper", "Set PostgreSQL database to use")
	serveCmd.Flags().StringVarP(&dbUser, "user", "u",
		"clipper", "Set PostgreSQL user to use")
	serveCmd.Flags().StringVarP(&dbPass, "pass", "c",
		"clipper", "Set PostgreSQL password to use")
	serveCmd.Flags().StringVarP(&redisAddr, "redis", "r",
		"redis:6379", "Set redis address")
	serveCmd.Flags().StringVarP(&redisPass, "redispass", "b",
		"", "Set redis address")
	serveCmd.Flags().StringVarP(&adminLogin, "adminlogin", "l",
		"admin", "Set default admin login")
	serveCmd.Flags().StringVarP(&adminPass, "adminpass", "x",
		"admin", "Set default admin pass")
	serveCmd.Flags().StringVarP(&jwtSecret, "jwt", "j",
		"veryverysecret", "Set jwt secret")
}
