# Decimal Usage Guide for Money Calculations

This guide explains how to use `shopspring/decimal` for precise money calculations in the payment system.

## Why Decimal for Money?

**Never use float64 for money calculations!**

```go
// ❌ WRONG - Floating point errors
var price float64 = 0.1
var total float64 = price * 3
fmt.Println(total) // Output: 0.30000000000000004 (!!!)

// ✅ CORRECT - Exact decimal arithmetic
price := decimal.NewFromFloat(0.1)
total := price.Mul(decimal.NewFromInt(3))
fmt.Println(total.String()) // Output: 0.3
```

## Core Principles

1. **Store amounts as int64** (smallest currency unit: cents, tenge, etc.)
2. **Use decimal for calculations** (conversions, percentages, splits)
3. **Convert to float64 only when calling external APIs** (if required)

## Common Operations

### Creating Decimals

```go
import "github.com/shopspring/decimal"

// From int64 (our standard storage format)
amountCents := int64(10050) // 100.50 KZT
dec := decimal.NewFromInt(amountCents)  // 10050

// From float (use sparingly, only for external API responses)
dec := decimal.NewFromFloat(100.50)

// From string (good for user input or config)
dec, err := decimal.NewFromString("100.50")
if err != nil {
    // Handle invalid decimal
}

// Constants
zero := decimal.Zero              // 0
one := decimal.NewFromInt(1)      // 1
hundred := decimal.NewFromInt(100) // 100
```

### Conversions

```go
// Convert from cents/tenge to decimal amount
amountCents := int64(10050) // Stored as 10050 (100.50 KZT)
amountDecimal := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))
fmt.Println(amountDecimal.String()) // "100.50"

// Convert decimal to cents/tenge for storage
amountDecimal := decimal.NewFromFloat(100.50)
amountCents := amountDecimal.Mul(decimal.NewFromInt(100)).IntPart()
// amountCents = 10050

// Convert to string for display
formatted := amountDecimal.StringFixed(2) // "100.50" (always 2 decimals)

// Convert to float64 (only for external APIs)
floatValue, exact := amountDecimal.Float64()
if !exact {
    // Value couldn't be exactly represented as float64
    log.Warn("precision loss in float conversion")
}
```

### Arithmetic Operations

```go
// Addition
price1 := decimal.NewFromInt(10050) // 100.50 KZT
price2 := decimal.NewFromInt(5025)  // 50.25 KZT
total := price1.Add(price2)         // 150.75 KZT (15075 cents)

// Subtraction
refund := total.Sub(price1)         // 50.25 KZT

// Multiplication
taxRate := decimal.NewFromFloat(0.12) // 12% tax
tax := price1.Mul(taxRate)             // 12.06 KZT

// Division
split := total.Div(decimal.NewFromInt(2)) // 75.375 KZT (split in half)

// Rounding (important!)
splitRounded := split.Round(2) // 75.38 KZT (round to 2 decimals)

// Negation
negative := price1.Neg() // -100.50 KZT
```

### Comparisons

```go
price1 := decimal.NewFromInt(10050)
price2 := decimal.NewFromInt(5025)

// Equality
if price1.Equal(price2) {
    // Amounts are equal
}

// Greater than
if price1.GreaterThan(price2) {
    // price1 > price2
}

// Less than
if price1.LessThan(price2) {
    // price1 < price2
}

// Greater than or equal
if price1.GreaterThanOrEqual(price2) {
    // price1 >= price2
}

// Compare returns -1, 0, or 1
switch price1.Cmp(price2) {
case -1:
    // price1 < price2
case 0:
    // price1 == price2
case 1:
    // price1 > price2
}

// Check if zero
if price.IsZero() {
    // Amount is zero
}

// Check if positive/negative
if price.IsPositive() {
    // Amount > 0
}

if price.IsNegative() {
    // Amount < 0
}
```

### Rounding

```go
amount := decimal.NewFromFloat(100.555)

// Round to 2 decimal places
rounded := amount.Round(2) // 100.56

// Round down (floor)
floor := amount.Floor() // 100.00

// Round up (ceiling)
ceil := amount.Ceil() // 101.00

// Truncate (remove decimal part)
truncated := amount.Truncate(2) // 100.55

// Round half up (banker's rounding)
rounded := amount.RoundBank(2) // 100.56
```

## Payment System Examples

### Example 1: Format Amount for Display

