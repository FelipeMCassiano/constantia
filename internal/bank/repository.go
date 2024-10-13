package bank

import (
	"database/sql"
	"fmt"

	"github.com/FelipeMCassiano/constantia/internal/domain"
)

type Repository interface {
	CreateUser(user *domain.User) error
	LoginUser(user *domain.User) (int, error)
	CreateTransaction(transactionRequest *domain.Transaction) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(user *domain.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO users (name, surname, CPF, balance) VALUES ($1,$2,$3,$4) RETURNING name", user.Name, user.Surname, user.CPF, user.Balance)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) LoginUser(user *domain.User) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var dbUser domain.User

	row := tx.QueryRow("SELECT id, name, surname FROM users WHERE CPF = $1", user.CPF)
	if row == nil {
		return 0, fmt.Errorf("user does not exists")
	}

	row.Scan(&dbUser.ID, &dbUser.Name, &dbUser.Surname, &dbUser.Balance)

	hasDifferentBetweenUsers := dbUser.Name != user.Name || dbUser.Surname != user.Surname
	if hasDifferentBetweenUsers {
		return 0, fmt.Errorf("Invalid credentials")
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return dbUser.ID, nil
}

func (r *repository) CreateTransaction(transactionRequest *domain.Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var senderBalance int
	var senderCPF string

	senderDetails := tx.QueryRow("SELECT balance, CPF FROM users WHERE id=$1", transactionRequest.IDSender)

	if senderDetails == nil {
		return fmt.Errorf("sender does not exists")
	}

	senderDetails.Scan(&senderBalance, &senderCPF)

	if senderBalance > transactionRequest.Value {
		return fmt.Errorf("not suficient balance")
	}

	recieverDetails := tx.QueryRow("SELECT balance FROM users WHERE CPF=$1", transactionRequest.CPFReciever)
	if recieverDetails == nil {
		return fmt.Errorf("reciever does not exists")
	}

	recieverCPF := transactionRequest.CPFReciever
	var recieverBalance int

	recieverDetails.Scan(&recieverBalance)

	transactionValue := transactionRequest.Value

	if _, err := tx.Exec("INSERT INTO transactions (senderCPF, recieverCPF, value) VALUES ($1,$2,$3)", senderCPF, recieverCPF, transactionValue); err != nil {
		return err
	}

	newSenderBalance := senderBalance - transactionValue

	if _, err := tx.Exec("UPDATE users SET balance=$1 WHERE CPF=$2", newSenderBalance, senderCPF); err != nil {
		return err
	}

	newRecieverBalance := recieverBalance + transactionValue
	if _, err := tx.Exec("UPDATE users SET balance=$1 WHERE CPF=$2", newRecieverBalance, recieverCPF); err != nil {
		return err
	}

	return tx.Commit()
}
