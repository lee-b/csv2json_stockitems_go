//
// Basic library for handling StockItem data in memory, loading
// from CSV, and writing to JSON format.
//
// Copyright (c) 2014 by Lee Braiden (leebraid@gmail.com)
//
package stockdatalib

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Monetary value, in cents rather than dollars
// This avoids precision issues with floating-point values.
type Cents int64

// number of items of stock
type Quantity int64

// Parses a Quantity from a string
func QuantityFromString(s string) (*Quantity, error) {
	if len(s) == 0 {
		// no quantity, but that's OK
		return nil, nil
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}

	quantity := Quantity(val)

	return &quantity, nil
}

// (presumably) unique identifier for an item of stock, at least within
// one organisation
type ItemId int64

// Represents an Item of stock.  Prices are represented in cents to avoid
// floating-point rounding issues, or as nil if no such price information
// exists.  Modifiers is a dynamically sized slice into an array of 0--4
// Modifier items.  Quantities are represented as 64-bit integers.  Fractions
// are assumed to be NOT needed; use smaller units in that case, presumably.
// Negative values ARE allowed in case needed for adjustment items etc.,
// since there is no reason to assume quantities can reach the max / min of
// a 64-bit signed integer's range.
//
// NOTE: it'd probably be more memory efficient and require less memory-
//       management to have price, cost, etc. be structs with a boolean
//		 is_set flag, rather than hold points to the values and do allocation
//       for each price. However, for the scope of test, this seems like a
//       reasonable internal representation, to get the results expected in
//       reasonable time via encoding/json, without coding for problems that
//       haven't arisen yet (aka, YAGNI / "You ain't gonna need it")
//
type StockItem struct {
	Item_id          ItemId    `json:"id"`
	Description      string    `json:"description"`
	Price            *Cents    `json:"price"`
	Cost             *Cents    `json:"cost"`
	Price_type       string    `json:"price_type"`
	Quantity_on_hand *Quantity `json:"quantity_on_hand"`
	Modifiers        []Modifier
}

// JSON encoder for Cents types.  This outputs the cent value as a dollar
// value, with 2 digits for cents, and no dollar sign.  If negative, a minus
// sign is prepended.
func (c *Cents) MarshalJSON() ([]byte, error) {
	// NOTE: I wanted to override the marshalling of the POINTER type itself
	//       here, hoping encoders/json would look at the field type, then
	//       directly look for a matching MarshalJSON.  I tried:
	//           func (c **Cents) MarshalJSON() ...
	//       too.  Neither do what I want though.  I guess MarshalJSON
	//       is called with addresses of normal types, rather than being
	//       called FOR the type specified in the first-place arg (*Cents)
	//       in this case. Pity, as it would have made the "nil" output,
	//       as specified in the output example, easy, while still using
	//       Go's built-in libraries.
	//       Rather than write ugly code that doesn't use JSON properly,
	//       I've assumed that letting Go's libraries write null instead
	//       is probably fine.
	sign := ""
	if (*c) < 0 {
		sign = "-"
	}
	dollars_only := int64(math.Abs(float64((*c) / 100)))
	cents_only := int64(math.Abs(float64((*c) % 100)))
	c_as_str := fmt.Sprintf("%s%d.%02d", sign, dollars_only, cents_only)
	return []byte(c_as_str), nil
}

// Parse values such as "-$1.99" into Cents(-199)
//
// NOTE: I do it this way because strconv.ParseFloat() would
//       introduce precision errors.  Another option might be Sscanf(), but
//       this should be faster for all we need to do.
//
func CentsFromDollarString(s string) (*Cents, error) {
	// early exit with nil cents value, if the string is empty
	if len(s) == 0 {
		return nil, nil
	}

	// default to positive numbers, unless we see a minus sign
	is_negative := false

	// get a slice from the string value, so we can modify the slice / view
	// of the string, as we parse more of it
	str := s[:]

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
		//err_msg := fmt.Sprintf("No dollar sign at '%x' from price string '%s'. Perhaps this is an unhandled currency?  Currency symbols are expected.", str, s)
		//return nil, errors.New(err_msg)
		// missing dollar sign, but some data is like this.  Continue, regardless.
	}

	// split the x.xx value into whole dollars and fractions (dollars and
	// cents)
	parts := strings.Split(str, ".")
	num_parts := len(parts)
	if num_parts == 0 || num_parts > 2 {
		err_msg := fmt.Sprintf("too many decimal points in currency value '%s'", s)
		return nil, errors.New(err_msg)
	}

	// handle the integer part (dollars)
	var dollar_val int64 = 0
	if len(parts[0]) > 0 {
		var dollar_err error
		dollar_val, dollar_err = strconv.ParseInt(parts[0], 10, 64)
		if dollar_err != nil {
			err_msg := fmt.Sprintf("Couldn't parse integer part ('%s') of dollar value '%s'", parts[0], s)
			return nil, errors.New(err_msg)
		}
	}

	var frac_val int64 = 0
	var frac_err error = nil

	if num_parts == 2 {
		// handle the integer part (dollars)
		frac_val, frac_err = strconv.ParseInt(parts[1], 10, 64)
		if frac_err != nil {
			err_msg := fmt.Sprintf("Couldn't parse fractional part ('%s') of dollar value '%s'", parts[1], s)
			return nil, errors.New(err_msg)
		}
	}

	if frac_val < 0 || frac_val > 99 {
		if frac_err != nil {
			err_msg := fmt.Sprintf("Fractional part of dollar value '%s' does NOT appear to be a valid number of cents.  0-99 is assumed!", s)
			return nil, errors.New(err_msg)
		}
	}

	// build the cents value from individual dollar and fractions-of-dollar
	// values
	val := (dollar_val * 100) + frac_val

	// flip the value's sign if we saw a minus in the string
	if is_negative {
		val = -val
	}

	cents := new(Cents)
	*cents = Cents(val)

	return cents, nil
}

