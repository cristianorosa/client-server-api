package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "database.db"
const endpointCambio = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
const timeoutHttp = time.Millisecond * 200
const timeoutDb = time.Millisecond * 10

// Estrutura para representar o JSON retornado pela API
type Cotacao struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

// Função para conectar ao banco de dados SQLite
func conectarBanco() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Função para criar a tabela se ela não existir
func criarTabela(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		valor TEXT,
		data TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}
}

// Função para salvar a cotação no banco de dados
func salvarCotacao(db *sql.DB, valor string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDb)
	defer cancel()

	// Usar o contexto para garantir o timeout
	query := `INSERT INTO cotacoes (valor) VALUES (?)`
	_, err := db.ExecContext(ctx, query, valor)
	return err
}

// Função para obter a cotação do dólar via API externa
func obterCotacao() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutHttp)
	defer cancel()

	// Requisição HTTP para a API de câmbio
	req, err := http.NewRequestWithContext(ctx, "GET", endpointCambio, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decodificar a resposta JSON
	var cotacao Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		return "", err
	}

	return cotacao.Usdbrl.Bid, nil
}

// Função para lidar com a requisição do cliente
func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	// Conectar ao banco
	db, err := conectarBanco()
	if err != nil {
		http.Error(w, "Erro ao conectar ao banco de dados", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Obter a cotação
	cotacao, err := obterCotacao()
	if err != nil {
		log.Printf("Erro ao acesar o endpoint: %v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Salvar a cotação no banco
	err = salvarCotacao(db, cotacao)
	if err != nil {
		log.Printf("Erro ao salvar na base de dados: %v", err.Error())
		http.Error(w, "Erro ao salvar na base de dados:"+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retornar a cotação para o cliente
	response := map[string]string{"bid": cotacao}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Conectar ao banco
	db, err := conectarBanco()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	defer db.Close()

	// Criar a tabela se não existir
	criarTabela(db)

	// Configurar o servidor HTTP
	http.HandleFunc("/cotacao", cotacaoHandler)
	log.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
