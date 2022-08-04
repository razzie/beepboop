package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/razzie/beepboop"
)

// command line args
var (
	RootDir   string
	RedisAddr string
	Port      int
)

func init() {
	flag.StringVar(&RootDir, "root", "", "Root directory to serve")
	flag.StringVar(&RedisAddr, "redis", "redis://localhost:6379", "Redis connection string")
	flag.IntVar(&Port, "port", 8080, "HTTP port")
	flag.Parse()

	log.SetOutput(os.Stdout)
}

func main() {
	srv := beepboop.NewServer()
	srv.AddMiddlewares(AuthMiddleware(RootDir))
	srv.AddPages(DirectoryPage(RootDir), AuthPage(RootDir))

	if err := srv.ConnectDB(RedisAddr); err != nil {
		log.Print(err)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Port), srv))
}
