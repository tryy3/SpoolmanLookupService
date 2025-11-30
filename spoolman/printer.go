package spoolman

import (
	"errors"

	"github.com/tryy3/SpoolmanLookupService/models"
)

var (
	// TODO: For now we hard code this but the plan is to dynamically load this from the database
	printers []models.PrinterData = []models.PrinterData{
		{
			ID: "bl_p1s_ams_1",
			Name: "Bambulab P1S",
			Endpoint: "https://192.168.1.70",
			AMSSlot: 1,
			PrinterType: "bambu_lab_p1s",
		},
		{
			ID: "bl_p1s_ams_2",
			Name: "Bambulab P1S",
			Endpoint: "https://192.168.1.70",
			AMSSlot: 2,
			PrinterType: "bambu_lab_p1s",
		},
		{
			ID: "bl_p1s_ams_3",
			Name: "Bambulab P1S",
			Endpoint: "https://192.168.1.70",
			AMSSlot: 3,
			PrinterType: "bambu_lab_p1s",
		},
		{
			ID: "bl_p1s_ams_4",
			Name: "Bambulab P1S",
			Endpoint: "https://192.168.1.70",
			AMSSlot: 4,
			PrinterType: "bambu_lab_p1s",
		},
	}
)

var ErrPrinterNotFound = errors.New("printer not found")

type PrinterService struct {

}

func NewPrinterService() *PrinterService {
	return &PrinterService{}
}

func (s *PrinterService) GetPrinterData(id string) (models.PrinterData, error) {
	for _, printer := range printers {
		if printer.ID == id {
			return printer, nil
		}
	}
	return models.PrinterData{}, ErrPrinterNotFound
}