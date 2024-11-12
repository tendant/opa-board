package hive

import (
	"context"
	"encoding/json"
	"net/http"
	"text/template"

	"log/slog"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
)

func RenderIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("../../templates/evaluate.html")
	if err != nil {
		slog.Error("Failed rendering template", "err", err)
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func EvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Policy string `json:"policy"`
		Data   string `json:"data"`
		Input  string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse `data` and `input` into objects
	var data, input map[string]interface{}
	dataStr := request.Data
	if dataStr == "" {
		dataStr = "{}"
	}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		http.Error(w, "Invalid JSON in 'data': "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal([]byte(request.Input), &input); err != nil {
		http.Error(w, "Invalid JSON in 'input': "+err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("inputs", "policy", request.Policy, "data", data, "input", input)

	// Create an in-memory store for the data
	store := inmem.NewFromObject(data)

	// Create OPA query
	ctx := context.Background()
	query, err := rego.New(
		rego.Query("data.example.allow"),
		rego.Module("policy.rego", request.Policy),
		rego.Input(input),
		rego.Store(store),
	).PrepareForEval(ctx)
	if err != nil {
		slog.Error("Failed evaluating opa policy", "err", err)
		http.Error(w, "Unable to prepare policy for evaluation", http.StatusInternalServerError)
		return
	}

	// Evaluate policy
	results, err := query.Eval(ctx)
	if err != nil || len(results) == 0 {
		http.Error(w, "Policy evaluation failed", http.StatusInternalServerError)
		return
	}

	// Render result dynamically
	tmpl, err := template.ParseFiles("../../templates/result.html")
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
		return
	}

	slog.Info("result", "results", results)

	response := map[string]interface{}{
		"result": results[0].Expressions[0].Value,
	}
	tmpl.Execute(w, response)
}