// Reads the first (title) row of a CSV StockItems file, and verifies that the
// field titles match what's expected.  Essentially, this checks that we're
// loading the right TYPE of file, before reading all the items.
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

// Individual modifiers for an item.
type Modifier struct {
	Name  string
	Price Cents
}

// Parse two strings into a Modifier object
func ModifierFromStrings(name_str, price_str string) (*Modifier, error) {
	if len(name_str) == 0 {
		// no modifier name, so raise an error
		return nil, errors.New("Attempted to create a modifier with no name.")
	}

	cents_ptr, cents_err := CentsFromDollarString(price_str)
	if cents_err != nil {
		return nil, cents_err
	}

	if cents_ptr == nil {
		err_msg := fmt.Sprintf("No cents found in dollar string %s, for modifier %s.  Cents are expected here.", price_str, name_str)
		err := errors.New(err_msg)
		return nil, err
	}

	mod_ptr := new(Modifier)
	mod_ptr.Name = name_str
	mod_ptr.Price = *cents_ptr

	return mod_ptr, nil
}

// Read an individual StockItem record from a CSV file, parsing it into
// a higher-level StockItem struct.
func (item *StockItem) ReadItem(reader csv.Reader) error {
	//
	// De-serialise the CSV row into a StockItem object.  Ensure
	// data is valid in the process, BEFORE storing it as a valid
	// object in mem.
	//
	const MIN_STOCK_ITEM_FIELDS = 5

	raw_dat, read_err := reader.Read()
	if read_err != nil {
		return read_err
	}

	// used frequently later
	raw_dat_len := len(raw_dat)

	if raw_dat_len < MIN_STOCK_ITEM_FIELDS {
		err_msg := fmt.Sprintf("StockItem has too few fields (%d); expected at least %d (up to price_type).", raw_dat_len, MIN_STOCK_ITEM_FIELDS, raw_dat, raw_dat_len)
		return errors.New(err_msg)
	}

	// parse the fields out of the CSV string values
	item_id_int64, item_id_err := strconv.ParseInt(raw_dat[0], 10, 64)
	if item_id_err != nil {
		return item_id_err
	}
	item.Item_id = ItemId(item_id_int64)

	item.Description = raw_dat[1]

	var price_err error
	item.Price, price_err = CentsFromDollarString(raw_dat[2])
	if price_err != nil {
		return price_err
	}

	var cost_err error
	item.Cost, cost_err = CentsFromDollarString(raw_dat[3])
	if cost_err != nil {
		return cost_err
	}

	switch raw_dat[4] {
	case "system":
		item.Price_type = "system"
	case "open":
		item.Price_type = "open"
	default:
		err_msg := fmt.Sprintf("Invalid price_type value: '%s'", raw_dat[4])
		return errors.New(err_msg)
	}

	if raw_dat_len >= (MIN_STOCK_ITEM_FIELDS + 1) {
		quantity_ptr, quantity_err := QuantityFromString(raw_dat[5])
		if quantity_err != nil {
			return quantity_err
		}
		item.Quantity_on_hand = quantity_ptr
	} else {
		item.Quantity_on_hand = nil
	}

	// load modifiers
	const all_modifiers_start_idx = 6

	// calculate how many modifiers we have (we should have n modifier fields,
	// which are in pairs of 2, so n / 2 complete modifier entries)
	num_modifiers := (raw_dat_len - all_modifiers_start_idx) / 2
	if num_modifiers > 4 {
		num_modifiers = 4
	}

	// make the list of modifiers as big as it can be, based on the number of
	// modifier fields present
	item.Modifiers = make([]Modifier, num_modifiers)

	// loop over all modifiers, extracting and parsing their field pairs into
	// Modifier structs, and store them in item.Modifiers
	for i := 0; i <= num_modifiers; i++ {
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

		name := raw_dat[modifier_idx]
		price := raw_dat[modifier_idx+1]

		mod, mod_error := ModifierFromStrings(name, price)
		if mod_error != nil {
			return mod_error
		}

		// mod loaded; set in modifiers array
		item.Modifiers[i] = *mod
	}

	// If we reach here, everything went well.  Return non-error.
	return nil
}
