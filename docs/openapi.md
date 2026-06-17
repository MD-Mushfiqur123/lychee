# Lychee API Reference

The full OpenAPI 3.0 specification is available at [`openapi.yaml`](./openapi.yaml).

## Quick Links

- **Server:** `http://localhost:11434` (default)
- **Spec version:** OpenAPI 3.0.3
- **API version:** 1.0.0

## Viewing the Spec

You can view and interact with the spec using any OpenAPI-compatible tool:

- [Swagger Editor](https://editor.swagger.io/) — paste the YAML or load from URL
- [Redoc](https://redocly.github.io/redoc/) — auto-generated docs viewer
- VS Code [OpenAPI extension](https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi)

## Endpoints

| Method   | Path               | Description              |
|----------|--------------------|--------------------------|
| `GET`    | `/`                | Health check / Dashboard |
| `GET`    | `/api/version`     | Server version           |
| `GET`    | `/api/tags`        | List local models        |
| `GET`    | `/api/ps`          | List running models      |
| `POST`   | `/api/generate`    | Text generation          |
| `POST`   | `/api/chat`        | Chat completion          |
| `POST`   | `/api/embed`       | Embeddings               |
| `POST`   | `/api/pull`        | Pull (download) model    |
| `POST`   | `/api/show`        | Show model info          |
| `DELETE` | `/api/delete`      | Delete model             |
| `POST`   | `/api/compose`     | Pipeline execution       |
| `POST`   | `/api/structured`  | Structured output        |
