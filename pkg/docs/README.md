# pkg/docs

Package `docs` fournit une API publique pour générer de la documentation HTML à partir de spécifications de métriques Prometheus, sans passer par le CLI.

## Installation

```bash
go get github.com/jycamier/promener/pkg/docs
```

## Usage rapide

### Cas le plus simple

```go
import "github.com/jycamier/promener/pkg/docs"

// Générer HTML depuis un fichier YAML
err := docs.GenerateHTMLFromFile("metrics.yaml", "output.html")
if err != nil {
    log.Fatal(err)
}
```

### Charger et générer séparément

```go
// Charger la spécification
spec, err := docs.LoadSpec("metrics.yaml")
if err != nil {
    log.Fatal(err)
}

// Inspecter la spec
fmt.Printf("Title: %s\n", spec.Info.Title)
fmt.Printf("Metrics: %d\n", len(spec.Metrics))

// Générer le HTML
html, err := docs.GenerateHTML(spec)
if err != nil {
    log.Fatal(err)
}

// Utiliser le HTML (écrire dans un fichier, servir via HTTP, etc.)
os.WriteFile("output.html", html, 0644)
```

### Travailler avec des bytes

```go
// Depuis des bytes YAML
yamlData := []byte(`
version: "1.0"
info:
  title: "My Metrics"
  version: "1.0.0"
  package: "metrics"
metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP requests"
`)

html, err := docs.GenerateHTMLFromBytes(yamlData)
if err != nil {
    log.Fatal(err)
}
```

### Réutiliser le générateur

Si vous générez plusieurs documents, réutilisez le générateur pour de meilleures performances :

```go
generator, err := docs.NewHTMLGenerator()
if err != nil {
    log.Fatal(err)
}

for _, specFile := range specFiles {
    spec, err := docs.LoadSpec(specFile)
    if err != nil {
        continue
    }

    generator.GenerateFile(spec, "output.html")
}
```

### Servir via HTTP

```go
http.HandleFunc("/metrics-docs", func(w http.ResponseWriter, r *http.Request) {
    html, err := docs.GenerateHTMLFromFile("metrics.yaml")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(html)
})

http.ListenAndServe(":8080", nil)
```

## API Reference

### Fonctions principales

#### `LoadSpec(path string) (*Specification, error)`
Charge et valide une spécification depuis un fichier YAML.

#### `LoadSpecFromBytes(data []byte) (*Specification, error)`
Charge et valide une spécification depuis des bytes YAML.

#### `GenerateHTML(spec *Specification) ([]byte, error)`
Génère la documentation HTML depuis une spécification. Retourne les bytes HTML.

#### `GenerateHTMLFile(spec *Specification, outputPath string) error`
Génère la documentation HTML et l'écrit dans un fichier.

#### `GenerateHTMLFromFile(specPath, outputPath string) error`
Fonction de commodité : charge un YAML et génère un HTML en une seule étape.

#### `GenerateHTMLFromFileToBytes(specPath string) ([]byte, error)`
Fonction de commodité : charge un YAML et retourne le HTML en bytes.

#### `GenerateHTMLFromBytes(yamlData []byte) ([]byte, error)`
Fonction de commodité : charge depuis des bytes YAML et retourne le HTML.

### Type HTMLGenerator

#### `NewHTMLGenerator() (*HTMLGenerator, error)`
Crée un nouveau générateur HTML. Utilisez-le si vous voulez générer plusieurs documents pour de meilleures performances.

#### `(*HTMLGenerator) Generate(spec *Specification) ([]byte, error)`
Génère le HTML depuis une spécification.

#### `(*HTMLGenerator) GenerateFile(spec *Specification, outputPath string) error`
Génère le HTML et l'écrit dans un fichier.

## Exemple complet

Voir [examples/using-pkg-docs/main.go](../../examples/using-pkg-docs/main.go) pour un exemple complet montrant :
- Génération simple
- Chargement et inspection de spec
- Génération depuis des bytes
- Réutilisation du générateur
- Serveur HTTP dynamique

Pour exécuter l'exemple :

```bash
cd examples/using-pkg-docs
go run main.go
```

## Format de spécification

Le package supporte deux modes :

### Mode mono-service

```yaml
version: "1.0"
info:
  title: "My Application"
  version: "1.0.0"
  package: "metrics"
servers:
  - url: "https://api.example.com"
    description: "Production"
metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP requests"
```

### Mode multi-services

```yaml
version: "1.0"
info:
  title: "My Platform"
  version: "1.0.0"
services:
  api-gateway:
    info:
      title: "API Gateway"
      version: "1.0.0"
    servers:
      - url: "https://api.example.com"
    metrics:
      requests_total:
        namespace: http
        type: counter
        help: "Total requests"
  user-service:
    info:
      title: "User Service"
      version: "1.0.0"
    metrics:
      # ...
```

## Licence

Voir le fichier LICENSE à la racine du projet.
