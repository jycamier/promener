package generator

// EnvTransformer is a function that transforms an EnvVarValue to language-specific code
type EnvTransformer func(EnvVarValue) string

// GoEnvTransformer generates Go code for environment variables
func GoEnvTransformer(e EnvVarValue) string {
	if !e.IsEnvVar {
		return `"` + e.LiteralValue + `"`
	}

	if e.DefaultValue != "" {
		return `getEnvOrDefault("` + e.EnvVar + `", "` + e.DefaultValue + `")`
	}

	return `os.Getenv("` + e.EnvVar + `")`
}

// DotNetEnvTransformer generates C# code for environment variables
func DotNetEnvTransformer(e EnvVarValue) string {
	if !e.IsEnvVar {
		return `"` + e.LiteralValue + `"`
	}

	if e.DefaultValue != "" {
		return `Environment.GetEnvironmentVariable("` + e.EnvVar + `") ?? "` + e.DefaultValue + `"`
	}

	return `Environment.GetEnvironmentVariable("` + e.EnvVar + `") ?? throw new InvalidOperationException("Environment variable ` + e.EnvVar + ` is required")`
}

// NodeJSEnvTransformer generates TypeScript code for environment variables
func NodeJSEnvTransformer(e EnvVarValue) string {
	if !e.IsEnvVar {
		return `'` + e.LiteralValue + `'`
	}

	if e.DefaultValue != "" {
		return `process.env.` + e.EnvVar + ` || '` + e.DefaultValue + `'`
	}

	return `process.env.` + e.EnvVar + ` || (() => { throw new Error('Environment variable ` + e.EnvVar + ` is required'); })()`
}
