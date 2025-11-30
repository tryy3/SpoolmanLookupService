package spoolman

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tryy3/SpoolmanLookupService/models"
)

type SpoolmanClient struct {
	APIURL string
}

func NewSpoolmanClient(apiURL string) *SpoolmanClient {
	return &SpoolmanClient{
		APIURL: apiURL,
	}
}

func (c *SpoolmanClient) GetSpoolData(spoolId string) (models.SpoolmanSpoolData, error) {
	response, err := http.Get(fmt.Sprintf("%s/spool/%s", c.APIURL, spoolId))
	if err != nil {
		return models.SpoolmanSpoolData{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return models.SpoolmanSpoolData{}, err
	}
	log.Printf("Spoolman API response: %s", string(body))
	var spoolData models.SpoolmanSpoolData
	if err := json.Unmarshal(body, &spoolData); err != nil {
		return models.SpoolmanSpoolData{}, err
	}

	return spoolData, nil
}