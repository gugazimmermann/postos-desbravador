package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"
)

type MyWindow struct {
	*walk.MainWindow
	ni *walk.NotifyIcon
}

type Data struct {
	OrganizationCode string `json:"organizationCode"`
	GasStationCode   string `json:"gasStationCode"`
	DatabaseIP       string `json:"databaseIP"`
	DatabasePort     string `json:"databasePort"`
	DatabaseUser     string `json:"databaseUser"`
	DatabasePassword string `json:"atabasePassword"`
}

type DadosBomba struct {
	ID            int
	Bomba         int
	ValorUnitario float64
	Litros        float64
	ValorTotal    float64
}

type MySQLConfig struct {
	DatabaseIP       string `json:"databaseIP"`
	DatabasePort     string `json:"databasePort"`
	DatabaseUser     string `json:"databaseUser"`
	DatabasePassword string `json:"atabasePassword"`
}

func main() {
	var (
		organizationCodeTE, gasStationCodeTE, databaseIPTE, databasePortTE, databaseUserTE, databasePasswordTE *walk.TextEdit
	)

	existingData, err := readDataFromRegistry()
	if err != nil {
		log.Fatal(err)
	}

	mw := new(MyWindow)
	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "Touch Sistemas - Postos",
		Size:     Size{Width: 400, Height: 200},
		Layout:   VBox{},
		Icon:     "icon.ico",
		OnSizeChanged: func() {
			if win.IsIconic(mw.Handle()) {
				mw.Hide()
			}
		},
		Children: []Widget{
			GroupBox{
				Title:  "Preencha os dados abaixo:",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "IP do Banco de Dados:"},
					TextEdit{AssignTo: &databaseIPTE, Text: existingData.DatabaseIP},
					Label{Text: "Porta do Banco de Dados:"},
					TextEdit{AssignTo: &databasePortTE, Text: existingData.DatabasePort},
					Label{Text: "Usuário do Banco de Dados:"},
					TextEdit{AssignTo: &databaseUserTE, Text: existingData.DatabaseUser},
					Label{Text: "Senha do Banco de Dados:"},
					TextEdit{AssignTo: &databasePasswordTE, Text: existingData.DatabasePassword},
					Label{Text: "Código da Organização:"},
					TextEdit{AssignTo: &organizationCodeTE, Text: existingData.OrganizationCode},
					Label{Text: "Código do Posto:"},
					TextEdit{AssignTo: &gasStationCodeTE, Text: existingData.GasStationCode},
				},
			},
			PushButton{
				Text: "Continuar",
				OnClicked: func() {
					databaseIP := strings.TrimSpace(databaseIPTE.Text())
					databasePort := strings.TrimSpace(databasePortTE.Text())
					databaseUser := strings.TrimSpace(databaseUserTE.Text())
					databasePassword := strings.TrimSpace(databasePasswordTE.Text())
					organizationCode := strings.TrimSpace(organizationCodeTE.Text())
					gasStationCode := strings.TrimSpace(gasStationCodeTE.Text())
					existingData.DatabaseIP = databaseIP
					existingData.DatabasePort = databasePort
					existingData.DatabaseUser = databaseUser
					existingData.DatabasePassword = databasePassword
					existingData.OrganizationCode = organizationCode
					existingData.GasStationCode = gasStationCode
					err := saveDataToRegistry(existingData)
					if err != nil {
						log.Fatal(err)
					}
					mw.Close()
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}
	mw.AddNotifyIcon()
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		*canceled = true
		mw.Hide()
		go sendDataPeriodically()
	})
	mw.Run()
}

