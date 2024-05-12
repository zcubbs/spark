package web

import (
	"github.com/charmbracelet/log"
	k8sJobs "github.com/zcubbs/spark/pkg/k8s/jobs"
	"html"
	"html/template"
	"net/http"
	"strings"
)

const (
	internalServerError = "Internal Error 500"
)

type Handler struct {
	k8sRunner *k8sJobs.Runner
	templates *template.Template
}

func NewHandler(k8sRunner *k8sJobs.Runner) (*Handler, error) {
	funcMap := template.FuncMap{
		"join": join, // Add the join function to the template's function map
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(FsTemplates, "templates/*.html")
	if err != nil {
		log.Error("failed to load templates", "error", err)
		return nil, err
	}

	return &Handler{
		k8sRunner: k8sRunner,
		templates: tmpl,
	}, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/static/", h.handleGetStaticFiles)
	mux.HandleFunc("/command/", h.handleGetCommand)
	mux.HandleFunc("/logs/", h.handleGetLogs)
	mux.HandleFunc("/tasks", h.handleGetTasks)
	mux.HandleFunc("/", h.handleIndex)
}

func (h *Handler) handleIndex(w http.ResponseWriter, _ *http.Request) {
	err := h.templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Error("failed to execute template", "error", err)
		http.Error(w, internalServerError, http.StatusInternalServerError)
	}
}

func (h *Handler) handleGetTasks(w http.ResponseWriter, _ *http.Request) {
	// Fetch tasks directly as structs from the DB
	tasks, err := h.k8sRunner.GetAllTasksFromDB()
	if err != nil {
		log.Error("failed to fetch jobs", "error", err)
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}

	// Since tasks are now directly retrieved as Task structs, there's no need to parse strings
	err = h.templates.ExecuteTemplate(w, "tasks.html", tasks)
	if err != nil {
		log.Error("failed to execute template", "error", err)
		http.Error(w, internalServerError, http.StatusInternalServerError)
	}
}

func (h *Handler) handleGetStaticFiles(w http.ResponseWriter, r *http.Request) {
	log.Debug("serving static file", "path", r.URL.Path)
	staticFileHandler := http.FileServer(http.FS(FsStaticFiles))
	staticFileHandler.ServeHTTP(w, r)
}

func (h *Handler) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/logs/")
	logs, err := h.k8sRunner.GetLogsForTaskFromDB(taskID)
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte("<html><body><pre>" + html.EscapeString(logs) + "</pre></body></html>"))
	if err != nil {
		log.Error("failed to write logs", "error", err)
		http.Error(w, internalServerError, http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleGetCommand(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/command/")
	command, err := h.k8sRunner.GetCommandForTaskFromDB(taskID)
	if err != nil {
		http.Error(w, "Failed to fetch command", http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte("<html><body><pre>" + html.EscapeString(join(command, " ")) + "</pre></body></html>"))
	if err != nil {
		log.Error("failed to write command", "error", err)
		http.Error(w, internalServerError, http.StatusInternalServerError)
		return
	}
}

func join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}
