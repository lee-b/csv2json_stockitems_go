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

type Quantity int64

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

func QuantityFromString(s *string) (*Quantity, error) {
	if len(*s) == 0 {
		// no quantity, but that's OK
		return nil, nil
	}

	val, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return nil, err
	}

	quantity := Quantity(val)

	return &quantity, nil
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
	Item_id          ItemId             `json:"id"`
	Description      string             `json:"description"`
	Price            *Cents             `json:"price"`
	Cost             *Cents             `json:"cost"`
	Price_type       PriceType          `json:"price_type"`
	Quantity_on_hand *Quantity          `json:"quantity_on_hand"`
	Modifiers        [4]Modifier
}

func CentsFromDollarString(s *string) (*Cents, error) {
	//
	// parse "$1.99" into Cents(199)
	//
	// NOTE: I do it this way because strconv.ParseFloat() would
	//       introduce precision errors
	//

	str := (*s)[:]

	is_negative := false

	// strip the minus and set a flag for it, if it exists
	if str[0] == '-' {
		is_negative = true
		str = str[1:]
	}

	// expect and handle the dollar sign, or generate an error
	if str[0] == '$' {
		str = str[1:]
	} else {
		// expected a dollar, didn't see one
		err_msg := fmt.Sprintf("No dollar sign at '%x' from price string '%s'. Perhaps this is an unhandled currency?  Currency symbols are expected.", str, *s)
		return nil, errors.New(err_msg)
	}

	parts := strings.Split(str, ".")
	num_parts := len(parts)
	if num_parts == 0 || num_parts > 2 {
		err_msg := fmt.Sprintf("too many decimal points in currency value '%s'", *s)
		return nil, errors.New(err_msg)
	}

	// handle the integer part (dollars)
	dollar_val, dollar_err := strconv.ParseInt(parts[0], 10, 64)
	if dollar_err != nil {
		err_msg := fmt.Sprintf("Couldn't parse integer part ('%s') of dollar value '%s'", parts[0], *s)
		return nil, errors.New(err_msg)
	}

	var frac_val int64 = 0
	var frac_err error = nil

	if num_parts == 2 {
		// handle the integer part (dollars)
		frac_val, frac_err = strconv.ParseInt(parts[1], 10, 64)
		if frac_err != nil {
			err_msg := fmt.Sprintf("Couldn't parse fractional part ('%s') of dollar value '%s'", parts[1], *s)
			return nil, errors.New(err_msg)
		}
	}

	if frac_val < 0 || frac_val > 99 {
		if frac_err != nil {
			err_msg := fmt.Sprintf("Fractional part of dollar value '%s' does NOT appear to be a valid number of cents.  0-99 is assumed!", *s)
			return nil, errors.New(err_msg)
		}
	}

	val := (dollar_val * 100) + frac_val

	// flip the value's sign if we saw a minus in the string
	if is_negative {
		val = -val
	}

	cents := new(Cents)
	*cents = Cents(val)

	return cents, nil
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

type Modifier struct {
	Name  *string
	Price Cents
}

func ModifierFromStrings(name_str_ptr, price_str_ptr *string) (*Modifier, error) {
	if len(*name_str_ptr) == 0 {
		// no modifier entry, so just return an empty one,
		// with no error
		return &Modifier{nil, 0}, nil
	}

	cents_ptr, cents_err := CentsFromDollarString(price_str_ptr)
	if cents_err != nil {
		return nil, cents_err
	}

	if cents_ptr == nil {
		err_msg := fmt.Sprintf("No cents found in dollar string %s, for modifier %s.  Cents are expected here.", *price_str_ptr, *name_str_ptr)
		err := errors.New(err_msg)
		return nil, err
	}

	mod_ptr := new(Modifier)
	mod_ptr.Name = name_str_ptr
	mod_ptr.Price = *cents_ptr

	return mod_ptr, nil
}

func (item *StockItem) Unmarshall(reader csv.Reader) error {
	//
	// De-serialise the CSV row into a StockItem object.  Ensure
	// data is valid in the process, BEFORE storing it as a valid
	// object in mem.
	//

	raw_dat, read_err := reader.Read()
	if read_err != nil {
		return read_err
	}

	// used frequently later
	raw_dat_len := len(raw_dat)

	// parse the fields out of the CSV string values
	item_id_int64, item_id_err := strconv.ParseInt(raw_dat[0], 10, 64)
	if item_id_err != nil {
		return item_id_err
	}
	item.Item_id = ItemId(item_id_int64)

	item.Description = raw_dat[1]

	var price_err error
	item.Price, price_err = CentsFromDollarString(&raw_dat[2])
	if price_err != nil {
		return price_err
	}

	var cost_err error
	item.Cost, cost_err = CentsFromDollarString(&raw_dat[3])
	if cost_err != nil {
		return cost_err
	}

	price_type_ptr, price_type_err := PriceTypeFromString(&raw_dat[4])
	if price_type_err != nil {
		return price_type_err
	}
	item.Price_type = *price_type_ptr

	quantity_ptr, quantity_err := QuantityFromString(&raw_dat[5])
	if quantity_err != nil {
		return quantity_err
	}
	item.Quantity_on_hand = quantity_ptr

	// load modifiers
	const all_modifiers_start_idx = 6

	for i := 0; i <= 3; i++ {
		modifier_idx := all_modifiers_start_idx + (i * 2)

		if raw_dat_len-1 < modifier_idx {
			// modifier not present; quit loop here
			break
		}

		if raw_dat_len-1 < modifier_idx+1 {
			// half of the modifier is present, so report an error
			name := raw_dat[modifier_idx]
			err_msg := fmt.Sprintf("StockItem id %d's modifier #%d (name: %s) has only one of two fields present. Expected all fields, or none.", name, item.Item_id, i)
			return errors.New(err_msg)
		}

		name := &raw_dat[modifier_idx]
		price := &raw_dat[modifier_idx+1]

		mod, mod_error := ModifierFromStrings(name, price)
		if mod_error != nil {
			return mod_error
		}

		// mod loaded; set in modifiers array
		item.Modifiers[i] = *mod
	}

	return nil
}
