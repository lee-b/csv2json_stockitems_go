package stockdatalib

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Cents int64

type PriceType int8

const (
	PT_SYSTEM = iota
	PT_OPEN   = iota
)

func PriceTypeFromString(s *string) (*PriceType, error) {
	//
	// parse price_type csv strings into enums:
	//
	//     "system" into PT_SYSTEM
	//     "open" to PT_OPEN, etc.
	//

	if len(*s) == 0 {
		return nil, errors.New("StockItem has an empty string for price_type!")
	}

	var pt PriceType

	switch *s {
	case "system":
		pt = PT_SYSTEM

	case "open":
		pt = PT_OPEN

	default:
		err_msg := fmt.Sprint("StockItem has an unrecognised price_type (%s)", s)
		return nil, errors.New(err_msg)
	}

	return &pt, nil
}

type Modifier struct {
	name  string
	price Cents
}

type ItemId int64

// NOTE: it'd probably be more memory efficient and
//       require less memory management to have
//       price, cost, etc. be structs with a boolean
//		 is_set flag, rather than hold points to the
//       values and do allocation for each price.
//       However, for the scope of test, this seems
//       like a reasonable way to get the results
//       expected, in reasonable time, without coding
//       for problems that haven't arisen yet (aka,
//       "You won't need it")
type StockItem struct {
	item_id          ItemId
	description      string
	price            *Cents
	cost             *Cents
	price_type       PriceType
	quantity_on_hand *int64
	modifiers        []Modifier
}

func CentsFromCsv(s *string) (*Cents, error) {
	//
	// parse "$1.99" into Cents(199)
	//

	if len(*s) == 0 {
		// value is empty, so we set the
		// cents pointer to nil.
		return nil, nil
	}

	if (*s)[0] != '$' {
		// handling other currencies are out of scope for
		// the test assumptions
		return nil, errors.New("Unrecognised currency!")
	}

	// remove dollar sign, now we know it's there
	val_str := (*s)[1:]

	// get integer and fractional parts
	parts := strings.Split(val_str, ".")
	if len(parts) < 1 {
		return nil, errors.New("Price is empty!")
	}

	int_part, int_part_err := strconv.ParseInt(parts[0], 10, 64)
	if int_part_err == nil {
		return nil, int_part_err
	}

	frac_part, frac_part_err := strconv.ParseInt(parts[1], 10, 64)
	if frac_part_err == nil {
		return nil, frac_part_err
	}

	val := int64(int_part * 100)
	val = val + frac_part

	cents := Cents(val)

	return &cents, nil
}

func VerifyCsvFields(reader csv.Reader) error {
	raw_dat, err := reader.Read()
	if err != nil {
		return err
	}

	expected_fields := []string{
		"item id", "description", "price", "cost", "price_type", "quantity_on_hand",
		"modifier_1_name", "modifier_1_price", "modifier_2_name",
		"modifier_2_price", "modifier_3_name", "modifier_3_price",
	}

	// check the number of fields matches
	num_fields := len(raw_dat)
	num_expected_fields := len(expected_fields)

	if num_fields != num_expected_fields {
		err_msg := fmt.Sprintf("Expected %s columns/fields in CSV file.  Saw %d fields on the first line.", num_expected_fields, num_fields)
		return errors.New(err_msg)
	}

	var last_field *string = nil

	for i := range expected_fields {
		if expected_fields[i] != raw_dat[i] {
			var err_msg string

			if last_field != nil {
				err_msg = fmt.Sprintf("CSV file's fields don't match the expected format.  Expected field %s after %s, not %s", expected_fields[i], last_field, raw_dat[i])
			} else {
				err_msg = fmt.Sprintf("CSV file's fields don't match the expected format.  Expected the first field to be %s but got %s", expected_fields[i], raw_dat[i])
			}

			return errors.New(err_msg)
		}
	}

	// if we reach here, all is ok
	return nil
}

func (item *StockItem) Unmarshall(reader csv.Reader) error {
	// De-serialise the CSV row into a StockItem object.  Ensure
	// data is valid in the process, BEFORE storing it as a valid
	// object in mem.

	raw_dat, read_err := reader.Read()
	if read_err != nil {
		return read_err
	}

	// parse the fields out of the CSV string values
	item_id_int64, item_id_err := strconv.ParseInt(raw_dat[0], 10, 64)
	if item_id_err != nil {
		return item_id_err
	}
	item.item_id = ItemId(item_id_int64)

	item.description = raw_dat[1]

	var price_err error
	item.price, price_err = CentsFromCsv(&raw_dat[2])
	if price_err != nil {
		return price_err
	}

	var cost_err error
	item.cost, cost_err = CentsFromCsv(&raw_dat[3])
	if cost_err != nil {
		return cost_err
	}

	price_type_ptr, price_type_err := PriceTypeFromString(&raw_dat[4])
	if price_type_err != nil {
		return price_type_err
	}
	item.price_type = *price_type_ptr

	return nil
}
