# Swagger Documenter Skill

Guidelines for Go REST handler annotations using `swaggo/swag`:

1. Every HTTP handler must contain tags (`@Tags`), accepted content types (`@Accept`), output types (`@Produce`), and parameters (`@Param`).
2. The model types (request and response) must be fully mapped and documented.
3. Keep route paths updated and match exactly the registered patterns on the Gin engine.
