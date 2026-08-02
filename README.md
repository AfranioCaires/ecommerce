# Mini E-commerce em Go

API HTTP de um mini e-commerce construída como monólito modular com Go, `net/http`, PostgreSQL, `sqlc`, JWT e Docker.

O projeto implementa autenticação de clientes, catálogo, estoque, pedidos, pagamentos simulados e checkout transacional. O schema é criado por migrations SQL versionadas; GORM e Gin não são utilizados.

## Requisitos

- Go 1.26.2 ou compatível;
- Docker com Docker Compose;
- Make para os comandos padronizados.

O gerador `sqlc` é executado com versão fixada por `go run`, portanto não exige instalação global.

## Configuração

Crie o arquivo de ambiente a partir do exemplo:

```bash
cp .env.example .env
```

As principais variáveis são:

```dotenv
APPLICATION_PORT=3000
APPLICATION_ENVIRONMENT=development
POSTGRESQL_DATA_SOURCE=host=localhost port=5432 user=afraniocaires password=postgres dbname=ecommerce sslmode=disable
JSON_WEB_TOKEN_SECRET=RED-DEAD-REDEMPTION-2
JSON_WEB_TOKEN_ISSUER=afranio
JSON_WEB_TOKEN_LIFETIME=15m
```

Troque o segredo JWT antes de usar a aplicação fora do ambiente local.

## Execução

Inicie o PostgreSQL e execute a API:

```bash
make database-up
make run
```

A API aplica automaticamente todas as migrations pendentes antes de abrir o servidor em `http://localhost:3000`.

Para executar tudo com Docker Compose:

```bash
make compose-up
docker compose logs -f application
```

Para encerrar os serviços:

```bash
make compose-down
```

## Migrations

As migrations reversíveis ficam em `internal/platform/database/migrations` e são incorporadas aos binários da API e do migrador.

```bash
make migrate-up
make migrate-down
```

A migration inicial cria:

- `customers`, incluindo `password_hash`;
- `products` e `stocks`;
- `orders`, `order_items` e `payments`;
- chaves estrangeiras entre clientes, pedidos, itens, produtos, estoque e pagamentos;
- índices usados pelas consultas paginadas.

## sqlc

As queries SQL ficam em `internal/platform/database/queries`. Para regenerar os arquivos Go tipados:

```bash
make sqlc
```

Para regenerar e falhar caso o resultado não esteja versionado:

```bash
make sqlc-check
```

Os repositories em `adapter/repository/sqlc` usam somente as queries geradas. O checkout compartilha uma `sql.Tx` entre catálogo, estoque, pedido e pagamento; qualquer erro causa rollback.

## Testes e cobertura

```bash
make test
make coverage
make vet
make check
```

`make coverage` mede statements globalmente com `-coverpkg=./...` e falha abaixo de 40%. Os testes cobrem domínio, casos de uso, SQL/repositories, transações, estoque, handlers, rotas, configuração e banco.

## Demonstração HTTP

Com a aplicação e o PostgreSQL do Compose em execução:

```bash
make demo
```

O script demonstra health check, cadastro, login, criação de produto, definição de estoque, checkout e listagem com `limit`/`offset`. Também mostra respostas de erro para acesso sem token, JSON inválido e estoque insuficiente.

## Build

```bash
make build
```

O executável é criado em `bin/ecommerce`.

## Rotas principais

| Método | Rota | Acesso |
| --- | --- | --- |
| `GET` | `/health` | Público |
| `POST` | `/api/authentication/register` | Público |
| `POST` | `/api/authentication/login` | Público |
| `GET` | `/api/products` | Público |
| `GET` | `/api/products/{productID}` | Público |
| `POST` | `/api/products` | Administrador |
| `PUT` | `/api/inventory/{productID}` | Administrador |
| `POST` | `/api/orders` | Autenticado |
| `GET` | `/api/orders?limit=20&offset=0` | Autenticado |
| `GET` | `/api/orders/{orderID}` | Autenticado |

Rotas protegidas recebem o token no cabeçalho:

```http
Authorization: Bearer ACCESS_TOKEN
```

Durante o desenvolvimento, um cliente pode ser promovido a administrador diretamente no PostgreSQL:

```sql
UPDATE customers
SET roles = 'CUSTOMER,ADMIN'
WHERE email = 'administrator@example.com';
```

Faça login novamente para emitir um token com os papéis atualizados.

## Estrutura

```text
cmd/api/                                  bootstrap e roteador net/http
cmd/migrate/                              executável de migrations
internal/*/domain/                        entidades e regras de domínio
internal/*/usecase/                       serviços e casos de uso
internal/*/adapter/http/                  handlers HTTP
internal/*/adapter/repository/sqlc/       repositories SQL
internal/platform/database/migrations/    migrations up/down
internal/platform/database/queries/       queries consumidas pelo sqlc
internal/platform/database/sqlc/          código Go gerado
internal/platform/transaction/            transações compartilhadas
scripts/http-flow.sh                      fluxo feliz e erros via HTTP
```
