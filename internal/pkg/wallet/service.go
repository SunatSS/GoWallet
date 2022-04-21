package wallet

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/SYSTEMTerror/GoWallet/internal/pkg/types"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("account not found")
	ErrExist           = errors.New("account already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrOutOfLimit      = errors.New("out of limit")
	ErrInternal        = errors.New("internal error")
	ErrExpired         = errors.New("expired")
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Exist checks if account with given phone exists. Returns false and nil or true and account
func (s *Service) Exist(ctx context.Context, phone string) (bool, *types.Account, int,  error) {
	acc := &types.Account{Phone: phone}
	err := s.pool.QueryRow(ctx, `SELECT id, balance, identified, name,  password, active, created FROM accounts WHERE phone = $1`, acc.Phone).Scan(&acc.ID, &acc.Balance, &acc.Identified, &acc.Username, &acc.Password, &acc.Active, &acc.Created)
	if err == pgx.ErrNoRows {
		return false, nil, http.StatusOK, nil
	}
	if err != nil {
		log.Println("Exist s.pool.QueryRow error:", err)
		return false, nil, http.StatusInternalServerError, ErrInternal
	}

	return true, acc, http.StatusOK, nil
}

func (s *Service) Register(ctx context.Context, item *types.RegInfo) (*types.Account, int, error) {
	acc := &types.Account{
		Balance:    0,
		Identified: false,
		Username:   item.Username,
		Phone:      item.Phone,
	}
	exist, _, _, err := s.Exist(ctx, item.Phone)
	if err != nil {
		log.Println("Register s.Exist error:", err)
		return nil, http.StatusInternalServerError, ErrInternal
	}
	if exist {
		log.Println("Register s.Exist account already exist")
		return nil, http.StatusConflict, ErrExist
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Register bcrypt.GenerateFromPassword Error:", err)
		return nil, http.StatusInternalServerError, ErrInternal
	}

	item.Password = string(hash)
	err = s.pool.QueryRow(ctx, `INSERT INTO accounts (name, phone, password) VALUES ($1, $2, $3) RETURNING id, active, created`, item.Username, item.Phone, item.Password).Scan(&acc.ID, &acc.Active, &acc.Created)
	if err != nil {
		log.Println("Register s.pool.QueryRow error:", err)
		return nil, http.StatusInternalServerError, ErrInternal
	}

	return acc, http.StatusOK, nil
}

// Transfer transfers money to/from account depending on sign of amount
func (s *Service) Transaction(ctx context.Context, item *types.Transaction) (*types.Transaction, int, error) {
	var limit int64
	acc, _, err := s.GetAccountByID(ctx, item.AccID)
	if err != nil {
		log.Println("Transaction s.GetAccountByID error:", err)
		return nil, http.StatusInternalServerError, ErrInternal
	}
	if !acc.Identified {
		limit = 10_000_00 // Dirams
	} else {
		limit = 100_000_00 // Dirams
	}

	if acc.Balance+item.Amount < 0 || acc.Balance+item.Amount>limit {
		log.Println("Transaction s.GetAccountByID error:", ErrOutOfLimit)
		return nil, http.StatusBadRequest, ErrOutOfLimit
	}

	err = s.pool.QueryRow(ctx, `INSERT INTO transactions (acc_id, amount) VALUES ($1, $2) RETURNING id, created`, item.AccID, item.Amount).Scan(&item.ID, &item.Created)
	if err != nil {
		log.Println("Transaction s.pool.QueryRow error:", err)
		return nil, http.StatusInternalServerError, ErrInternal
	}

	_, err = s.pool.Exec(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, item.Amount, item.AccID)
	if err != nil {
		log.Println("Transaction s.pool.QueryRow error:", err)
		return nil, http.StatusInternalServerError, ErrInternal
	}
	return item, http.StatusOK, nil
}

//GetTransactionsPerMonth returns transactions, sum of transactions and amount of transactions per month last month
func (s *Service) GetTransactionsPerMonth(ctx context.Context, accID int64) ([]*types.Transaction, int64, int64, int, error) {
	transactions := []*types.Transaction{}
	var sum int64
	var count int64

	rows, err := s.pool.Query(ctx, `SELECT id, acc_id, amount, created FROM transactions WHERE acc_id = $1 AND to_char(created, 'Mon') = to_char(current_date, 'Mon')`, accID)
	if err != nil {
		log.Println("GetTransactionsPerMonth s.pool.Query error:", err)
		return nil, 0, 0, http.StatusInternalServerError, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		var transaction types.Transaction
		err = rows.Scan(&transaction.ID, &transaction.AccID, &transaction.Amount, &transaction.Created)
		if err != nil {
			log.Println("GetTransactionsPerMonth rows.Scan error:", err)
			return nil, 0, 0, http.StatusInternalServerError, ErrInternal
		}
		transactions = append(transactions, &transaction)
		sum += transaction.Amount
		count++
	}

	return transactions, sum, count, http.StatusOK, nil
}

func (s *Service) GetAccountByID(ctx context.Context, id int64) (*types.Account, int, error) {
	acc := &types.Account{}
	err := s.pool.QueryRow(ctx, `SELECT id, balance, identified, name, phone, password, active, created FROM accounts WHERE id = $1`, id).Scan(&acc.ID, &acc.Balance, &acc.Identified, &acc.Username, &acc.Phone, &acc.Password, &acc.Active, &acc.Created)
	if err == pgx.ErrNoRows {
		log.Println("GetAccountByID s.pool.QueryRow no rows:", err)
		return nil, http.StatusNotFound, ErrNotFound
	}
	if err != nil {
		log.Println("GetAccountByID s.pool.QueryRow error:", err)
		return nil, http.StatusInternalServerError, ErrInternal
	}

	return acc, http.StatusOK, nil
}

func (s *Service) Identify(ctx context.Context, id int64) (int, error) {
	_, err := s.pool.Exec(ctx, `UPDATE accounts SET identified = true WHERE id = $1`, id)
	if err != nil {
		log.Println("Identify s.pool.Exec error:", err)
		return http.StatusInternalServerError, ErrInternal
	}

	return http.StatusOK, nil
}
