package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"

	_ "github.com/lib/pq"
)

type App struct {
	Router *http.ServeMux
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {

	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = http.NewServeMux()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) GetProduct(w http.ResponseWriter, r *http.Request) {

	//read dynamic id params
	id, err := strconv.Atoi(path.Base(r.URL.Path))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	p := Product{
		ID: id,
	}

	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, p)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) GetProducts(w http.ResponseWriter, r *http.Request) {

	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}

	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)

}

func (a *App) CreateProduct(w http.ResponseWriter, r *http.Request) {

	var p Product

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) UpdateProduct(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(path.Base(r.URL.Path))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	var p Product

	decoder := json.NewDecoder(r.Body)

	if err = decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload!")
		return
	}

	defer r.Body.Close()

	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) DeleteProduct(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(path.Base(r.URL.Path))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	p := Product{ID: id}

	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.GetProducts)
	a.Router.HandleFunc("/createproduct", a.CreateProduct)
	a.Router.HandleFunc("/getproduct/", a.GetProduct)
	a.Router.HandleFunc("/updateproduct/", a.UpdateProduct)
	a.Router.HandleFunc("/deleteproduct/", a.DeleteProduct)
}
