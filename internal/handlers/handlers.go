package handlers

import (
	"encoding/json"
	"food-order-api/internal/models"
	"food-order-api/internal/queue"
	"food-order-api/internal/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo  *repository.Repository
	queue *queue.QueueManager
}

func NewHandler(repo *repository.Repository, qm *queue.QueueManager) *Handler {
	return &Handler{repo: repo, queue: qm}
}

func (h *Handler) GetMenu(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.GetMenu(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, items)
}

func (h *Handler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	item, err := h.repo.GetMenuItem(r.Context(), id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	respondWithJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateItemAvailability(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var payload models.UpdateAvailabilityPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if payload.Availability != "in_stock" && payload.Availability != "out_of_stock" {
		http.Error(w, "Availability must be 'in_stock' or 'out_of_stock'", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateItemAvailability(r.Context(), id, payload.Availability); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var payload models.PlaceOrderPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if len(payload.Items) == 0 {
		http.Error(w, "Order items cannot be empty", http.StatusBadRequest)
		return
	}

	order, err := h.repo.CreateOrder(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Trigger Async Pipeline
	_ = h.queue.PublishOrderEvent(r.Context(), models.OrderPlacedEvent{
		OrderID:    order.ID,
		TotalCents: order.TotalCents,
		Timestamp:  time.Now(),
	})

	respondWithJSON(w, http.StatusCreated, order)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	order, err := h.repo.GetOrder(r.Context(), id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	respondWithJSON(w, http.StatusOK, order)
}

func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var payload models.UpdateStatusPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{"received": true, "preparing": true, "ready": true, "completed": true}
	if !validStatuses[payload.Status] {
		http.Error(w, "Invalid target status type", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateOrderStatus(r.Context(), id, payload.Status); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}