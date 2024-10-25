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
			"exp":        time.Now().Add(time.Hour * 72).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, err := token.SignedString([]byte(secretKey))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": t})
	}
}

func (b *BankController) CreateTransaction() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if identifier := getIdentifier(c); identifier == 0 {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		var transactionRequest *domain.Transaction
		if err := c.BodyParser(transactionRequest); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(err.Error())
		}

		err := b.bankService.CreateTransaction(transactionRequest)
		if err != nil {
			switch err {
			case bank.NotSufficientBalanceError:
				return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
			case bank.RecieverDoesNotExistsError:
				return c.Status(fiber.StatusBadRequest).JSON(err.Error())
			case bank.SenderDoesNotExistsError:
				return c.Status(fiber.StatusBadRequest).JSON(err.Error())

			default:
				return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
			}
		}

		return c.SendStatus(fiber.StatusCreated)
	}
}

func (b *BankController) GetLastTransactions() fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := getIdentifier(c)

		transactions, err := b.bankService.GetLastTransaction(identifier)
		if err != nil {
			switch err {
			case bank.NoneTransaction:
				return c.Status(fiber.StatusNoContent).JSON(err.Error())
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(err.Error())

			}
		}

		return c.Status(fiber.StatusOK).JSON(transactions)
	}
}

func getIdentifier(c *fiber.Ctx) int {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	identifier := claims["identifier"].(int)
	return identifier
}
