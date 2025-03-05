package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/xKARASb/Calculator/pkg/dir"
)

var client http.Client

type Expression struct {
	Data   string `json:"data"`
	ID     uint32
	Status string
	Result float64
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, dir.Get_template_file("index.html"))
}

func calculate(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, dir.Get_template_file("calc.html"))
}

func showID(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Expression string `json:"expression"`
	}{r.FormValue("expression")}
	code, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	resp, err := client.Post("http://localhost:8080/api/v1/calculate", "application/json", bytes.NewReader(code))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var m map[string]uint32
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.ParseFiles(dir.Get_template_file("showid.html"))
	if err != nil {
		panic(err)
	}
	tmpl.Execute(w, fmt.Sprintf("ID=%v", m["id"]))
}

type APIGetExpressionsResult struct {
	Expressions []Expression `json:"expressions"`
}

func expressions(w http.ResponseWriter, r *http.Request) {
	resp, err := client.Get("http://localhost:8080/api/v1/expressions")
	if err != nil {
		panic(err)
	}
	var res APIGetExpressionsResult
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.ParseFiles(dir.Get_template_file("expressions.html"))
	if err != nil {
		panic(err)
	}
	tmpl.Execute(w, res.Expressions)
}
func HandleToRouter(router *mux.Router) {
	router.HandleFunc("/api/v1/web", index)
	router.HandleFunc("/api/v1/web/calculate", calculate)
	router.HandleFunc("/api/v1/web/expressions", expressions)
	router.HandleFunc("/api/v1/web/showid", showID)
}
