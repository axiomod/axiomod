package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// greeter is a small test interface for interface-binding helpers.
type greeter interface {
	Greet() string
}

type englishGreeter struct{}

func (g *englishGreeter) Greet() string { return "hello" }

func newEnglishGreeter() *englishGreeter { return &englishGreeter{} }

func TestNewModule(t *testing.T) {
	m := NewModule("test-module")
	require.NotNil(t, m)
	assert.Equal(t, "test-module", m.Name)
	assert.Empty(t, m.Providers)
	assert.Empty(t, m.Invocations)
	assert.Empty(t, m.Options)
}

func TestModule_BuilderChaining(t *testing.T) {
	m := NewModule("test-module")

	returned := m.Provide(newEnglishGreeter).Invoke(func(g *englishGreeter) {}).Option(fx.Supply("extra"))

	assert.Same(t, m, returned, "builder methods must return the module for chaining")
	assert.Len(t, m.Providers, 1)
	assert.Len(t, m.Invocations, 1)
	assert.Len(t, m.Options, 1)
}

func TestModule_Build_Empty(t *testing.T) {
	m := NewModule("empty")

	app := fxtest.New(t, m.Build())
	app.RequireStart().RequireStop()
}

func TestModule_Build_ProvidesAndInvokes(t *testing.T) {
	var invoked bool
	var received *englishGreeter

	m := NewModule("full").
		Provide(newEnglishGreeter).
		Invoke(func(g *englishGreeter) {
			invoked = true
			received = g
		}).
		Option(fx.Supply("supplied-value"))

	var suppliedValue string
	app := fxtest.New(t,
		m.Build(),
		fx.Invoke(func(s string) { suppliedValue = s }),
	)
	app.RequireStart().RequireStop()

	assert.True(t, invoked)
	assert.NotNil(t, received)
	assert.Equal(t, "supplied-value", suppliedValue)
}

func TestProvideAs(t *testing.T) {
	var received greeter

	app := fxtest.New(t,
		fx.Provide(ProvideAs(newEnglishGreeter, new(greeter))),
		fx.Invoke(func(g greeter) { received = g }),
	)
	app.RequireStart().RequireStop()

	require.NotNil(t, received)
	assert.Equal(t, "hello", received.Greet())
}

func TestProvideNamed(t *testing.T) {
	var received *englishGreeter

	app := fxtest.New(t,
		fx.Provide(ProvideNamed(newEnglishGreeter, "primary")),
		fx.Invoke(fx.Annotate(
			func(g *englishGreeter) { received = g },
			fx.ParamTags(`name:"primary"`),
		)),
	)
	app.RequireStart().RequireStop()

	require.NotNil(t, received)
	assert.Equal(t, "hello", received.Greet())
}

func TestProvideNamedAs(t *testing.T) {
	var received greeter

	app := fxtest.New(t,
		fx.Provide(ProvideNamedAs(newEnglishGreeter, "primary", new(greeter))),
		fx.Invoke(fx.Annotate(
			func(g greeter) { received = g },
			fx.ParamTags(`name:"primary"`),
		)),
	)
	app.RequireStart().RequireStop()

	require.NotNil(t, received)
	assert.Equal(t, "hello", received.Greet())
}

func TestExtractType(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"value type", englishGreeter{}},
		{"pointer type", &englishGreeter{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractType(tt.input)
			assert.IsType(t, &englishGreeter{}, got, "ExtractType must return a pointer to the element type")
		})
	}
}

func TestPrintDependencyGraph(t *testing.T) {
	app := fx.New(fx.NopLogger)
	assert.NotPanics(t, func() {
		PrintDependencyGraph(app)
	})
}
