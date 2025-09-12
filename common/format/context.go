package format

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var splitParamsRegex = regexp.MustCompile(`'[^']*'|\S+`)

type TagFn func(ctx *FieldContext)

type FieldContext struct {
	Type            reflect.Type
	Tag             reflect.StructTag
	ValidationRules []string
	Required        bool
	// ShouldOmit is a special case for omitnil and omitempty tags
	ShouldOmit      bool
	Pattern         string
	Format          string
	Min             *float64
	ExclusiveMin    bool
	Max             *float64
	ExclusiveMax    bool
	HasOrConditions bool
	OneOf           []any

	shouldDive bool
}

func NewFieldContext(t reflect.Type, tag reflect.StructTag) *FieldContext {
	tagVal := tag.Get("validate")
	if tagVal == "" || tagVal == "-" {
		return nil
	}

	rules := strings.Split(tagVal, ",")
	return &FieldContext{
		Type:            t,
		Tag:             tag,
		ValidationRules: rules,
		HasOrConditions: strings.Contains(tagVal, "|"),
		shouldDive:      strings.Contains(tagVal, "dive"),
	}
}

func (c *FieldContext) finalize() {
	v := c.Tag.Get("validate")
	if v == "" || v == "-" {
		return
	}

	if slices.Contains(c.ValidationRules, "omitempty") || slices.Contains(c.ValidationRules, "omitnil") {
		c.ShouldOmit = true
	}

	c.Pattern += "$"

	if c.HasOrConditions && c.Pattern != "" && c.Format != "" {
		c.Format = ""
	}

	if IsKnown(c.Format) && c.Pattern != "" {
		c.Pattern = ""
	}
}

// AddPattern ensures that validator tag contains "OR" condition and extend existing pattern or writes a new one
func (c *FieldContext) AddPattern(pattern string) {
	if c.Pattern == "" {
		c.Pattern = "^"
	}

	if !c.HasOrConditions {
		c.Pattern += pattern
		return
	}

	prefix := ""
	if c.Pattern != "^" {
		prefix = "|"
	}

	pattern = fmt.Sprintf("%s(%s)", prefix, pattern)
	c.Pattern += pattern
}

func (c *FieldContext) LookupString(key string) (string, error) {
	val := ""
	for _, rule := range c.ValidationRules {
		if !strings.Contains(rule, key) {
			continue
		}

		lt := strings.Split(rule, "=")
		if len(lt) != 2 {
			return "", fmt.Errorf("invalid format rule: '%s'", rule)
		}

		val = lt[1]
	}

	return val, nil
}

func (c *FieldContext) LookupFloat(key string) (float64, error) {
	v, err := c.LookupString(key)
	if err != nil {
		return 0, fmt.Errorf("lookup string: %w", err)
	}

	return strconv.ParseFloat(v, 64)
}

func (c *FieldContext) LookupStringSlice(key string) ([]string, error) {
	s, err := c.LookupString(key)
	if err != nil {
		return nil, fmt.Errorf("lookup string: %w", err)
	}

	items := splitParamsRegex.FindAllString(s, -1)
	for i := 0; i < len(items); i++ {
		items[i] = strings.ReplaceAll(items[i], "'", "")
	}

	return items, nil
}

func (c *FieldContext) LookupFloatSlice(key string) ([]float64, error) {
	s, err := c.LookupString(key)
	if err != nil {
		return nil, fmt.Errorf("lookup string: %w", err)
	}

	items := strings.Split(s, " ")
	result := make([]float64, 0, len(items))

	for _, v := range items {
		res, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("parse float: %w", err)
		}
		result = append(result, res)
	}

	return result, nil
}

func (c *FieldContext) TagNames() []string {
	names := make([]string, 0, len(c.ValidationRules))
	for _, rule := range c.ValidationRules {
		parts := strings.Split(rule, "=")
		names = append(names, parts[0])
	}

	return names
}
