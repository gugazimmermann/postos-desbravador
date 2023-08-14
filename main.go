package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/natefinch/lumberjack.v2"
)

type MyWindow struct {
	*walk.MainWindow
	ni *walk.NotifyIcon
}

type Data struct {
	OrganizationCode string `json:"organizationCode"`
	GasStationCode   string `json:"gasStationCode"`
	DBIp             string `json:"dbIP"`
	DBPort           string `json:"dbPort"`
	DBDatabase       string `json:"dbDatabase"`
	DBUser           string `json:"dbUser"`
	DBPwd            string `json:"dbPwd"`
	DBRole           string `json:"dbRole"`
	DBCompanyID      string `json:"dbCompanyID"`
}

type PumpRows struct {
	CdAbastecimento int
	DhAbastecimento time.Time
	QtVolume        float64
	VlUnitario      float64
	VlTotal         float64
	FlLancado       int
	DsBico          string
	NrBico          int
	DsApelido       string
}

type PumpRowsData struct {
	GasStationTransactionID int
	Quantity                float64
	UnitValue               float64
	TotalValue              float64
	Processed               int
	Date                    string
	PumpNumber              int
	FuelName                string
	CompanyName             string
}

type PumpsData struct {
	OrganizationCode string
	GasStationCode   string
	PumpRowsData     []PumpRowsData
}

type ApiConfig struct {
	Url string
}

func showErrorMessageBox(message string) {
	walk.MsgBox(nil, "Error", message, walk.MsgBoxIconError|walk.MsgBoxOK)
}

func fatalError(message string, err error) {
	if err != nil {
		log.Printf("Erro: %s: %v", message, err)
		showErrorMessageBox(fmt.Sprintf("Erro: %s: %v", message, err))
	} else {
		log.Printf("Erro: %s", message)
		showErrorMessageBox(fmt.Sprintf("Erro: %s", message))
	}
	os.Exit(1)
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
	dbIP, _, err := k.GetStringValue("DBIp")
	if err != nil {
		return data, err
	}
	dbPort, _, err := k.GetStringValue("DBPort")
	if err != nil {
		return data, err
	}
	dbDatabase, _, err := k.GetStringValue("DBDatabase")
	if err != nil {
		return data, err
	}
	dbUser, _, err := k.GetStringValue("DBUser")
	if err != nil {
		return data, err
	}
	dbPwd, _, err := k.GetStringValue("DBPwd")
	if err != nil {
		return data, err
	}
	dbRole, _, err := k.GetStringValue("DBRole")
	if err != nil {
		return data, err
	}
	dbCompanyID, _, err := k.GetStringValue("DBCompanyID")
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
	data.DBIp = dbIP
	data.DBPort = dbPort
	data.DBDatabase = dbDatabase
	data.DBUser = dbUser
	data.DBPwd = dbPwd
	data.DBRole = dbRole
	data.DBCompanyID = dbCompanyID
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
	err = k.SetStringValue("DBIp", data.DBIp)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DBPort", data.DBPort)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DBDatabase", data.DBDatabase)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DBUser", data.DBUser)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DBPwd", data.DBPwd)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DBRole", data.DBRole)
	if err != nil {
		return err
	}
	err = k.SetStringValue("DBCompanyID", data.DBCompanyID)
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

func sendData(existingData Data, pumpsData PumpsData, apiConfig ApiConfig) {
	jsonData, err := json.Marshal(pumpsData)
	if err != nil {
		log.Println("Error marshaling data:", err)
		return
	}
	req, err := http.NewRequest("POST", apiConfig.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		log.Println("Status Code:", resp.StatusCode)
	}
}