```go
// internal/payments/domain/service.go
func (s *Service) FormatAmount(amount int64, currency string) string {
    // Convert from cents/tenge to decimal
    amountDecimal := decimal.NewFromInt(amount).Div(decimal.NewFromInt(100))
    formatted := amountDecimal.StringFixed(2) // Always 2 decimals

    switch currency {
    case "KZT":
        return fmt.Sprintf("%s KZT", formatted)
    case "USD", "EUR":
        return fmt.Sprintf("%s %s", formatted, currency)
    default:
        return fmt.Sprintf("%d %s", amount, currency)
    }
}

// Usage:
amount := int64(10050) // 100.50 KZT in cents
formatted := service.FormatAmount(amount, "KZT")
// Output: "100.50 KZT"
```

### Example 2: Partial Refund Calculation

```go
// internal/payments/service/payment/refund_payment.go
func (uc *RefundPaymentUseCase) calculatePartialRefund(
    originalAmount int64,
    refundPercent float64,
) int64 {
    // Convert to decimal for precise calculation
    original := decimal.NewFromInt(originalAmount)
    percent := decimal.NewFromFloat(refundPercent)

    // Calculate refund amount
    refund := original.Mul(percent).Round(0) // Round to nearest cent

    return refund.IntPart()
}

// Usage:
originalAmount := int64(10000) // 100.00 KZT
refundPercent := 0.5           // 50% refund

refundAmount := calculatePartialRefund(originalAmount, refundPercent)
// refundAmount = 5000 (50.00 KZT)
```

### Example 3: Split Payment Among Multiple Parties

```go
func SplitPaymentEqually(totalAmount int64, numParties int) []int64 {
    total := decimal.NewFromInt(totalAmount)
    parties := decimal.NewFromInt(int64(numParties))

    // Calculate equal split
    perParty := total.Div(parties).Round(0)

    // Handle rounding remainder
    splits := make([]int64, numParties)
    sum := decimal.Zero

    for i := 0; i < numParties-1; i++ {
        splits[i] = perParty.IntPart()
        sum = sum.Add(perParty)
    }

    // Last party gets the remainder to ensure exact total
    splits[numParties-1] = total.Sub(sum).IntPart()

    return splits
}

// Usage:
total := int64(10001) // 100.01 KZT
splits := SplitPaymentEqually(total, 3)
// splits = [3334, 3334, 3333] (33.34, 33.34, 33.33 KZT)
// sum = 10001 (exact!)
```

### Example 4: Currency Conversion

```go
func ConvertCurrency(amountCents int64, fromCurrency, toCurrency string, exchangeRate float64) int64 {
    // Convert to decimal for precise calculation
    amount := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))
    rate := decimal.NewFromFloat(exchangeRate)

    // Convert
    converted := amount.Mul(rate)

    // Round to 2 decimals and convert back to cents
    convertedCents := converted.Mul(decimal.NewFromInt(100)).Round(0)

    return convertedCents.IntPart()
}

// Usage:
amountKZT := int64(50000) // 500.00 KZT
exchangeRate := 0.0021    // 1 KZT = 0.0021 USD

amountUSD := ConvertCurrency(amountKZT, "KZT", "USD", exchangeRate)
// amountUSD = 105 (1.05 USD)
```

### Example 5: Calculate Tax

```go
func CalculateTax(amountCents int64, taxRate float64) (subtotal, tax, total int64) {
    // Convert to decimal
    amount := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))
    rate := decimal.NewFromFloat(taxRate)

    // Calculate tax
    taxAmount := amount.Mul(rate).Round(2)
    totalAmount := amount.Add(taxAmount)

    // Convert back to cents
    subtotal = amountCents
    tax = taxAmount.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
    total = totalAmount.Mul(decimal.NewFromInt(100)).Round(0).IntPart()

    return
}

// Usage:
price := int64(10000) // 100.00 KZT
subtotal, tax, total := CalculateTax(price, 0.12) // 12% tax

// subtotal = 10000 (100.00 KZT)
// tax = 1200 (12.00 KZT)
// total = 11200 (112.00 KZT)
```

### Example 6: Convert Gateway Response

```go
// When external API returns float64
func ConvertGatewayAmount(gatewayAmount float64) int64 {
    // Use decimal to avoid precision errors
    dec := decimal.NewFromFloat(gatewayAmount)
    cents := dec.Mul(decimal.NewFromInt(100)).Round(0)
    return cents.IntPart()
}

// When calling external API (if it requires float64)
func PrepareGatewayAmount(amountCents int64) float64 {
    dec := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))
    amount, _ := dec.Float64()
    return amount
}

// Usage in refund service:
if isPartialRefund {
    amountDecimal := decimal.NewFromInt(refundAmount).Div(decimal.NewFromInt(100))
    amount, _ := amountDecimal.Float64()
    gatewayAmount = &amount
}
```

