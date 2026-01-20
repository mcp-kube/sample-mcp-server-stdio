package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func logMsg(prefix, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000000")
	fmt.Fprintf(os.Stderr, "[%s] %s %s\n", timestamp, prefix, message)
}

type WordCountArgs struct {
	Text string `json:"text" jsonschema:"The text to analyze"`
}

type FormatCurrencyArgs struct {
	Amount   float64 `json:"amount" jsonschema:"The numeric amount to format"`
	Currency string  `json:"currency" jsonschema:"Currency code (USD, EUR, GBP, JPY)"`
}

type SlugifyArgs struct {
	Text string `json:"text" jsonschema:"The text to convert to a URL-friendly slug"`
}

type RomanNumeralArgs struct {
	Number *int    `json:"number,omitempty" jsonschema:"Decimal number to convert to Roman (1-3999)"`
	Roman  *string `json:"roman,omitempty" jsonschema:"Roman numeral to convert to decimal"`
}

type TemperatureConvertArgs struct {
	Value    float64 `json:"value" jsonschema:"The temperature value to convert"`
	FromUnit string  `json:"from_unit" jsonschema:"Source temperature unit (celsius, fahrenheit, or kelvin)"`
	ToUnit   string  `json:"to_unit" jsonschema:"Target temperature unit (celsius, fahrenheit, or kelvin)"`
}

func handleWordCount(ctx context.Context, req *mcp.CallToolRequest, args WordCountArgs) (*mcp.CallToolResult, any, error) {
	logMsg("[TOOL]", fmt.Sprintf("word_count called with text length: %d", len(args.Text)))

	words := 0
	if len(strings.TrimSpace(args.Text)) > 0 {
		words = len(strings.Fields(args.Text))
	}

	lines := strings.Count(args.Text, "\n") + 1
	if args.Text == "" {
		lines = 0
	}

	chars := len(args.Text)
	charsNoSpaces := len(strings.ReplaceAll(strings.ReplaceAll(args.Text, " ", ""), "\n", ""))

	result := map[string]any{
		"words":                    words,
		"characters":               chars,
		"characters_no_whitespace": charsNoSpaces,
		"lines":                    lines,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Words: %d\nCharacters: %d\nCharacters (no whitespace): %d\nLines: %d",
					words, chars, charsNoSpaces, lines),
			},
		},
	}, result, nil
}

func handleFormatCurrency(ctx context.Context, req *mcp.CallToolRequest, args FormatCurrencyArgs) (*mcp.CallToolResult, any, error) {
	logMsg("[TOOL]", fmt.Sprintf("format_currency called: %.2f %s", args.Amount, args.Currency))

	var symbol string
	var decimals int

	switch args.Currency {
	case "USD":
		symbol = "$"
		decimals = 2
	case "EUR":
		symbol = "€"
		decimals = 2
	case "GBP":
		symbol = "£"
		decimals = 2
	case "JPY":
		symbol = "¥"
		decimals = 0
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Unsupported currency: %s", args.Currency),
				},
			},
			IsError: true,
		}, nil, nil
	}

	var formatted string
	if decimals == 0 {
		formatted = fmt.Sprintf("%s%.0f", symbol, args.Amount)
	} else {
		formatted = fmt.Sprintf("%s%.2f", symbol, args.Amount)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: formatted,
			},
		},
	}, map[string]any{"formatted": formatted}, nil
}

func handleSlugify(ctx context.Context, req *mcp.CallToolRequest, args SlugifyArgs) (*mcp.CallToolResult, any, error) {
	logMsg("[TOOL]", fmt.Sprintf("slugify called with text: %s", args.Text))

	slug := strings.ToLower(args.Text)
	slug = strings.TrimSpace(slug)

	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: slug,
			},
		},
	}, map[string]any{"slug": slug}, nil
}

func intToRoman(num int) string {
	if num < 1 || num > 3999 {
		return ""
	}

	values := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}

	var result strings.Builder
	for i := 0; i < len(values); i++ {
		for num >= values[i] {
			result.WriteString(symbols[i])
			num -= values[i]
		}
	}

	return result.String()
}

