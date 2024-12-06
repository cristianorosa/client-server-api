package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const serverURL = "http://localhost:8080/cotacao"
const arquivoCotacao = "cotacao.txt"
const timeout = 300 * time.Millisecond

// Função para realizar a requisição HTTP e obter a cotação
func obterCotacaoDoServidor() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Requisição HTTP para o servidor
	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Ler a resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extrair a cotação do corpo da resposta
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result["bid"], nil
}

// Função para salvar a cotação em um arquivo
func salvarCotacaoNoArquivo(cotacao string) error {
	conteudo := fmt.Sprintf("Dólar: %s", cotacao)
	return os.WriteFile(arquivoCotacao, []byte(conteudo), 0644)
}

func main() {
	// Obter a cotação do servidor
	cotacao, err := obterCotacaoDoServidor()
	if err != nil {
		log.Fatalf("Erro ao obter cotação do servidor: %v", err)
		return
	}

	// Salvar a cotação no arquivo
	err = salvarCotacaoNoArquivo(cotacao)
	if err != nil {
		log.Fatalf("Erro ao salvar cotação no arquivo: %v", err)
		return
	}

	log.Printf("Cotação salva em %s: Dólar: %s", arquivoCotacao, cotacao)
}
