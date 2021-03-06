package stockdatalib

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// helper function to build a reader object for all of our tests, from an
// in-memory string
func MakeDataReader(t *testing.T) *csv.Reader {
	csv_dat := "item id,description,price,cost,price_type,quantity_on_hand,modifier_1_name,modifier_1_price,modifier_2_name,modifier_2_price,modifier_3_name,modifier_3_price\n111010,Coffee,$1.25,$0.80,system,100000,Small,-$0.25,Medium,$0.00,Large,$0.30\n"
	buf := strings.NewReader(csv_dat)
	csv_reader := csv.NewReader(buf)
	return csv_reader
}

// helper function to load the first record/StockItem from our in-memory
// string/reader, as created by MakeDataReader()
func LoadFirstItem(t *testing.T) *StockItem {
	stock_dat_reader := MakeDataReader(t)

	columns_err := VerifyCsvFields(*stock_dat_reader)
	if columns_err != nil {
		t.Error(columns_err)
	}

	stock_item := new(StockItem)
	unmarshall_err := stock_item.ReadItem(*stock_dat_reader)
	if unmarshall_err != nil {
		t.Error(unmarshall_err)
	}

	return stock_item
}

// test that we can load stock data from memory
func TestStockDataCsvImport(t *testing.T) {
	LoadFirstItem(t)
}

// test that we can Parse Ids properly
func TestStockDataCSVParsedId(t *testing.T) {
	stock_item := LoadFirstItem(t)

	if stock_item.Item_id != 111010 {
		t.Fail()
	}
}

// test that we can Parse Descriptions properly
func TestStockDataCSVParsedDescription(t *testing.T) {
	stock_item := LoadFirstItem(t)

	if stock_item.Description != "Coffee" {
		t.Fail()
	}
}

// test that we can Parse Prices properly
func TestStockDataCSVParsedPrice(t *testing.T) {
	stock_item := LoadFirstItem(t)

	var expected_val Cents = 125

	if stock_item.Price != nil {
		if *stock_item.Price != expected_val {
			fmt.Printf("stock_item.Price is %d", *stock_item.Price)
			t.Fail()
		}
	} else {
		// price is nil; that's OK
	}
}

// test that we can Parse Costs properly
func TestStockDataCSVParsedCost(t *testing.T) {
	stock_item := LoadFirstItem(t)

	var expected_val Cents = 80

	if stock_item.Cost != nil {
		if *stock_item.Cost != expected_val {
			fmt.Printf("stock_item.Cost is %d", *stock_item.Cost)
			t.Fail()
		}
	} else {
		// cost is Nil; that's OK
	}
}

// test that we can Parse PriceTypes properly
func TestStockDataCSVParsedPriceType(t *testing.T) {
	stock_item := LoadFirstItem(t)

	switch stock_item.Price_type {
	case "system":
		// OK
	case "open":
		// OK
	default:
		// unknown price type
		t.Fail()
	}
}

// test that we can Encode JSON properly
func TestStockDataJsonEncode(t *testing.T) {
	stock_item := LoadFirstItem(t)

	var expected_dat = []byte("{\"id\":111010,\"description\":\"Coffee\",\"price\":1.25,\"cost\":0.80,\"price_type\":\"system\",\"quantity_on_hand\":100000,\"Modifiers\":[{\"Name\":\"Small\",\"Price\":-0.25},{\"Name\":\"Medium\",\"Price\":0.00},{\"Name\":\"Large\",\"Price\":0.30}]}")

	encoded_bytes, err := json.Marshal(stock_item)
	if err != nil {
		t.Error(err)
	}

	if len(encoded_bytes) != len(expected_dat) {
		err_msg := fmt.Sprintf("encoded_byte's length is wrong (encoded_bytes are:\n'%s'\n, expected bytes are:\n'%s'\n", encoded_bytes, expected_dat)
		t.Error(err_msg)
	}

	for idx, encoded_byte := range expected_dat {
		if encoded_byte != expected_dat[idx] {
			t.Fail()
		}
	}
}
