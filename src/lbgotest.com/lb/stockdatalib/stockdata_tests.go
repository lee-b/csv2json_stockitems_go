package stockdatalib

import (
	"encoding/csv"
	//	"encoding/json"
	"math/big"
	"strings"
	"testing"
)

func MakeDataReader(t *testing.T) *csv.Reader {
	csv_dat := "item id,description,price,cost,price_type,quantity_on_hand,modifier_1_name,modifier_1_price,modifier_2_name,modifier_2_price,modifier_3_name,modifier_3_price\n111010,Coffee,$1.25,$0.80,system,100000,Small,-­‐$0.25,Medium,$0.00,Large,$0.30\n"
	buf := strings.NewReader(csv_dat)
	csv_reader := csv.NewReader(buf)
	return csv_reader
}

func LoadFirstItem(t *testing.T) *StockItem {
	stock_dat_reader := MakeDataReader(t)

	stock_item := new(StockItem)
	unmarshall_err := stock_item.Unmarshall(*stock_dat_reader)
	if unmarshall_err != nil {
		t.Error("Couldn't unmarshall the first stock item")
	}

	return stock_item
}

func TestStockDataCsvImport(t *testing.T) {
	rec := LoadFirstItem(t)
}

func TestStockDataCSVUnmarshalledId(t *testing.T) {
	stock_item := LoadFirstItem(t)

	if stock_item.item_id != 111010 {
		t.Fail()
	}
}

func TestStockDataCSVUnmarshalledDescription(t *testing.T) {
	stock_item := LoadFirstItem(t)

	if stock_item.description != "Coffee" {
		t.Fail()
	}
}

func TestStockDataCSVUnmarshalledPrice(t *testing.T) {
	stock_item := LoadFirstItem(t)

	var expected_val big.Rat
	expected_val.SetString("1.25", 10)

	if big.Rat(stock_item.price) != expected_val {
		t.Fail()
	}
}

func TestStockDataCSVUnmarshalledCost(t *testing.T) {
	stock_item := LoadFirstItem(t)

	var expected_val big.Rat
	expected_val.SetString("0.80", 10)

	if big.Rat(stock_item.cost) != expected_val {
		t.Fail()
	}
}
