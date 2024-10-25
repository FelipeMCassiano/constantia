package bank

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/FelipeMCassiano/constantia/internal/domain"
)

var (
	NotSufficientBalanceError  = errors.New("not sufficient balance")
	SenderDoesNotExistsError   = errors.New("sender does not exists")
	RecieverDoesNotExistsError = errors.New("reciever does not exists")
	NoneTransaction            = errors.New("none transaction yet")
)

type Repository interface {
	CreateUser(user *domain.User) error
	LoginUser(user *domain.User) (int, error)
	CreateTransaction(transactionRequest *domain.Transaction) error
	GetLastTransactions(id int) ([]domain.Transaction, error)
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
		return SenderDoesNotExistsError
	}

	senderDetails.Scan(&senderBalance, &senderCPF)

	if senderBalance > transactionRequest.Value {
		return NotSufficientBalanceError
	}

	recieverDetails := tx.QueryRow("SELECT balance FROM users WHERE CPF=$1", transactionRequest.CPFReciever)
	if recieverDetails == nil {
		return RecieverDoesNotExistsError
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

func (r *repository) GetLastTransactions(id int) ([]domain.Transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, sender, cpf_sender, cpf_reciever, value, created_at FROM transactions WHERE id=$1 ", id)
	if err != nil {
		return nil, err
	}

	if rows != nil {
		defer rows.Close()

		var lastTransactions []domain.Transaction

		for rows.Next() {
			var transaction domain.Transaction

			if err := rows.Scan(&transaction.ID, &transaction.IDSender, &transaction.CPFSender, &transaction.CPFReciever, &transaction.Value, &transaction.CreatedAt); err != nil {
				return nil, err
			}

			lastTransactions = append(lastTransactions, transaction)

		}

		return lastTransactions, nil
	}

	return nil, NoneTransaction
}
