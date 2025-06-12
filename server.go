package main

import (
	"flag"
	"fmt"
	"github.com/headliner38/graphql-forum/graph/storage"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/headliner38/graphql-forum/graph"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
		fmt.Println("Используется порт по умолчанию(8080), чтобы использовать другой порт задайте переменную окружения PORT")
	}

	storageType := flag.String("storage", "memory", "Тип хранилища: memory или postgres")
	pgConnStr := flag.String("pg-conn", "postgres://user:3276@localhost:5433/postgres?sslmode=disable", "PostgreSQL connection string")
	flag.Parse()

	var repo storage.Repository

	switch *storageType {
	case "postgres":
		pgRepo, err := storage.NewPostgresRepository(*pgConnStr)
		if err != nil {
			log.Fatalf("Failed to init PostgreSQL: %v", err)
		}
		repo = pgRepo
	default:
		repo = storage.NewMemoryRepository()
		log.Println("Using in-memory storage")
	}

	cfg := graph.DepthAndLengthConfig{
		MaxCommentDepth:  10, // подумать над реализацей через переменные окружения
		MaxCommentLength: 2000,
	}

	resolver, err := graph.NewResolver(cfg, repo)
	if err != nil {
		log.Fatalf("Failed to create resolver: %v", err)
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
	}))

	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // разрешение на все соединения для тестов
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
