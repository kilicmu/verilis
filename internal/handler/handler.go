package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler 处理HTTP请求
type Handler struct {
}

// NewHandler 创建新的处理器
func NewHandler() *Handler {
	return &Handler{}
}

// HealthCheck 健康检查端点
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HelloWorld 示例端点
func (h *Handler) HelloWorld(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Hello, World!",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
