# Mini E-commerce em Go

API de um mini e-commerce construída como monólito modular com Go, Gin, PostgreSQL, GORM, JWT e Docker.

O projeto implementa autenticação, catálogo, estoque, pedidos, pagamentos simulados e checkout transacional. A documentação interativa da API é disponibilizada com Swagger.

## Requisitos

Para executar a aplicação localmente, instale:

- Go 1.26.2 ou uma versão compatível;
- Docker com Docker Compose;
- Make, caso queira utilizar os comandos padronizados do projeto.

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

Altere o segredo JWT antes de utilizar a aplicação fora do ambiente local.

## Execução local

Inicie apenas o PostgreSQL:

```bash
make database-up
```

Execute a API:

```bash
make run
```

O comando equivalente sem Make é:

```bash
go run ./cmd/api
```

A API ficará disponível em `http://localhost:3000`.

## Execução com Docker Compose

Construa e inicie a aplicação e o PostgreSQL:

```bash
make compose-up
```

Para acompanhar os logs:

```bash
docker compose logs -f application
```

Para encerrar os serviços:

```bash
make compose-down
```

## Verificação da aplicação

Consulte o health check:

```bash
curl http://localhost:3000/health
```

Resposta esperada:

```json
{"status":"UP"}
```

## Swagger

Com a aplicação em execução, acesse:

```text
http://localhost:3000/swagger/index.html
```

Para regenerar os arquivos do Swagger após alterar handlers, DTOs ou rotas:

```bash
make swagger
```

## Testes e análise estática

Execute os testes sem utilizar resultados armazenados em cache:

```bash
make test
```

Execute a análise estática:

```bash
make vet
```

Execute formatação, testes e análise estática em sequência:

```bash
make check
```

## Build

Gere o executável em `bin/ecommerce`:

```bash
make build
```

## Rotas principais

| Método | Rota | Acesso |
| --- | --- | --- |
| `GET` | `/health` | Público |
| `POST` | `/api/authentication/register` | Público |
| `POST` | `/api/authentication/login` | Público |
| `GET` | `/api/products` | Público |
| `GET` | `/api/products/:productID` | Público |
| `POST` | `/api/products` | Administrador |
| `PUT` | `/api/inventory/:productID` | Administrador |
| `POST` | `/api/orders` | Autenticado |
| `GET` | `/api/orders` | Autenticado |
| `GET` | `/api/orders/:orderID` | Autenticado |

Rotas protegidas esperam o token no cabeçalho:

```http
Authorization: Bearer ACCESS_TOKEN
```

Durante o desenvolvimento, um cliente pode ser promovido a administrador diretamente no PostgreSQL:

```sql
UPDATE users
SET roles = 'CUSTOMER,ADMIN'
WHERE email = 'administrator@example.com';
```

Faça login novamente depois da alteração para receber um token com os papéis atualizados.

## Estrutura

```text
cmd/api/                         composição, roteador e rotas
internal/authentication/         cadastro, login e autorização
internal/catalog/                produtos e paginação
internal/inventory/              controle e reserva de estoque
internal/order/                  pedidos e ownership
internal/payment/                pagamento simulado
internal/checkout/               fluxo transacional de compra
internal/platform/               configuração e infraestrutura compartilhada
docs/swagger/                    documentação Swagger gerada
```

Cada módulo separa domínio, casos de uso e adapters. As dependências apontam dos adapters para os casos de uso e destes para o domínio.
