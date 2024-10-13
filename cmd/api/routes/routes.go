package routes

import (
	"database/sql"

	"github.com/FelipeMCassiano/constantia/cmd/api/handler"
	"github.com/FelipeMCassiano/constantia/internal/bank"
	"github.com/gofiber/fiber/v2"
)

type Router struct {
	eng *fiber.App
	rg  fiber.Router
	db  *sql.DB
}

func NewRouter(eng *fiber.App, db *sql.DB) *Router {
	return &Router{eng: eng, db: db}
}

func (r *Router) buildRoutes() {
	repo := bank.NewRepository(r.db)
	service := bank.NewService(repo)
	handler := handler.NewBank(service)

	r.rg = r.eng.Group("")

	// insert routers later

	r.rg.Post("/register", handler.RegisterUser())
}
