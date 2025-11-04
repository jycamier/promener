package generator

//go:generate mockgen -source=builder.go -destination=mocks/mock_builder.go -package=mocks TemplateDataBuilder

// Ensure the concrete builder types implement the interface.
var (
	_ TemplateDataBuilder = (*GoTemplateDataBuilder)(nil)
	_ TemplateDataBuilder = (*DotNetTemplateDataBuilder)(nil)
	_ TemplateDataBuilder = (*NodeJSTemplateDataBuilder)(nil)
)