## Best Practices

### ✅ DO

```go
// ✅ Store money as int64 (cents/tenge)
type Payment struct {
    Amount int64 // Stored as 10050 for 100.50 KZT
}

// ✅ Use decimal for all calculations
refund := decimal.NewFromInt(originalAmount).Mul(percent)

// ✅ Round after calculations
rounded := calculation.Round(2)

// ✅ Always specify precision in Round/Truncate
amount.Round(2) // 2 decimal places for currency

// ✅ Use StringFixed for display
display := amount.StringFixed(2) // "100.50"

// ✅ Check for zero before division
if divisor.IsZero() {
    return errors.New("division by zero")
}
split := total.Div(divisor)

// ✅ Use Cmp/Equal for comparisons
if amount1.Equal(amount2) {
    // Amounts match
}
```

### ❌ DON'T

```go
// ❌ Never use float64 for money storage
type Payment struct {
    Amount float64 // WRONG!
}

// ❌ Don't use float arithmetic
total := price * 0.5 // WRONG! Use decimal.Mul

// ❌ Don't forget to round
display := amount.String() // Might show "100.5555555"

// ❌ Don't convert to float64 unless absolutely necessary
floatAmount := float64(amountCents) / 100 // WRONG! Use decimal

// ❌ Don't assume Float64() is exact
amount, _ := decimal.Float64() // Might lose precision!

// ❌ Don't use == for decimal comparison
if amount1 == amount2 { // Won't work with decimals
```

## Testing with Decimal

```go
func TestCalculateRefund(t *testing.T) {
    tests := []struct {
        name           string
        originalAmount int64
        refundPercent  float64
        expectedRefund int64
    }{
        {
            name:           "50% refund",
            originalAmount: 10000, // 100.00 KZT
            refundPercent:  0.5,
            expectedRefund: 5000, // 50.00 KZT
        },
        {
            name:           "33% refund with rounding",
            originalAmount: 10000,
            refundPercent:  0.33,
            expectedRefund: 3300, // 33.00 KZT
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := calculatePartialRefund(tt.originalAmount, tt.refundPercent)
            assert.Equal(t, tt.expectedRefund, result)
        })
    }
}
```

## Common Pitfalls

### Pitfall 1: Rounding Errors in Splits

```go
// ❌ WRONG - rounding errors accumulate
func SplitBad(total int64, parts int) []int64 {
    perPart := float64(total) / float64(parts)
    splits := make([]int64, parts)
    for i := 0; i < parts; i++ {
        splits[i] = int64(perPart) // Truncates!
    }
    return splits
}

// Total: 10000 (100.00 KZT), Parts: 3
// Result: [3333, 3333, 3333] = 9999 (MISSING 1 cent!)

// ✅ CORRECT - handle remainder
func SplitGood(total int64, parts int) []int64 {
    totalDec := decimal.NewFromInt(total)
    partsDec := decimal.NewFromInt(int64(parts))
    perPart := totalDec.Div(partsDec).Round(0)

    splits := make([]int64, parts)
    sum := decimal.Zero

    for i := 0; i < parts-1; i++ {
        splits[i] = perPart.IntPart()
        sum = sum.Add(perPart)
    }
    splits[parts-1] = totalDec.Sub(sum).IntPart()

    return splits
}
```

### Pitfall 2: Precision Loss in Float Conversion

```go
// ❌ WRONG - precision loss
amount := decimal.NewFromFloat(100.555)
floatVal, _ := amount.Float64()
back := decimal.NewFromFloat(floatVal)
// back might not equal amount!

// ✅ CORRECT - keep as decimal, only convert when necessary
amount := decimal.NewFromFloat(100.555)
// ... do all calculations in decimal ...
// Only convert to float64 when calling external API
if needFloat {
    floatVal, exact := amount.Float64()
    if !exact {
        log.Warn("precision loss in conversion")
    }
}
```

## Resources

- [shopspring/decimal Documentation](https://pkg.go.dev/github.com/shopspring/decimal)
- [GitHub Repository](https://github.com/shopspring/decimal)
- [Why Floats Are Bad for Money](https://stackoverflow.com/questions/3730019/why-not-use-double-or-float-to-represent-currency)

## Summary

1. **Always store money as int64** (smallest currency unit)
2. **Use decimal for calculations** (conversions, splits, percentages)
3. **Round appropriately** after calculations
4. **Convert to float64 only** when calling external APIs
5. **Test edge cases** (rounding, splits, conversions)

---

*Generated as part of the library payment system refactoring*
