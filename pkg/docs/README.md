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

### Builder pour multi-service

Le Builder permet d'agréger plusieurs spécifications en une seule documentation HTML multi-service :

```go
// Charger des specs depuis différentes sources
apiSpec, _ := docs.LoadSpec("api-gateway.yaml")
userSpec, _ := docs.LoadSpecFromURL("https://user-service/metrics.yaml")
orderData := []byte(`...`) // YAML bytes
orderSpec, _ := docs.LoadSpecFromBytes(orderData)

// Construire le HTML multi-service
builder := docs.NewHTMLBuilder("My Platform", "1.0.0")
builder.SetDescription("Platform-wide metrics documentation")
builder.AddFromSpec(apiSpec)    // Merge tous les services de apiSpec
builder.AddFromSpec(userSpec)   // Merge tous les services de userSpec
builder.AddFromSpec(orderSpec)  // Merge tous les services de orderSpec

// Générer le HTML final
html, err := builder.BuildHTML()
if err != nil {
    log.Fatal(err)
}

// Ou écrire directement dans un fichier
err = builder.BuildHTMLFile("platform-docs.html")
```

Le Builder supporte le chaînage de méthodes :

```go
html, err := docs.NewHTMLBuilder("Platform", "1.0.0").
    SetDescription("All services").
    AddFromSpec(spec1).
    AddFromSpec(spec2).
    BuildHTML()
```

### Charger depuis différentes sources

```go
// Depuis un fichier local
spec1, err := docs.LoadSpec("metrics.yaml")

// Depuis une URL HTTP
spec2, err := docs.LoadSpecFromURL("https://api.example.com/metrics.yaml")

// Depuis des bytes YAML
yamlData := []byte(`...`)
spec3, err := docs.LoadSpecFromBytes(yamlData)

// Depuis un io.Reader (ex: HTTP response body)
resp, _ := http.Get("https://...")
spec4, err := docs.LoadSpecFromReader(resp.Body)
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

### Chargement de spécifications

#### `LoadSpec(path string) (*Specification, error)`
Charge et valide une spécification depuis un fichier YAML local.

#### `LoadSpecFromBytes(data []byte) (*Specification, error)`
Charge et valide une spécification depuis des bytes YAML.

#### `LoadSpecFromURL(url string) (*Specification, error)`
Charge et valide une spécification depuis une URL HTTP. L'URL doit retourner du contenu YAML.

#### `LoadSpecFromReader(r io.Reader) (*Specification, error)`
Charge et valide une spécification depuis un io.Reader (ex: http.Response.Body).

### Génération HTML simple

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

### Builder (Multi-service)

#### `NewHTMLBuilder(title, version string) *HTMLBuilder`
Crée un nouveau builder pour construire une documentation multi-service.

#### `(*HTMLBuilder) SetDescription(description string) Builder`
Définit la description de la documentation. Retourne le builder pour permettre le chaînage.

#### `(*HTMLBuilder) AddFromSpec(spec *Specification) Builder`
Fusionne tous les services d'une spécification dans le builder. Si un service avec le même nom existe déjà, il sera écrasé. Retourne le builder pour permettre le chaînage.

#### `(*HTMLBuilder) AddService(name string, service Service) Builder`
Ajoute un service individuel avec le nom spécifié. Retourne le builder pour permettre le chaînage.

#### `(*HTMLBuilder) BuildHTML() ([]byte, error)`
Génère le HTML final depuis tous les services agrégés. Valide la spécification avant la génération.

#### `(*HTMLBuilder) BuildHTMLFile(outputPath string) error`
Génère le HTML et l'écrit dans un fichier.

#### `(*HTMLBuilder) GetSpecification() *Specification`
Retourne la spécification sous-jacente en cours de construction.

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

Les spécifications sont organisées autour de `services`. Un service unique affichera la documentation directement, plusieurs services afficheront un sélecteur.

### Exemple avec un service

```yaml
version: "1.0"
info:
  title: "My Application"
  version: "1.0.0"
services:
  default:
    info:
      title: "My Service"
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

### Exemple avec plusieurs services

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
      package: "ApiGateway.Metrics"
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
      package: "UserService.Metrics"
    metrics:
      users_total:
        namespace: app
        type: counter
        help: "Total users"
```

## Licence

Voir le fichier LICENSE à la racine du projet.
