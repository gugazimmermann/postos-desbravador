package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"golang.org/x/sys/windows/registry"
)

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

type Pumps struct {
	CdAbastecimento int
	DhAbastecimento time.Time
	DhProcessamento time.Time
	QtVolume        float64
	VlUnitario      float64
	VlTotal         float64
	CdBico          int
	CdEmpresa       int
	FlLancado       int
	FlCancelado     int
	FlTipo          int
	DsBico          string
	DsApelido       string
}

type PumpData struct {
	OrganizationCode        string
	GasStationCode          string
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

type ApiConfig struct {
	Url string
}

func main() {
	sendDataPeriodically()
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

func sendData(existingData Data, pumpsData []PumpData, apiConfig ApiConfig) {
	jsonData, err := json.Marshal(pumpsData)
	if err != nil {
		log.Println("Error marshaling data:", err)
		return
	}
	log.Println("Sending Data")
	log.Println(apiConfig)
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
	if resp.StatusCode != http.StatusCreated {
		log.Println("Unexpected status code:", resp.StatusCode)
	} else {
		log.Println("Data sent successfully:", pumpsData)
	}
}

func readDatabase(existingData Data, apiConfig ApiConfig) {
	log.Println("Getting Data")
	pgConnStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable", existingData.DBUser, existingData.DBDatabase, existingData.DBPwd, existingData.DBIp, existingData.DBPort)
	pgDB, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Println(err)
		return
	}
	defer pgDB.Close()
	_, err = pgDB.Exec(fmt.Sprintf("SET SESSION AUTHORIZATION %s", pq.QuoteIdentifier(existingData.DBRole)))
	if err != nil {
		log.Fatal(err)
		return
	}
	// pgQuery := fmt.Sprintf("SELECT * FROM get_transactions_by_company_id(%s)", existingData.DBCompanyID)
	pgQuery := fmt.Sprintf(`SELECT
								a.cdabastecimento,
								a.dhabastecimento,
								a.dhprocessamento,
								a.qtvolume,
								a.vlunitario,
								a.vltotal,
								a.cdbico,
								a.cdempresa,
								a.fllancado,
								a.flcancelado,
								a.fltipo,
								b.dsbico,
								e.dsapelido
							FROM
								dah.abastecimento a
							JOIN
								dah.bico b ON a.cdbico = b.cdbico
							JOIN
								dah.empresa e ON a.cdempresa = e.cdempresa
							WHERE
								a.flcancelado = 0 and
								a.fltipo = 0 and
								a.dhabastecimento >= (NOW() - INTERVAL '1 hour' ) and
								a.cdempresa = %s
							ORDER BY
								a.cdabastecimento DESC`, existingData.DBCompanyID)
	rows, err := pgDB.Query(pgQuery)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	pumpsData := []PumpData{}
	for rows.Next() {
		pumps := Pumps{}
		err := rows.Scan(
			&pumps.CdAbastecimento, &pumps.DhAbastecimento,
			&pumps.DhProcessamento, &pumps.QtVolume,
			&pumps.VlUnitario, &pumps.VlTotal,
			&pumps.CdBico, &pumps.CdEmpresa,
			&pumps.FlLancado, &pumps.FlCancelado,
			&pumps.FlTipo, &pumps.DsBico, &pumps.DsApelido,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		pumpsData = append(pumpsData, PumpData{
			OrganizationCode:        existingData.OrganizationCode,
			GasStationCode:          existingData.GasStationCode,
			GasStationTransactionID: pumps.CdAbastecimento,
			Quantity:                pumps.QtVolume,
			UnitValue:               pumps.VlUnitario,
			TotalValue:              pumps.VlTotal,
			Processed:               pumps.FlLancado,
			Date:                    pumps.DhAbastecimento.Format("2006-01-02 15:04:05"),
			PumpNumber:              pumps.CdBico,
			FuelName:                pumps.DsBico,
			CompanyName:             pumps.DsApelido,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	if len(pumpsData) > 0 {
		sendData(existingData, pumpsData, apiConfig)
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