func romanToInt(s string) (int, error) {
	s = strings.ToUpper(s)
	romanMap := map[rune]int{
		'I': 1,
		'V': 5,
		'X': 10,
		'L': 50,
		'C': 100,
		'D': 500,
		'M': 1000,
	}

	result := 0
	prevValue := 0

	for i := len(s) - 1; i >= 0; i-- {
		char := rune(s[i])
		value, ok := romanMap[char]
		if !ok {
			return 0, fmt.Errorf("invalid Roman numeral character: %c", char)
		}

		if value < prevValue {
			result -= value
		} else {
			result += value
		}
		prevValue = value
	}

	return result, nil
}

func handleRomanNumeral(ctx context.Context, req *mcp.CallToolRequest, args RomanNumeralArgs) (*mcp.CallToolResult, any, error) {
	logMsg("[TOOL]", "roman_numeral called")

	if args.Number != nil && args.Roman != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Please provide either 'number' or 'roman', not both",
				},
			},
			IsError: true,
		}, nil, nil
	}

	if args.Number == nil && args.Roman == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Please provide either 'number' or 'roman'",
				},
			},
			IsError: true,
		}, nil, nil
	}

	if args.Number != nil {
		num := *args.Number
		if num < 1 || num > 3999 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Number must be between 1 and 3999",
					},
				},
				IsError: true,
			}, nil, nil
		}

		roman := intToRoman(num)
		logMsg("[TOOL]", fmt.Sprintf("Converted %d to %s", num, roman))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: roman,
				},
			},
		}, map[string]any{"roman": roman}, nil
	}

	decimal, err := romanToInt(*args.Roman)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: err.Error(),
				},
			},
			IsError: true,
		}, nil, nil
	}

	logMsg("[TOOL]", fmt.Sprintf("Converted %s to %d", *args.Roman, decimal))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%d", decimal),
			},
		},
	}, map[string]any{"decimal": decimal}, nil
}

func handleTemperatureConvert(ctx context.Context, req *mcp.CallToolRequest, args TemperatureConvertArgs) (*mcp.CallToolResult, any, error) {
	logMsg("[TOOL]", fmt.Sprintf("temperature_convert called: %.2f %s to %s", args.Value, args.FromUnit, args.ToUnit))

	if args.FromUnit == args.ToUnit {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("%.2f", args.Value),
				},
			},
		}, map[string]any{"result": args.Value}, nil
	}

	toCelsius := func(value float64, unit string) (float64, error) {
		switch unit {
		case "celsius":
			return value, nil
		case "fahrenheit":
			return (value - 32) * 5 / 9, nil
		case "kelvin":
			return value - 273.15, nil
		default:
			return 0, fmt.Errorf("unknown unit: %s", unit)
		}
	}

	fromCelsius := func(value float64, unit string) (float64, error) {
		switch unit {
		case "celsius":
			return value, nil
		case "fahrenheit":
			return value*9/5 + 32, nil
		case "kelvin":
			return value + 273.15, nil
		default:
			return 0, fmt.Errorf("unknown unit: %s", unit)
		}
	}

	celsius, err := toCelsius(args.Value, args.FromUnit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: err.Error(),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result, err := fromCelsius(celsius, args.ToUnit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: err.Error(),
				},
			},
			IsError: true,
		}, nil, nil
	}

	if math.IsNaN(result) || math.IsInf(result, 0) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Temperature conversion resulted in invalid value",
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%.2f", result),
			},
		},
	}, map[string]any{"result": result}, nil
}

func main() {
	logMsg("[MAIN]", "Starting stdio MCP server")

	impl := &mcp.Implementation{
		Name:    "sample-mcp-server-stdio",
		Version: "1.0.0",
	}

	server := mcp.NewServer(impl, nil)

	logMsg("[MAIN]", "Registering tools")

	mcp.AddTool(server, &mcp.Tool{
		Name:        "word_count",
		Description: "Analyze text and count words, characters, and lines",
	}, handleWordCount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "format_currency",
		Description: "Format a number as currency with proper symbol and decimal places",
	}, handleFormatCurrency)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "slugify",
		Description: "Convert text to a URL-friendly slug (lowercase, hyphens, no special characters)",
	}, handleSlugify)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "roman_numeral",
		Description: "Convert between decimal numbers (1-3999) and Roman numerals",
	}, handleRomanNumeral)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "temperature_convert",
		Description: "Convert temperatures between Celsius, Fahrenheit, and Kelvin",
	}, handleTemperatureConvert)

	logMsg("[MAIN]", "Starting server on stdio")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil && err != io.EOF {
		log.Fatalf("[ERROR] Server error: %v", err)
	}

	logMsg("[MAIN]", "Server stopped")
}
