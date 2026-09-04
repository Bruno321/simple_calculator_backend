package model

// OperandRequest is used by operations that accept a list of operands.
// A pointer distinguishes an omitted key from an explicitly supplied empty list.
type OperandRequest struct {
	Operands *[]*float64 `json:"operands"`
}

type ExponentiationRequest struct {
	Base     *float64 `json:"base"`
	Exponent *float64 `json:"exponent"`
}

type SquareRootRequest struct {
	Radicand *float64 `json:"radicand"`
}

type PercentageRequest struct {
	Value      *float64 `json:"value"`
	Percentage *float64 `json:"percentage"`
}
