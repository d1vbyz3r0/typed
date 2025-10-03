package format

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/common/typing"
	"log/slog"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var splitParamsRegex = regexp.MustCompile(`'[^']*'|\S+`)

type TagFn func(ctx *FieldContext)

type FieldContext struct {
	Type         reflect.Type
	Required     bool
	Nullable     bool
	Not          []any
	Format       string
	Min          *float64
	ExclusiveMin bool
	Max          *float64
	ExclusiveMax bool
	MinItems     uint64
	MaxItems     uint64
	MinLength    uint64
	MaxLength    uint64
	OneOf        []any

	// child will handle child elems validation context, created by "dive" tag.
	// Each child will have child recursively, until dive tags found
	child *FieldContext
	// shouldOmit is a special case for omitnil and omitempty tags
	shouldOmit      bool
	pattern         string
	patterns        []string
	validationRules []string
	hasOrConditions bool
}

// NewFieldContext creates FieldContext with all required children.
// If no valid "validate" tag found, or tag marked as "-" it returns nil
func NewFieldContext(t reflect.Type, tag reflect.StructTag) *FieldContext {
	tagVal := tag.Get("validate")
	if tagVal == "" || tagVal == "-" {
		return nil
	}

	rules := strings.Split(tagVal, ",")
	ctx := buildContext(typing.DerefReflectPtr(t), rules)
	ctx.finalize()
	return ctx
}

func buildContext(t reflect.Type, rules []string) *FieldContext {
	segments := splitByDive(rules)
	var (
		root *FieldContext
		curr *FieldContext
	)

	for _, seg := range segments {
		segRules := stripKeysBlock(filterEmpty(seg))
		shouldOmit := false
		segRules = slices.DeleteFunc(segRules, func(s string) bool {
			if s == "omitempty" || s == "omitnil" {
				shouldOmit = true
				return true
			}
			return false
		})

		isNullable := t.Kind() == reflect.Slice ||
			t.Kind() == reflect.Map ||
			t.Kind() == reflect.Pointer

		node := &FieldContext{
			Type:            t,
			Nullable:        isNullable,
			validationRules: segRules,
			shouldOmit:      shouldOmit,
			hasOrConditions: strings.Contains(strings.Join(segRules, ","), "|"),
		}

		for _, rule := range node.tagNames() {
			applyFormat, ok := Formats[rule]
			if !ok {
				slog.Warn("validation rule not found, skipping", "rule", rule)
				continue
			}

			applyFormat(node)
		}

		if root == nil {
			root = node
		} else {
			curr.child = node
		}
		curr = node

		t = typing.DerefReflectPtr(t)
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
			t = t.Elem()
		}
	}

	return root
}

func (c *FieldContext) finalize() {
	// if slices.Contains(c.validationRules, "omitempty") || slices.Contains(c.validationRules, "omitnil") {
	// 	c.shouldOmit = true
	// }

	// if c.pattern != "" {
	//	c.pattern += "$"
	// }

	if len(c.patterns) > 0 {
		p := strings.Join(c.patterns, "|")
		if len(c.patterns) > 1 {
			c.pattern = fmt.Sprintf("^(%s)$", p)
		} else {
			c.pattern = fmt.Sprintf("^%s$", p)
		}
	}

	if c.hasOrConditions && c.pattern != "" && c.Format != "" {
		c.Format = ""
	}

	if IsKnown(c.Format) && c.pattern != "" {
		c.pattern = ""
	}

	for child := c.child; child != nil; child = child.child {
		child.finalize()
	}
}

func (c *FieldContext) AddPattern(pattern string) {
	c.patterns = append(c.patterns, fmt.Sprintf("(%s)", pattern))
}

func (c *FieldContext) LookupString(key string) (string, error) {
	val := ""
	for _, rule := range c.validationRules {
		if !strings.Contains(rule, key) {
			continue
		}

		parts := strings.Split(rule, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid format rule: '%s'", rule)
		}

		if parts[0] != key {
			continue
		}

		val = parts[1]
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

func (c *FieldContext) LookupUint(key string) (uint64, error) {
	v, err := c.LookupString(key)
	if err != nil {
		return 0, fmt.Errorf("lookup string: %w", err)
	}

	return strconv.ParseUint(v, 10, 64)
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

func (c *FieldContext) tagNames() []string {
	names := make([]string, 0, len(c.validationRules))
	for _, rule := range c.validationRules {
		if strings.Contains(rule, "|") {
			parts := strings.Split(rule, "|")
			names = append(names, parts...)
		} else {
			parts := strings.Split(rule, "=")
			names = append(names, parts[0])
		}
	}

	return names
}
