package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	_ "github.com/go-pg/pg/v10"
	_ "github.com/go-pg/pg/v10/orm"
	_ "github.com/lib/pq"
)

func runCronJobs() {
	s := gocron.NewScheduler(time.UTC)

	s.Every(15).Minutes().Do(func() {

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

		// Loop sobre as configurações das bases de dados
		for nome, config := range configs {
			// Conexão banco de recurso
			clientConnection := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
				config.User, config.Password, config.Host, config.Port, config.DBName)

			db, err := sql.Open("postgres", clientConnection)
			if err != nil {
				fmt.Printf("Erro ao abrir a conexão com %s: %v\n", nome, err)
				continue
			}

			// Ping pra testar a conexão
			err = db.Ping()
			if err != nil {
				fmt.Printf("Erro ao conectar ao %s: %v\n", nome, err)
			}

			rows, err := db.Query("SELECT")
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			//Conexão com o banco de destino dos dados
			connection2 := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
				"postgres", "teste", "localhost", 5432, "postgres")

			// Abrir a conexão com o banco de dados que vai ser importado os dados
			db2, err := sql.Open("postgres", connection2)
			if err != nil {
				fmt.Printf("Erro ao abrir a conexão no banco de destino dos dados com %s: %v\n", nome, err)
				continue
			}

			// Testa conexão banco de destino
			err = db2.Ping()
			if err != nil {
				fmt.Printf("Erro ao conectar ao banco de destino %s: %v\n", nome, err)
			}

			for rows.Next() {
				var columns string

				stmt, err := db2.Prepare("INSERT ")
				if err != nil {
					log.Fatal(err)
				}
				defer stmt.Close()

				_, err = stmt.Exec(columns)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println("Item importados com sucesso.")
			}

			// Verifica se houve algum erro durante o processo de percorrer as linhas
			if err = rows.Err(); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Banco %s finalizado:", config.DBName)

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
