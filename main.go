package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
	DatabasePassword string `json:"databasePassword"`
}

type DadosBomba struct {
	Id         int
	PumpNumber int
	UnitValue  float64
	Quantity   float64
	TotalValue float64
	Date       time.Time
}

type ApiConfig struct {
	Url string
}

func main() {
	var (
		organizationCodeTE, gasStationCodeTE, databaseIPTE, databasePortTE, databaseUserTE, databasePasswordTE *walk.TextEdit
	)

	existingData, err := readDataFromRegistry()
	if err != nil {
		log.Println(err)
	}

	mw := new(MyWindow)
	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "Touch Sistemas - Postos",
		Size:     Size{Width: 400, Height: 200},
		Layout:   VBox{},
		Icon:     "touchsistemas.ico",
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
						log.Println(err)
					}
					mw.Close()
				},
			},
		},
	}.Create()); err != nil {
		log.Println(err)
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
		log.Println(err)
	}
	icon, err := walk.Resources.Image("touchsistemas.ico")
	if err != nil {
		log.Println(err)
	}
	mw.ni.SetIcon(icon)
	if err := mw.ni.SetToolTip("Touch Sistemas Postos - Integração com Desbravador"); err != nil {
		log.Println(err)
	}
	exitAction := walk.NewAction()
	if err := exitAction.SetText("Fechar"); err != nil {
		log.Println(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := mw.ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Println(err)
	}

	mw.ni.SetVisible(true)
	if err := mw.ni.ShowMessage("Touch Sistemas Postos", "Integração com Desbravador."); err != nil {
		log.Println(err)
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

func sendData(existingData Data, apiConfig ApiConfig, dadosBomba DadosBomba) {
	dados := map[string]string{
		"organizationCode": existingData.OrganizationCode,
		"gasStationCode":   existingData.GasStationCode,
		"pumpNumber":       strconv.Itoa(dadosBomba.PumpNumber),
		"quantity":         strconv.FormatFloat(dadosBomba.Quantity, 'f', -1, 64),
		"unitValue":        strconv.FormatFloat(dadosBomba.UnitValue, 'f', -1, 64),
		"totalValue":       strconv.FormatFloat(dadosBomba.TotalValue, 'f', -1, 64),
		"date":             dadosBomba.Date.Format("2006-01-02 15:04:05"),
	}
	jsonData, err := json.Marshal(dados)
	if err != nil {
		log.Println(err)
	}
	req, err := http.NewRequest("POST", apiConfig.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			log.Println(resp.StatusCode)
		} else {
			log.Println(dados)
		}
	}
}

func readDatabase(existingData Data, apiConfig ApiConfig) {
	pgCconnStr := fmt.Sprintf("user=%s dbname=desbravador password=%s host=%s port=%s sslmode=disable", existingData.DatabaseUser, existingData.DatabasePassword, existingData.DatabaseIP, existingData.DatabasePort)
	pgDB, err := sql.Open("postgres", pgCconnStr)
	if err != nil {
		log.Println(err)
	}
	defer pgDB.Close()
	rows, err := pgDB.Query("SELECT * FROM bombas")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		dadosBomba := DadosBomba{}
		err := rows.Scan(&dadosBomba.Id, &dadosBomba.PumpNumber, &dadosBomba.UnitValue, &dadosBomba.Quantity, &dadosBomba.TotalValue, &dadosBomba.Date)
		if err != nil {
			log.Println(err)
		}
		sendData(existingData, apiConfig, dadosBomba)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
}

func sendDataPeriodically() {
	existingData, err := readDataFromRegistry()
	if err != nil {
		log.Println(err)
	}
	err = godotenv.Load()
	if err != nil {
		log.Println("Erro ao carregar o arquivo .env:", err)
	}
	apiConfig := ApiConfig{
		Url: os.Getenv("API_URL"),
	}
	readDatabase(existingData, apiConfig)
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		readDatabase(existingData, apiConfig)
	}
}
