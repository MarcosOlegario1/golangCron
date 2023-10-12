package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	_ "github.com/go-pg/pg/v10"
	_ "github.com/go-pg/pg/v10/orm"
	_ "github.com/lib/pq"
)

func hello(nome string) {
	message := fmt.Sprintf("opa, %v", nome)
	fmt.Println(message)
}

func runCronJobs() {
	s := gocron.NewScheduler(time.UTC)

	s.Every(2).Seconds().Do(func() {

		arquivo, err := os.ReadFile("databases.json")
		if err != nil {
			fmt.Println("Erro ao ler o arquivo:", err)
			return
		}

		var configs map[string]DatabaseConfig
		err = json.Unmarshal(arquivo, &configs)
		if err != nil {
			fmt.Println("Erro ao decodificar o JSON:", err)
			return
		}

		// Loop sobre as configurações de banco de dados no mapa
		for nome, config := range configs {
			// Criar string de conexão com o banco de dados
			connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
				config.User, config.Password, config.Host, config.Port, config.DBName)

			// Abrir a conexão com o banco de dados
			db, err := sql.Open("postgres", connectionString)
			if err != nil {
				fmt.Printf("Erro ao abrir a conexão com %s: %v\n", nome, err)
				continue
			}

			// Tentar fazer uma consulta no banco de dados
			err = db.Ping()
			if err != nil {
				fmt.Printf("Erro ao conectar ao %s: %v\n", nome, err)
			} else {
				fmt.Printf("Conexão bem-sucedida ao %s\n", nome)
			}

			// Não se esqueça de fechar a conexão após usar
			defer db.Close()
		}

	})

	s.StartBlocking()
}

func main() {
	runCronJobs()
}

type DatabaseConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}
