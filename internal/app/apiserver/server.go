package apiserver

import (
	"canteen-go/internal/app/store"
	"canteen-go/model"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	sessionName        = "go"
	ctxKeyUser  ctxKey = iota
	ctxKeyRequestID
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated         = errors.New("not authenticated")
)

type ctxKey int8

type server struct {
	router       *mux.Router
	logger       *logrus.Logger
	store        store.Store
	sessionStore sessions.Store
}

func newServer(store store.Store, sessionStore sessions.Store) *server {
	s := &server{
		router:       mux.NewRouter(),
		logger:       logrus.New(),
		store:        store,
		sessionStore: sessionStore,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.HandleFunc("/users", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/sessions", s.handleSessionsCreate()).Methods("POST")

	admin := s.router.PathPrefix("/admin").Subrouter()
	admin.Use(s.authenticateUser)
	admin.Use(s.checkAdmin)
	admin.HandleFunc("/menuItem", s.handleMenuItemCreate()).Methods("POST")

	private := s.router.PathPrefix("/private").Subrouter()
	private.Use(s.authenticateUser)
	private.HandleFunc("/whoami", s.handleWhoami()).Methods("GET")
	private.HandleFunc("/orders", s.handleCreateOrder()).Methods("POST")
}

func (s *server) handleCreateOrder() http.HandlerFunc {
	type respondOrder struct {
		Id         int                `json:"id"`
		OrderItems []*model.OrderItem `json:"order_item"`
		CreatedAt  time.Time          `json:"created_At"`
		TotalPrice int                `json:"total_price"`
	}

	type requests struct {
		MenuItemId []int `json:"menu_item_id"`
		Quantity   []int `json:"quantity"`
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		userId, _ := s.getUserId(writer, request)

		req := &requests{}
		if err := json.NewDecoder(request.Body).Decode(req); err != nil {
			s.error(writer, request, http.StatusBadRequest, err)
			return
		}

		var totalAmount int
		for i := 0; i < len(req.MenuItemId); i++ {
			price := s.store.MenuItem().GetPrice(req.MenuItemId[i])
			s.logger.Info(price, req.MenuItemId[i])
			totalAmount += price * req.Quantity[i]
		}

		o := &model.Order{
			UserId:      userId,
			TotalAmount: totalAmount,
		}

		exception := s.store.Order().Create(o)
		if exception != nil {
			s.error(writer, request, http.StatusUnprocessableEntity, exception)
			return
		}

		var orderItems []*model.OrderItem

		for i := 0; i < len(req.MenuItemId); i++ {
			mi := &model.OrderItem{
				OrderId:    o.ID,
				MenuItemId: req.MenuItemId[i],
				Quantity:   req.Quantity[i],
			}

			if err := s.store.OrderItem().Create(mi); err != nil {
				if e := s.store.Order().Delete(o.ID); e != nil {
					s.error(writer, request, http.StatusInternalServerError, e)
					return
				}
				s.error(writer, request, http.StatusUnprocessableEntity, err)
				return
			}

			orderItems = append(orderItems, mi)
		}

		respondOrder := respondOrder{
			Id:         o.ID,
			CreatedAt:  o.CreatedAt,
			TotalPrice: totalAmount,
			OrderItems: orderItems,
		}

		s.respond(writer, request, http.StatusCreated, respondOrder)
	}
}

func (s *server) checkAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := r.Context().Value(ctxKeyUser).(*model.User)
		if !ok {
			s.error(w, r, http.StatusUnauthorized, errors.New("unauthorized access: missing user information"))
			return
		}
		s.logger.Info(u)
		if u.Role != "admin" {
			s.error(w, r, http.StatusForbidden, errors.New("insufficient privileges: requires admin role"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Infof(
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start))
	})
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(id.(int))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *server) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
}

func (s *server) handleMenuItemCreate() http.HandlerFunc {
	type request struct {
		Name        string `json:"name"`
		Price       int    `json:"price"`
		Description string `json:"description"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		// Валидация данных
		if req.Price < 0 {
			s.error(w, r, http.StatusBadRequest, errors.New("price must be positive"))
			return
		}
		menuItem := &model.MenuItem{
			Name:        req.Name,
			Price:       req.Price,
			Description: req.Description,
		}

		// Сохранение продукта в базе данных
		if err := s.store.MenuItem().Create(menuItem); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, menuItem)
	}
}

func (s *server) handleUsersCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email:    req.Email,
			Password: req.Password,
			Role:     "user",
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *server) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session.Values["user_id"] = u.ID
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) error(writer http.ResponseWriter, request *http.Request, code int, err error) {
	s.respond(writer, request, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *server) getUserId(writer http.ResponseWriter, request *http.Request) (int, error) {
	session, err := s.sessionStore.Get(request, sessionName)

	if err != nil {
		s.error(writer, request, http.StatusInternalServerError, err)
		return 0, err
	}

	userId, authorized := session.Values["user_id"]

	if !authorized {
		s.error(writer, request, http.StatusUnauthorized, errNotAuthenticated)
		return 0, nil
	}

	return userId.(int), nil
}
