package hive

import (
	"context"
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
)

func RenderIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func EvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Policy string                 `json:"policy"`
		Data   map[string]interface{} `json:"data"`
		Input  map[string]interface{} `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Create an in-memory store for the data
	store := inmem.NewFromObject(request.Data)

	// Create OPA query
	ctx := context.Background()
	query, err := rego.New(
		rego.Query("data.example.allow"),
		rego.Module("policy.rego", request.Policy),
		rego.Input(request.Input),
		rego.Store(store),
	).PrepareForEval(ctx)
	if err != nil {
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
	tmpl, err := template.ParseFiles("templates/result.html")
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"result": results[0].Expressions[0].Value,
	}
	tmpl.Execute(w, response)
}
