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
	RedisPw   string
	RedisDb   int
	Port      int
)

func init() {
	flag.StringVar(&RootDir, "root", "", "Root directory to serve")
	flag.StringVar(&RedisAddr, "redis-addr", "localhost:6379", "Redis hostname:port")
	flag.StringVar(&RedisPw, "redis-pw", "", "Redis password")
	flag.IntVar(&RedisDb, "redis-db", 0, "Redis database (0-15)")
	flag.IntVar(&Port, "port", 8080, "HTTP port")
	flag.Parse()

	log.SetOutput(os.Stdout)
}

func main() {
	srv := beepboop.NewServer()
	srv.AddMiddlewares(AuthMiddleware(RootDir))
	srv.AddPages(DirectoryPage(RootDir), AuthPage(RootDir))

	if err := srv.ConnectDB(RedisAddr, RedisPw, RedisDb); err != nil {
		log.Print(err)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Port), srv))
}
