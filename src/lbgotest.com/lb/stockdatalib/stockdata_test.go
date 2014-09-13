package stockdatalib

import (
	"encoding/csv"
	//	"encoding/json"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func MakeDataReader(t *testing.T) *csv.Reader {
	csv_dat := "item id,description,price,cost,price_type,quantity_on_hand,modifier_1_name,modifier_1_price,modifier_2_name,modifier_2_price,modifier_3_name,modifier_3_price\n111010,Coffee,$1.25,$0.80,system,100000,Small,-$0.25,Medium,$0.00,Large,$0.30\n"
	buf := strings.NewReader(csv_dat)
	csv_reader := csv.NewReader(buf)
	return csv_reader
}

func LoadFirstItem(t *testing.T) *StockItem {
	stock_dat_reader := MakeDataReader(t)

	columns_err := VerifyCsvFields(*stock_dat_reader)
	if columns_err != nil {
		t.Error(columns_err)
	}

	stock_item := new(StockItem)
	unmarshall_err := stock_item.Unmarshall(*stock_dat_reader)
	if unmarshall_err != nil {
		t.Error(unmarshall_err)
	}

	return stock_item
}

func TestStockDataCsvImport(t *testing.T) {
	LoadFirstItem(t)
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

	var expected_val Cents = 125

	if stock_item.price != nil {
		if *stock_item.price != expected_val {
			fmt.Printf("stock_item.price is %d", *stock_item.price)
			t.Fail()
		}
	} else {
		// price is nil; that's OK
	}
}

func TestStockDataCSVUnmarshalledCost(t *testing.T) {
	stock_item := LoadFirstItem(t)

	var expected_val Cents = 80

	if stock_item.cost != nil {
		if *stock_item.cost != expected_val {
			fmt.Printf("stock_item.cost is %d", *stock_item.cost)
			t.Fail()
		}
	} else {
		// cost is Nil; that's OK
	}
}

func TestStockDataCSVUnmarshalledPriceType(t *testing.T) {
	stock_item := LoadFirstItem(t)

	switch stock_item.price_type {
	case PT_SYSTEM:
		// OK
	case PT_OPEN:
		// OK
	default:
		// unknown price type
		t.Fail()
	}
}

func TestStockDataJsonEncode(t *testing.T) {
	stock_item := LoadFirstItem(t)

	b, err := json.Marshal(stock_item)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("JSON-encoded StockData looks like:\n\n\t%s\n\n", b)
	}
}
