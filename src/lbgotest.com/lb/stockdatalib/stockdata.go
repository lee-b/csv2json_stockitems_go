package stockdatalib

import (
	"encoding/csv"
	"math/big"
)

type Monetary big.Rat

type Modifier struct {
	name  string
	price Monetary
}

type StockItem struct {
	item_id          int
	description      string
	price            Monetary
	cost             Monetary
	price_type       string
	quantity_on_hand int64
	modifiers        []Modifier
}

func (*StockItem) Unmarshall(reader csv.Reader) error {
	// TODO: read string->string value map from Reader,
	//       and populate item with its normalised values
	return nil
}
