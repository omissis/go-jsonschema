package tests_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	testInField "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/inField"
	testNumString "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/numString"
	testTemporal "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/temporalVariant"
	testWithDefault "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/withDefault"
)

// TestOneOfPrimitiveMarshalByValue marshals a parent holding a required wrapper
// by value. With pointer-receiver Marshal methods, encoding/json could not find
// MarshalJSON on the unaddressable field and yaml.v3 never looks for one there,
// so both emitted the wrapper's unexported fields as `{}`.
func TestOneOfPrimitiveMarshalByValue(t *testing.T) {
	t.Parallel()

	var v testInField.InField

	require.NoError(t, json.Unmarshal([]byte(`{"name": "a", "value": "x"}`), &v))

	out, err := json.Marshal(v)
	require.NoError(t, err)
	require.JSONEq(t, `{"name": "a", "value": "x"}`, string(out))

	y, err := yamlv3.Marshal(v)
	require.NoError(t, err)

	var back testInField.InField

	require.NoError(t, yamlv3.Unmarshal(y, &back), "yaml: %s", y)

	got, ok := back.Value.AsString()
	require.True(t, ok, "yaml: %s", y)
	require.Equal(t, "x", got)
}

// TestOneOfPrimitiveYAMLUnquotedTimestamp decodes unquoted YAML dates, which
// yaml.v3 tags `!!timestamp` rather than `!!str`. They are strings in the
// schema's terms, for a temporal variant and for a plain one alike.
func TestOneOfPrimitiveYAMLUnquotedTimestamp(t *testing.T) {
	t.Parallel()

	var tv testTemporal.TemporalVariant

	require.NoError(t, yamlv3.Unmarshal([]byte("timeOrNumber: 2024-01-02T03:04:05Z\n"), &tv))

	when, ok := tv.TimeOrNumber.AsDateTime()
	require.True(t, ok)
	require.True(t, when.Equal(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)), "got %v", when)

	var ns testNumString.NumString

	require.NoError(t, yamlv3.Unmarshal([]byte("value: 2024-01-02\n"), &ns))

	s, ok := ns.Value.AsString()
	require.True(t, ok)
	require.Equal(t, "2024-01-02", s)
}

// TestOneOfPrimitiveDateAndTimeVariants decodes `date` and `time` variants into
// the types an ordinary field of those formats gets. time.Time parses only
// RFC 3339 date-times, so it rejected "2024-01-02" and "03:04:05Z" and took a
// full date-time for a date.
func TestOneOfPrimitiveDateAndTimeVariants(t *testing.T) {
	t.Parallel()

	var v testTemporal.TemporalVariant

	require.NoError(t, json.Unmarshal([]byte(`{"day": "2024-01-02", "clock": "03:04:05Z"}`), &v))

	day, ok := v.Day.AsDate()
	require.True(t, ok)
	require.Equal(t, "2024-01-02", day.Format(time.DateOnly))

	_, ok = v.Clock.AsTime()
	require.True(t, ok)

	out, err := json.Marshal(v)
	require.NoError(t, err)
	require.JSONEq(t, `{"day": "2024-01-02", "clock": "03:04:05"}`, string(out))

	var wrong testTemporal.TemporalVariant

	require.Error(t, json.Unmarshal([]byte(`{"day": "2024-01-02T00:00:00Z"}`), &wrong),
		"a date variant must not take a date-time")

	// YAML goes through the same types: an unquoted date is a !!timestamp.
	var y testTemporal.TemporalVariant

	require.NoError(t, yamlv3.Unmarshal([]byte("day: 2024-01-02\nclock: \"03:04:05Z\"\n"), &y))

	text, err := yamlv3.Marshal(y)
	require.NoError(t, err)

	var back testTemporal.TemporalVariant

	require.NoError(t, yamlv3.Unmarshal(text, &back), "yaml: %s", text)

	backDay, ok := back.Day.AsDate()
	require.True(t, ok, "yaml: %s", text)
	require.Equal(t, "2024-01-02", backDay.Format(time.DateOnly))
}

// TestOneOfPrimitiveDefault covers defaults on primitive union wrappers. The
// default is decoded into the wrapper, so it reads back exactly as the same
// value sent on the wire would, through JSON and YAML alike, and a value that
// is present replaces it.
func TestOneOfPrimitiveDefault(t *testing.T) {
	t.Parallel()

	assertDefaults := func(t *testing.T, v testWithDefault.WithDefault) {
		t.Helper()

		debug, ok := v.Debug.AsBool()
		assert.True(t, ok)
		assert.False(t, debug)

		tab, ok := v.Tab.AsNumber()
		assert.True(t, ok)
		assert.InDelta(t, 1.0, tab, 0)

		label, ok := v.Label.AsString()
		assert.True(t, ok)
		assert.Equal(t, "auto", label)

		nullable, ok := v.Nullable.AsString()
		assert.True(t, ok)
		assert.Equal(t, "x", nullable)

		assert.True(t, v.Mismatch.IsZero(), "a default the union does not accept is left out")
	}

	t.Run("json", func(t *testing.T) {
		t.Parallel()

		var v testWithDefault.WithDefault

		require.NoError(t, json.Unmarshal([]byte(`{}`), &v))
		assertDefaults(t, v)
	})

	t.Run("yaml", func(t *testing.T) {
		t.Parallel()

		var v testWithDefault.WithDefault

		require.NoError(t, yamlv3.Unmarshal([]byte(`{}`), &v))
		assertDefaults(t, v)
	})

	t.Run("an explicit null stays null where the union accepts it", func(t *testing.T) {
		t.Parallel()

		var v testWithDefault.WithDefault

		require.NoError(t, json.Unmarshal([]byte(`{"nullable":null}`), &v))
		assert.True(t, v.Nullable.IsNull())

		require.NoError(t, yamlv3.Unmarshal([]byte(`nullable: null`), &v))
		assert.True(t, v.Nullable.IsNull())
	})

	t.Run("a present value replaces the default", func(t *testing.T) {
		t.Parallel()

		var v testWithDefault.WithDefault

		require.NoError(t, json.Unmarshal([]byte(`{"debug":"verbose"}`), &v))

		debug, ok := v.Debug.AsString()
		assert.True(t, ok)
		assert.Equal(t, "verbose", debug)
	})
}
