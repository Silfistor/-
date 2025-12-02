package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Car struct {
	ID           int    `json:"id"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	PriceThousand int    `json:"price_thousands"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("postgres", "host=localhost user=autosalon_user password=secure_password dbname=autosalon sslmode=disable")
	if err != nil {
		log.Fatal("❌ Не удалось открыть подключение к БД:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("❌ Не удалось подключиться к PostgreSQL:", err)
	}
	log.Println("✅ Успешное подключение к PostgreSQL")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Автосалон — Поиск</title>
  <style>
    body { font-family: Arial, sans-serif; padding: 20px; background: #f9f9f9; }
    h1 { color: #2c3e50; }
    form { background: white; padding: 20px; border-radius: 8px; max-width: 500px; }
    label { display: block; margin: 10px 0 5px; font-weight: bold; }
    input, select, button {
      width: 100%;
      padding: 8px;
      margin-bottom: 10px;
      border: 1px solid #ccc;
      border-radius: 4px;
      box-sizing: border-box;
    }
    button {
      background: #3498db;
      color: white;
      cursor: pointer;
    }
    button:hover { background: #2980b9; }
  </style>
</head>
<body>
  <h1>Поиск автомобилей в автосалоне</h1>
  <form action="/search" method="GET">
    <label for="field">Поле поиска:</label>
    <select name="field" id="field" required>
      <option value="brand">Марка</option>
      <option value="model">Модель</option>
      <option value="year">Год выпуска</option>
      <option value="price">Цена (тыс. руб)</option>
    </select>

    <label for="q">Значение:</label>
    <input type="text" name="q" id="q" placeholder="Например: Toyota, 2022, 2500" required>

    <button type="submit">Найти</button>
  </form>
</body>
</html>`
	t := template.Must(template.New("home").Parse(tmpl))
	t.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	field := r.URL.Query().Get("field")

	if query == "" || field == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var rows *sql.Rows
	var err error

	switch field {
	case "brand":
		rows, err = db.Query(`
			SELECT c.id, b.name, m.name, c.year, c.price_thousands
			FROM cars c
			JOIN models m ON c.model_id = m.id
			JOIN brands b ON m.brand_id = b.id
			WHERE b.name ILIKE $1`, "%"+query+"%")
	case "model":
		rows, err = db.Query(`
			SELECT c.id, b.name, m.name, c.year, c.price_thousands
			FROM cars c
			JOIN models m ON c.model_id = m.id
			JOIN brands b ON m.brand_id = b.id
			WHERE m.name ILIKE $1`, "%"+query+"%")
	case "year":
		year, convErr := strconv.Atoi(query)
		if convErr != nil {
			http.Error(w, "Год должен быть числом", http.StatusBadRequest)
			return
		}
		rows, err = db.Query(`
			SELECT c.id, b.name, m.name, c.year, c.price_thousands
			FROM cars c
			JOIN models m ON c.model_id = m.id
			JOIN brands b ON m.brand_id = b.id
			WHERE c.year = $1`, year)
	case "price":
		price, convErr := strconv.Atoi(query)
		if convErr != nil {
			http.Error(w, "Цена должна быть числом", http.StatusBadRequest)
			return
		}
		rows, err = db.Query(`
			SELECT c.id, b.name, m.name, c.year, c.price_thousands
			FROM cars c
			JOIN models m ON c.model_id = m.id
			JOIN brands b ON m.brand_id = b.id
			WHERE c.price_thousands = $1`, price)
	default:
		http.Error(w, "Недопустимое поле поиска", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("Ошибка запроса к БД: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cars []Car
	for rows.Next() {
		var c Car
		if err := rows.Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.PriceThousand); err != nil {
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		cars = append(cars, c)
	}

	// JSON-режим
	if r.URL.Query().Get("json") == "1" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cars)
		return
	}

	// HTML-режим
	fieldLabels := map[string]string{
		"brand": "Марка",
		"model": "Модель",
		"year":  "Год выпуска",
		"price": "Цена (тыс. руб)",
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Результаты поиска — Автосалон</title>
  <style>
    body { font-family: Arial, sans-serif; padding: 20px; background: #f9f9f9; }
    h2 { color: #2c3e50; }
    table { width: 100%; border-collapse: collapse; margin: 20px 0; background: white; }
    th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
    th { background: #ecf0f1; }
    a { display: inline-block; margin-top: 20px; color: #3498db; text-decoration: none; }
    a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <h2>Результаты поиска: "{{.Query}}" в поле "{{.FieldLabel}}"</h2>
  {{if .Results}}
    <table>
      <tr><th>Марка</th><th>Модель</th><th>Год</th><th>Цена (тыс. руб)</th></tr>
      {{range .Results}}
        <tr><td>{{.Brand}}</td><td>{{.Model}}</td><td>{{.Year}}</td><td>{{.PriceThousand}}</td></tr>
      {{end}}
    </table>
  {{else}}
    <p>Ничего не найдено.</p>
  {{end}}
  <a href="/">← Новый поиск</a>
</body>
</html>`

	t := template.Must(template.New("results").Parse(tmpl))
	t.Execute(w, struct {
		Query      string
		FieldLabel string
		Results    []Car
	}{
		Query:      query,
		FieldLabel: fieldLabels[field],
		Results:    cars,
	})
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler)

	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
