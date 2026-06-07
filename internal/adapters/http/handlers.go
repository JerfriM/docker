package http

import (
	"context"
	"encoding/json"
	"io"
	"mi-api-go/internal/core/ports"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	svc ports.UserService
}

func NewUserHandler(svc ports.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type registerReq struct {
	Name      string `json:"name"`
	Matricula string `json:"matricula"`
	Password  string `json:"password"`
}

type updateReq struct {
	Name      string `json:"name"`
	Matricula string `json:"matricula"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	user, err := h.svc.Register(r.Context(), req.Name, req.Matricula, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct{ Matricula, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	token, err := h.svc.Login(r.Context(), req.Matricula, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "No encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateUser(r.Context(), id, req.Name, req.Matricula); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario actualizado"})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario eliminado"})
}

func (h *UserHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	bucketName := os.Getenv("S3_BUCKET")

	if bucketName == "" {
		r.ParseMultipartForm(5 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error al obtener el archivo", http.StatusBadRequest)
			return
		}
		defer file.Close()

		uploadDir := "./uploads"
		os.MkdirAll(uploadDir, os.ModePerm)
		dst, err := os.Create(filepath.Join(uploadDir, handler.Filename))
		if err != nil {
			http.Error(w, "Error al guardar archivo", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":  "Archivo subido exitosamente",
			"filename": handler.Filename,
			"url":      "/uploads/" + handler.Filename,
		})
		return
	}

	r.ParseMultipartForm(5 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error al obtener el archivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		http.Error(w, "Error al configurar AWS", http.StatusInternalServerError)
		return
	}

	client := s3.NewFromConfig(cfg)
	key := "uploads/" + handler.Filename
	contentType := handler.Header.Get("Content-Type")

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &key,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		http.Error(w, "Error al subir a S3", http.StatusInternalServerError)
		return
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	url := "https://" + bucketName + ".s3." + region + ".amazonaws.com/" + key

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Archivo subido a S3",
		"filename": handler.Filename,
		"url":      url,
	})
}