func readDatabase(existingData Data, apiConfig ApiConfig) {
	pgConnStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable", existingData.DBUser, existingData.DBDatabase, existingData.DBPwd, existingData.DBIp, existingData.DBPort)
	pgDB, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		fatalError("Erro ao abrir o banco de dados:", err)
	}
	defer pgDB.Close()
	_, err = pgDB.Exec(fmt.Sprintf("SET SESSION AUTHORIZATION %s", pq.QuoteIdentifier(existingData.DBRole)))
	if err != nil {
		fatalError("Erro ao definir ROLE no banco de dados:", err)
	}
	// pgQuery := fmt.Sprintf("SELECT * FROM get_transactions_by_company_id(%s)", existingData.DBCompanyID)
	pgQuery := fmt.Sprintf(`SELECT
								a.cdabastecimento,
								a.dhabastecimento,
								a.qtvolume,
								a.vlunitario,
								a.vltotal,
								a.fllancado,
								b.dsbico,
								b.nrbico,
								e.dsapelido
							FROM
								dah.abastecimento a
							JOIN
								dah.bico b ON a.cdbico = b.cdbico
							JOIN
								dah.empresa e ON a.cdempresa = e.cdempresa
							WHERE
								a.fllancado = 0 and
								a.flcancelado = 0 and
								a.fltipo = 0 and
								a.dhabastecimento >= (NOW() - INTERVAL '12 hours' ) and
								a.cdempresa = %s
							ORDER BY
								a.dhprocessamento DESC`, existingData.DBCompanyID)
	rows, err := pgDB.Query(pgQuery)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	pumpRowData := []PumpRowsData{}
	for rows.Next() {
		pumpRow := PumpRows{}
		err := rows.Scan(
			&pumpRow.CdAbastecimento, &pumpRow.DhAbastecimento,
			&pumpRow.QtVolume, &pumpRow.VlUnitario, &pumpRow.VlTotal,
			&pumpRow.FlLancado, &pumpRow.DsBico, &pumpRow.NrBico,
			&pumpRow.DsApelido,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		pumpRowData = append(pumpRowData, PumpRowsData{
			GasStationTransactionID: pumpRow.CdAbastecimento,
			Quantity:                pumpRow.QtVolume,
			UnitValue:               pumpRow.VlUnitario,
			TotalValue:              pumpRow.VlTotal,
			Processed:               pumpRow.FlLancado,
			Date:                    pumpRow.DhAbastecimento.Format("2006-01-02 15:04:05"),
			PumpNumber:              pumpRow.NrBico,
			FuelName:                pumpRow.DsBico,
			CompanyName:             pumpRow.DsApelido,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	pumpsData := PumpsData{
		OrganizationCode: existingData.OrganizationCode,
		GasStationCode:   existingData.GasStationCode,
		PumpRowsData:     pumpRowData,
	}
	sendData(existingData, pumpsData, apiConfig)
}

func sendDataPeriodically() {
	existingData, err := readDataFromRegistry()
	if err != nil {
		fatalError("Erro ao ler dados do registro", err)
	}
	err = godotenv.Load()
	if err != nil {
		fatalError("Erro ao carregar o arquivo .env", err)
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

func main() {
	logPath := "./touchsistemas.log"
	logName := filepath.Base(logPath)
	logFile := &lumberjack.Logger{
		Filename:   logName,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))

	var (
		oCodeTE, gsCodeTE, dbCompanyIDTE, dbIPTE, dbPortTE, dbNameTE, dbUserTE, dbPwdTE, dbRoleTE *walk.TextEdit
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
					TextEdit{AssignTo: &dbIPTE, Text: existingData.DBIp},
					Label{Text: "Porta do Banco de Dados:"},
					TextEdit{AssignTo: &dbPortTE, Text: existingData.DBPort},
					Label{Text: "Nome do Banco de Dados:"},
					TextEdit{AssignTo: &dbNameTE, Text: existingData.DBDatabase},
					Label{Text: "Usuário do Banco de Dados:"},
					TextEdit{AssignTo: &dbUserTE, Text: existingData.DBUser},
					Label{Text: "Senha do Banco de Dados:"},
					TextEdit{AssignTo: &dbPwdTE, Text: existingData.DBPwd},
					Label{Text: "Grupo (Role) do Banco de Dados:"},
					TextEdit{AssignTo: &dbRoleTE, Text: existingData.DBRole},
					Label{Text: "ID do Posto no Banco de Dados:"},
					TextEdit{AssignTo: &dbCompanyIDTE, Text: existingData.DBCompanyID},
					Label{Text: "Código da Organização:"},
					TextEdit{AssignTo: &oCodeTE, Text: existingData.OrganizationCode},
					Label{Text: "Código do Posto:"},
					TextEdit{AssignTo: &gsCodeTE, Text: existingData.GasStationCode},
				},
			},
			PushButton{
				Text: "Continuar",
				OnClicked: func() {
					dbIP := strings.TrimSpace(dbIPTE.Text())
					dbPort := strings.TrimSpace(dbPortTE.Text())
					dbDatabase := strings.TrimSpace(dbNameTE.Text())
					dbUser := strings.TrimSpace(dbUserTE.Text())
					dbPassword := strings.TrimSpace(dbPwdTE.Text())
					dbRole := strings.TrimSpace(dbRoleTE.Text())
					dbCompanyID := strings.TrimSpace(dbCompanyIDTE.Text())
					organizationCode := strings.TrimSpace(oCodeTE.Text())
					gasStationCode := strings.TrimSpace(gsCodeTE.Text())
					existingData.DBIp = dbIP
					existingData.DBPort = dbPort
					existingData.DBDatabase = dbDatabase
					existingData.DBUser = dbUser
					existingData.DBPwd = dbPassword
					existingData.DBRole = dbRole
					existingData.DBCompanyID = dbCompanyID
					existingData.OrganizationCode = organizationCode
					existingData.GasStationCode = gasStationCode
					err := saveDataToRegistry(existingData)
					if err != nil {
						fatalError("Erro ao salvar dados no registro", err)
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

	defer logFile.Close()
	mw.Run()
}
