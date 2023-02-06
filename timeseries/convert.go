package timeseries

import (
	"errors"

	"github.com/apache/arrow/go/v12/arrow"
	"github.com/segmentio/parquet-go"
)

func ParquetFieldToArrowField(pf parquet.Field) (arrow.Field, error) {
	typ, err := ParquetNodeToType(pf)
	if err != nil {
		return arrow.Field{}, err
	}
	return arrow.Field{
		Name:     pf.Name(),
		Type:     typ,
		Nullable: pf.Optional(),
	}, nil
}

// ParquetNodeToType converts a parquet node to an arrow type.
// Rules from: https://github.com/apache/parquet-format/blob/master/LogicalTypes.md
func ParquetNodeToType(n parquet.Node) (arrow.DataType, error) {
	if !n.Leaf() {
		switch {
		case hasMapFields(n):
			return mapType(n)
		case hasListFields(n):
			return listType(n)
		default:
			arrowFields := make([]arrow.Field, 0, len(n.Fields()))
			for _, f := range n.Fields() {
				af, err := ParquetFieldToArrowField(f)
				if err != nil {
					return nil, err
				}
				arrowFields = append(arrowFields, af)
			}
			return arrow.StructOf(arrowFields...), nil
		}
	}

	t := n.Type()
	lt := t.LogicalType()
	var dt arrow.DataType
	switch {
	case lt != nil:
		switch {
		case lt.UTF8 != nil:
			dt = &arrow.StringType{}
		case lt.Integer != nil:
			switch lt.Integer.BitWidth {
			case 64:
				if lt.Integer.IsSigned {
					dt = &arrow.Int64Type{}
				} else {
					dt = &arrow.Uint64Type{}
				}
			case 32:
				if lt.Integer.IsSigned {
					dt = &arrow.Int32Type{}
				} else {
					dt = &arrow.Uint32Type{}
				}
			default:
				return nil, errors.New("unsupported int bit width")
			}
		case lt.UUID != nil:
			dt = &arrow.FixedSizeBinaryType{
				ByteWidth: 16,
			}
		case lt.Timestamp != nil:
			dt = &arrow.Int64Type{}
		default:
			return nil, errors.New("unsupported logical type: " + n.Type().String())
		}
	case t.Kind() == parquet.Boolean:
		dt = &arrow.BooleanType{}
	case t.Kind() == parquet.Double:
		dt = &arrow.Float64Type{}
	default:
		return nil, errors.New("unsupported type: " + n.Type().String())
	}

	if n.Repeated() {
		dt = arrow.ListOf(dt)
	}

	return dt, nil
}

// https://github.com/apache/parquet-format/blob/master/LogicalTypes.md#maps
func hasMapFields(n parquet.Node) bool {
	// toplevel group requiredto be repeated group key_value with
	if !(len(n.Fields()) == 1 && n.Fields()[0].Repeated() && n.Fields()[0].Name() == "key_value") {
		return false
	}

	// can only be two fields, a key field and an optional value field
	if !(len(n.Fields()[0].Fields()) >= 1 && len(n.Fields()[0].Fields()) <= 2) {
		return false
	}

	// Find and validate key field
	for _, f := range n.Fields()[0].Fields() {
		if f.Name() == "key" && !f.Required() {
			return false
		}
	}

	return true
}

// https://github.com/apache/parquet-format/blob/master/LogicalTypes.md#lists
func hasListFields(n parquet.Node) bool {
	if !((n.Optional() || n.Required()) && len(n.Fields()) == 1) {
		return false
	}

	list := n.Fields()[0]
	if !(list.Name() == "list" && list.Repeated() && len(list.Fields()) == 1) {
		return false
	}

	element := list.Fields()[0]
	if !(element.Required() || element.Optional()) {
		return false
	}

	return true
}

func mapType(n parquet.Node) (arrow.DataType, error) {
	var key arrow.DataType
	var value arrow.DataType
	var err error
	value = &arrow.BooleanType{} // Default to boolean types for the value
	for _, field := range n.Fields()[0].Fields() {
		switch field.Name() {
		case "key":
			key, err = ParquetNodeToType(field)
			if err != nil {
				return nil, err
			}
		case "value":
			value, err = ParquetNodeToType(field)
			if err != nil {
				return nil, err
			}
		}
	}

	return arrow.MapOf(key, value), nil
}

func listType(n parquet.Node) (arrow.DataType, error) {
	listType, err := ParquetNodeToType(n.Fields()[0].Fields()[0])
	if err != nil {
		return nil, err
	}
	return arrow.ListOf(listType), nil
}
