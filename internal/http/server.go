package http

import (
	"WBTechL0/internal/config"
	"WBTechL0/internal/service"
	"fmt"
	"html/template"
	"net/http"
)

// Server Структура Сервера
type Server struct {
	svc *service.OrderService
	cfg config.HttpServer
}

// New - Конструктор для создания нового httpServer
func New(svc *service.OrderService, cfgHttp config.HttpServer) *Server {
	return &Server{svc: svc, cfg: cfgHttp}
}

// Start - Метод для запуска HTTP сервера
func (s *Server) Start() {
	m := http.NewServeMux()

	m.HandleFunc("GET /id", handleMain)
	m.HandleFunc("GET /id/{uid}", handleGetOrder(s.svc))
	m.HandleFunc("POST /id", handlePostOrder)
	port := fmt.Sprintf(":%v", s.cfg.Port)
	if err := http.ListenAndServe(port, m); err != nil {
		s.svc.Sl.Error("Could not start server: %v", err)
	}
	// Запуск сервера
	s.svc.Sl.Info("Starting HTTP server")
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "templates/index.html") // Отправляем HTML
}

func handleGetOrder(svc *service.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		orderUID := r.PathValue("uid")
		order := svc.GetOrder(orderUID)

		if order == nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		// Парсинг шаблона
		tmpl, err := template.ParseFiles("templates/order.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Выполнение шаблона с данными
		err = tmpl.Execute(w, order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handlePostOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		orderUID := r.FormValue("order_uid") // Предполагается, что у вас есть поле с именем order_uid в форме
		url := fmt.Sprintf("/id/%v", orderUID)
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}
}
