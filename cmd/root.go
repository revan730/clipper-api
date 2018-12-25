package cmd

import (
	"fmt"
	"os"

	"github.com/revan730/clipper-api/src"
	"github.com/revan730/clipper-api/log"
	"github.com/revan730/clipper-api/types"
	"github.com/spf13/cobra"
)

var (
	logVerbose bool
	serverPort int
	dbAddr     string
	db         string
	dbUser     string
	dbPass     string
	adminLogin string
	adminPass  string
	jwtSecret  string
	rabbitAddr string
	ciAddr     string
	cdAddr     string
)

var rootCmd = &cobra.Command{
	Use:   "clipper-api",
	Short: "REST API microservice of Clipper CI\\CD",
}

var serveCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		config := &types.Config{
			Port:          serverPort,
			DBAddr:        dbAddr,
			DB:            db,
			DBUser:        dbUser,
			DBPassword:    dbPass,
			AdminLogin:    adminLogin,
			AdminPassword: adminPass,
			JWTSecret:     jwtSecret,
			RabbitAddress: rabbitAddr,
			CIAddress:     ciAddr,
			CDAddress:     cdAddr,
		}
		logger := log.NewLogger(logVerbose)
		server := src.NewServer(logger, config).Routes()
		server.Run()
	},
}

// Execute runs application with provided cli params
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
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
	serveCmd.Flags().StringVarP(&adminLogin, "adminlogin", "l",
		"admin", "Set default admin login")
	serveCmd.Flags().StringVarP(&adminPass, "adminpass", "x",
		"admin", "Set default admin pass")
	serveCmd.Flags().StringVarP(&jwtSecret, "jwt", "j",
		"veryverysecret", "Set jwt secret")
	serveCmd.Flags().StringVarP(&rabbitAddr, "rabbitmq", "t",
		"amqp://guest:guest@localhost:5672", "Set rabbitmq address")
	serveCmd.Flags().StringVarP(&ciAddr, "ci", "g",
		"ci-worker:8080", "Set CI gRPC address")
	serveCmd.Flags().StringVarP(&cdAddr, "cd", "",
		"cd-worker:8080", "Set CD gRPC address")
}
