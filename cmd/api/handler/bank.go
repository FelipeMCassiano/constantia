package handler

import (
	"time"

	"github.com/FelipeMCassiano/constantia/internal/bank"
	"github.com/FelipeMCassiano/constantia/internal/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type BankController struct {
	bankService bank.Service
}

func NewBank(s bank.Service) *BankController {
	return &BankController{
		bankService: s,
	}
}

// change this later
const secretKey = "secrete"

func (b *BankController) RegisterUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		request := new(domain.User)
		if err := c.BodyParser(request); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		err := b.bankService.RegisterUser(request)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		return c.SendStatus(fiber.StatusCreated)
	}
}

func (b *BankController) LoginUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		request := new(domain.User)
		if err := c.BodyParser(request); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		identifier, err := b.bankService.LoginUser(request)

		claims := jwt.MapClaims{
			"identifier": identifier,
			"exp":        time.Now().Add(time.Hour * 72).UTC(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, err := token.SignedString([]byte(secretKey))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": t})
	}
}

func getIdentifier(c *fiber.Ctx) int {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	identifier := claims["identifier"].(int)
	return identifier
}
