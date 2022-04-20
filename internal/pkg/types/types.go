package types

import (
	"log"
	"os"
	"time"
)

// Loggers for INFO and ERROR messages
var (
	InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	ErrorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)
)

// Type RegInfo is structure with required fields for registration
type RegInfo struct {
	Username string
	Phone    string
	Password string
}

type Account struct {
	ID       	int64    	`json:"id"`
	Balance  	int64	  	`json:"balance"`
	Identified	bool      	`json:"indentified"`

	Username 	string   	`json:"username"`
	Phone		string    	`json:"phone"`
	Password 	string    	`json:"password"`

	Active   	bool      	`json:"active"`
	Created  	time.Time	`json:"created"`
}

// Type TokenInfo is structure with required fields for getting token (login = phone)
type TokenInfo struct {
	Login 		string `json:"login"`
	Password 	string `json:"password"`
}

type Token struct {
	Token		string    `json:"token"`
	AccID		int64     `json:"acc_id"`

	Expires    	time.Time `json:"expires"`
	Created    	time.Time `json:"created"`
}

type Transaction struct {
	ID			int64		`json:"id"`
	AccID		int64		`json:"acc_id"`
	Amount		int64		`json:"amount"`

	Created		time.Time	`json:"created"`
}

// Type TransactionInfo is structure with statistics of transactions per current month
type TransactionsPerMonth struct {
	Sum				int64			`json:"sum"`
	Count			int64			`json:"count"`
	Transactions	[]*Transaction	`json:"transactions"`
}
