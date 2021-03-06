package app

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/SYSTEMTerror/GoWallet/internal/app/middleware"
	"github.com/SYSTEMTerror/GoWallet/internal/pkg/types"
	"github.com/SYSTEMTerror/GoWallet/internal/pkg/wallet"
	"github.com/gorilla/mux"
)

type Server struct {
	mux       *mux.Router
	walletSvc *wallet.Service
	secretKey string
}

func NewServer(mux *mux.Router, walletSvc *wallet.Service, secretKey string) *Server {
	return &Server{mux: mux, walletSvc: walletSvc, secretKey: secretKey}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Init() {
	s.mux.Use(middleware.Logger)
	s.mux.Use(middleware.LoggersFuncs)

	walletUserIDMd := middleware.UserID()

	walletSubrouter := s.mux.PathPrefix("/api/wallet").Subrouter()
	walletSubrouter.Use(walletUserIDMd)

	walletSubrouter.HandleFunc("/exist/{phone}", s.handleExist).Methods("GET")
	walletSubrouter.HandleFunc("/register", s.handleRegister).Methods("POST")
	walletSubrouter.HandleFunc("/transaction", s.handleTransaction).Methods("POST")
	walletSubrouter.HandleFunc("/transactions", s.handleGetTransactionsPerMonth).Methods("GET")
	walletSubrouter.HandleFunc("/account", s.handleGetAccount).Methods("GET")
	walletSubrouter.HandleFunc("/balance", s.handleBalance).Methods("GET")
	walletSubrouter.HandleFunc("/identify", s.handleIdentify).Methods("POST")
}

func (s *Server) handleExist(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleExist started.")

	if !verify(r, "", s.secretKey) {
		loggers.ErrorLogger.Println("handleExist verify error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	phone := mux.Vars(r)["phone"]
	exist, acc, statusCode, err := s.walletSvc.Exist(r.Context(), phone)
	if err != nil {
		loggers.ErrorLogger.Println("handleExist s.walletSvc.Exist error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}
	var mes interface{}
	if exist {
		mes = acc
	} else {
		mes = "Account not exist"
	}
	err = jsoner(w, mes, statusCode, s.secretKey)
	if err != nil {
		loggers.ErrorLogger.Println("handleExist jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleExist finished with any error.")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleRegister started.")

	var item *types.RegInfo
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		loggers.ErrorLogger.Println("handleRegister json.NewDecoder error:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !verify(r, fmt.Sprintf("%s", item), s.secretKey) {
		loggers.ErrorLogger.Println("handleExist verify error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	acc, statusCode, err := s.walletSvc.Register(r.Context(), item)
	if err != nil {
		loggers.ErrorLogger.Println("handleRegister s.walletSvc.Register error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, acc, statusCode, s.secretKey)
	if err != nil {
		loggers.ErrorLogger.Println("handleRegisterCustomer jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleRegisterCustomer finished with any error.")
}

func (s *Server) handleTransaction(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleTransaction started.")

	var item *types.Transaction
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		loggers.ErrorLogger.Println("handleTransaction json.NewDecoder error:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !verify(r, fmt.Sprintf("%d", item), s.secretKey) {
		loggers.ErrorLogger.Println("handleExist verify error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := middleware.GetUserID(r.Context())
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify middleware.Authentication error:", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if id != item.AccID {
		loggers.ErrorLogger.Println("handleTransaction id != item.AccID")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	transaction, statusCode, err := s.walletSvc.Transaction(r.Context(), item)
	if err != nil {
		loggers.ErrorLogger.Println("handleTransaction s.walletSvc.Transaction error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, transaction, statusCode, s.secretKey)
	if err != nil {
		loggers.ErrorLogger.Println("handleTransaction jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleTransaction finished with any error.")
}

func (s *Server) handleGetTransactionsPerMonth(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleGetTransactionsPerMonth started.")

	if !verify(r, "", s.secretKey) {
		loggers.ErrorLogger.Println("handleExist verify error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := middleware.GetUserID(r.Context())
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify middleware.Authentication error:", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	transactions, sum, count, statusCode, err := s.walletSvc.GetTransactionsPerMonth(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetTransactionsPerMonth s.walletSvc.GetTransactionsPerMonth error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, types.TransactionsPerMonth{Sum: sum, Count: count, Transactions: transactions}, statusCode, s.secretKey)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetTransactionsPerMonth jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleGetTransactionsPerMonth finished with any error.")
}

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleGetAccount started.")

	if !verify(r, "", s.secretKey) {
		loggers.ErrorLogger.Println("handleExist verify error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := middleware.GetUserID(r.Context())
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify middleware.Authentication error:", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	account, statusCode, err := s.walletSvc.GetAccountByID(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetAccount s.walletSvc.GetAccountByID error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, account, statusCode, s.secretKey)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetAccount jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleGetAccount finished with any error.")
}

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleBalance started.")

	if !verify(r, "", s.secretKey) {
		loggers.ErrorLogger.Println("handleExist verify error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := middleware.GetUserID(r.Context())
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify middleware.Authentication error:", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	acc, statusCode, err := s.walletSvc.GetAccountByID(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleBalance s.walletSvc.GetAccountByID error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, acc.Balance, statusCode, s.secretKey)
	if err != nil {
		loggers.ErrorLogger.Println("handleBalance jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleBalance finished with any error.")
}

func (s *Server) handleIdentify(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleIdentify started.")

	if !verify(r, "", s.secretKey) {
		loggers.ErrorLogger.Println("handleExist verify error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := middleware.GetUserID(r.Context())
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify middleware.Authentication error:", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	statusCode, err := s.walletSvc.Identify(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify s.walletSvc.Identify error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, "Account was identified", statusCode, s.secretKey)
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleIdentify finished with any error.")
}

//function jsoner marshal interfaces to json and write to response writer
func jsoner(w http.ResponseWriter, v interface{}, code int, secretKey string) error {
	data, err := json.Marshal(v)
	if err != nil {
		log.Println("jsoner json.Marshal error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("X-Digest", hasher(string(data), secretKey))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		log.Println("jsoner w.Write error:", err)
		return err
	}
	return nil
}

//fuction hasher create hmac-sha1 hash from string and secret key
func hasher(s string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(s))
	return "sha1=" + hex.EncodeToString(h.Sum(nil))
}

//function verify gets hmac-sha1 hash from request header and compare it with request body
func verify(r *http.Request, s string, secret string) bool {
	h := r.Header.Get("X-Digest")
	if h == "" {
		return false
	}
	return h == hasher(s, secret)
}