func (mw *MyWindow) AddNotifyIcon() {
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}
	icon, err := walk.Resources.Image("icon.ico")
	if err != nil {
		log.Fatal(err)
	}
	mw.ni.SetIcon(icon)
	if err := mw.ni.SetToolTip("Touch Sistemas Postos - Integração com Desbravador"); err != nil {
		log.Fatal(err)
	}
	exitAction := walk.NewAction()
	if err := exitAction.SetText("Fechar"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := mw.ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}

	mw.ni.SetVisible(true)
	if err := mw.ni.ShowMessage("Touch Sistemas Postos", "Integração com Desbravador."); err != nil {
		log.Fatal(err)
	}
}

func readDataFromRegistry() (Data, error) {
	var data Data
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\TouchSistemas\Postos`, registry.QUERY_VALUE)
	if err != nil {
		if err != registry.ErrNotExist {
			return data, err
		}
		return data, nil
	}
	defer k.Close()
	databaseIP, _, err := k.GetStringValue("DatabaseIP")
	if err != nil {
		return data, err
	}
	databasePort, _, err := k.GetStringValue("DatabasePort")
	if err != nil {
		return data, err
	}
	databaseUser, _, err := k.GetStringValue("DatabaseUser")
	if err != nil {
		return data, err
	}
	databasePassword, _, err := k.GetStringValue("DatabasePassword")
	if err != nil {
		return data, err
	}
	organizationCode, _, err := k.GetStringValue("OrganizationCode")
	if err != nil {
		return data, err
	}
	gasStationCode, _, err := k.GetStringValue("GasStationCode")
	if err != nil {
		return data, err
	}
	data.DatabaseIP = databaseIP
	data.DatabasePort = databasePort
	data.DatabaseUser = databaseUser
	data.DatabasePassword = databasePassword
	data.OrganizationCode = organizationCode
	data.GasStationCode = gasStationCode
	return data, nil
}

func saveDataToRegistry(data Data) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\TouchSistemas\Postos`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	err = k.SetStringValue("DatabaseIP", data.DatabaseIP)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DatabasePort", data.DatabasePort)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DatabaseUser", data.DatabaseUser)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DatabasePassword", data.DatabasePassword)
	if err != nil {
		return err
	}
	err = k.SetStringValue("OrganizationCode", data.OrganizationCode)
	if err != nil {
		return err
	}
	err = k.SetStringValue("GasStationCode", data.GasStationCode)
	if err != nil {
		return err
	}
	return nil
}

func readDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env:", err)
	}
	MySQLConfig.dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	mysqlConnStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	mysqlDB, err := sql.Open("mysql", mysqlConnStr)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco de dados MySQL:", err)
	}
	defer mysqlDB.Close()

	existingData, err := readDataFromRegistry()
	if err != nil {
		log.Fatal(err)
	}
	connStr := fmt.Sprintf("user=%s dbname=desbravador password=%s host=%s port=%s sslmode=disable", existingData.DatabaseUser, existingData.DatabasePassword, existingData.DatabaseIP, existingData.DatabasePort)
	pgDB, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer pgDB.Close()
	rows, err := pgDB.Query("SELECT id, bomba, valor_unitario, litros, valor_total FROM bombas")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		dadosBomba := DadosBomba{}
		err := rows.Scan(&dadosBomba.ID, &dadosBomba.Bomba, &dadosBomba.ValorUnitario, &dadosBomba.Litros, &dadosBomba.ValorTotal)
		if err != nil {
			log.Fatal(err)
		}
		_, err = mysqlDB.Exec("INSERT INTO tabela (bomba, valor_unitario, litros, valor_total) VALUES (?, ?, ?, ?, ?)", dadosBomba.Bomba, dadosBomba.ValorUnitario, dadosBomba.Litros, dadosBomba.ValorTotal)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("ID: %d, Bomba: %d, Valor Unitário: %.2f, Litros: %.2f, Valor Total: %.2f", dadosBomba.ID, dadosBomba.Bomba, dadosBomba.ValorUnitario, dadosBomba.Litros, dadosBomba.ValorTotal)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func sendDataPeriodically() {
	readDatabase()
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-ticker.C:
			readDatabase()
		}
	}
}
