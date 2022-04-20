package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/SYSTEMTerror/GoWallet/internal/app/middleware"
	"github.com/SYSTEMTerror/GoWallet/internal/pkg/types"
	"github.com/SYSTEMTerror/GoWallet/internal/pkg/wallet"
	"github.com/gorilla/mux"
)

type Server struct {
	mux       *mux.Router
	walletSvc *wallet.Service
}

func NewServer(mux *mux.Router, walletSvc *wallet.Service) *Server {
	return &Server{mux: mux, walletSvc: walletSvc}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Init() {
	s.mux.Use(middleware.Logger)
	s.mux.Use(middleware.LoggersFuncs)

	walletAuthenticateMd := middleware.Authenticate(s.walletSvc.IDByToken)

	walletSubrouter := s.mux.PathPrefix("/api/wallet").Subrouter()
	walletSubrouter.Use(walletAuthenticateMd)

	walletSubrouter.HandleFunc("/exist/{phone}", s.handleExist).Methods("GET")
	walletSubrouter.HandleFunc("/register", s.handleRegister).Methods("POST")
	walletSubrouter.HandleFunc("/token", s.handleToken).Methods("POST")
	walletSubrouter.HandleFunc("/transaction", s.handleTransaction).Methods("POST")
	walletSubrouter.HandleFunc("/transactions/{id}", s.handleGetTransactionsPerMonth).Methods("GET")
	walletSubrouter.HandleFunc("/account/{id}", s.handleGetAccount).Methods("GET")
	walletSubrouter.HandleFunc("/balance/{id}", s.handleBalance).Methods("GET")
	walletSubrouter.HandleFunc("/identify/{id}", s.handleIdentify).Methods("POST")
}

func (s *Server) handleExist(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleExist started.")

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
	err = jsoner(w, mes, statusCode)
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

	acc, statusCode, err := s.walletSvc.Register(r.Context(), item)
	if err != nil {
		loggers.ErrorLogger.Println("handleRegister s.walletSvc.Register error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, acc, statusCode)
	if err != nil {
		loggers.ErrorLogger.Println("handleRegisterCustomer jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleRegisterCustomer finished with any error.")
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	loggers, err := middleware.GetLoggers(r.Context())
	if err != nil {
		log.Println("LOGGERS DON'T WORK!!!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	loggers.InfoLogger.Println("handleToken started.")

	var item *types.TokenInfo
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		loggers.ErrorLogger.Println("handleToken json.NewDecoder error:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, statusCode, err := s.walletSvc.Token(r.Context(), item)
	if err != nil {
		loggers.ErrorLogger.Println("handleToken s.walletSvc.Token error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, token, statusCode)
	if err != nil {
		loggers.ErrorLogger.Println("handleToken jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleToken finished with any error.")
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

	transaction, statusCode, err := s.walletSvc.Transaction(r.Context(), item)
	if err != nil {
		loggers.ErrorLogger.Println("handleTransaction s.walletSvc.Transaction error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, transaction, statusCode)
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

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetTransactionsPerMonth strconv.ParseInt error:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	transactions, sum, count, statusCode, err := s.walletSvc.GetTransactionsPerMonth(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetTransactionsPerMonth s.walletSvc.GetTransactionsPerMonth error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, types.TransactionsPerMonth{Sum: sum, Count: count, Transactions: transactions}, statusCode)
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

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetAccount strconv.ParseInt error:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	
	account, statusCode, err := s.walletSvc.GetAccountByID(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleGetAccount s.walletSvc.GetAccountByID error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, account, statusCode)
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
	
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		loggers.ErrorLogger.Println("handleBalance strconv.ParseInt error:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	acc, statusCode, err := s.walletSvc.GetAccountByID(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleBalance s.walletSvc.GetAccountByID error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	err = jsoner(w, acc.Balance, statusCode)
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

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify strconv.ParseInt error:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	statusCode, err := s.walletSvc.Identify(r.Context(), id)
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify s.walletSvc.Identify error:", err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}


	err = jsoner(w, "Account was identified", statusCode)
	if err != nil {
		loggers.ErrorLogger.Println("handleIdentify jsoner error:", err)
		return
	}
	loggers.InfoLogger.Println("handleIdentify finished with any error.")
}

//function jsoner marshal interfaces to json and write to response writer
func jsoner(w http.ResponseWriter, v interface{}, code int) error {
	data, err := json.Marshal(v)
	if err != nil {
		log.Println("jsoner json.Marshal error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		log.Println("jsoner w.Write error:", err)
		return err
	}
	return nil
}
