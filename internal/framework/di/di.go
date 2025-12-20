package di

import (
	"fmt"
	"reflect"

	"go.uber.org/fx"
)

// Module represents a collection of providers and invocations
type Module struct {
	// Name is the name of the module
	Name string

	// Providers are the providers for the module
	Providers []interface{}

	// Invocations are the invocations for the module
	Invocations []interface{}

	// Options are additional fx options
	Options []fx.Option
}

// NewModule creates a new module
func NewModule(name string) *Module {
	return &Module{
		Name:        name,
		Providers:   []interface{}{},
		Invocations: []interface{}{},
		Options:     []fx.Option{},
	}
}

// Provide adds providers to the module
func (m *Module) Provide(providers ...interface{}) *Module {
	m.Providers = append(m.Providers, providers...)
	return m
}

// Invoke adds invocations to the module
func (m *Module) Invoke(invocations ...interface{}) *Module {
	m.Invocations = append(m.Invocations, invocations...)
	return m
}

// Option adds fx options to the module
func (m *Module) Option(options ...fx.Option) *Module {
	m.Options = append(m.Options, options...)
	return m
}

// Build builds the module into an fx.Option
func (m *Module) Build() fx.Option {
	var options []fx.Option

	// Add providers
	if len(m.Providers) > 0 {
		options = append(options, fx.Provide(m.Providers...))
	}

	// Add invocations
	if len(m.Invocations) > 0 {
		options = append(options, fx.Invoke(m.Invocations...))
	}

	// Add additional options
	options = append(options, m.Options...)

	// Return combined options
	return fx.Options(options...)
}

// ProvideAs provides a value as a specific interface
func ProvideAs(val interface{}, as interface{}) interface{} {
	return fx.Annotate(
		val,
		fx.As(as),
	)
}

// ProvideNamed provides a value with a name
func ProvideNamed(val interface{}, name string) interface{} {
	return fx.Annotate(
		val,
		fx.ResultTags(`name:"`+name+`"`),
	)
}

// ProvideNamedAs provides a value with a name as a specific interface
func ProvideNamedAs(val interface{}, name string, as interface{}) interface{} {
	return fx.Annotate(
		val,
		fx.ResultTags(`name:"`+name+`"`),
		fx.As(as),
	)
}

// ExtractType extracts the type of a value for use with fx.As
func ExtractType(val interface{}) interface{} {
	t := reflect.TypeOf(val)

	// If val is a pointer, get the element type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Create a new pointer to the type
	ptr := reflect.New(t).Interface()

	return ptr
}

// PrintDependencyGraph prints the dependency graph of an fx.App
func PrintDependencyGraph(app *fx.App) {
	fmt.Println("Dependency Graph:")
	fmt.Printf("%+v\n", app)
}
