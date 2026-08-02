#!/bin/sh
set -eu

base_url="${BASE_URL:-http://localhost:3000}"
demo_suffix="$(date +%s)"
administrator_email="administrator-${demo_suffix}@example.com"
customer_email="customer-${demo_suffix}@example.com"
password="password"

json_request() {
	method="$1"
	path="$2"
	token="$3"
	body="$4"

	if [ -n "$token" ]; then
		curl --silent --show-error \
			--request "$method" \
			--header "Content-Type: application/json" \
			--header "Authorization: Bearer $token" \
			--data "$body" \
			"${base_url}${path}"
	else
		curl --silent --show-error \
			--request "$method" \
			--header "Content-Type: application/json" \
			--data "$body" \
			"${base_url}${path}"
	fi
}

show_request() {
	method="$1"
	path="$2"
	token="$3"
	body="$4"

	echo
	echo "${method} ${path}"
	if [ -n "$token" ]; then
		curl --silent --show-error \
			--request "$method" \
			--header "Content-Type: application/json" \
			--header "Authorization: Bearer $token" \
			--data "$body" \
			--write-out '\nHTTP %{http_code}\n' \
			"${base_url}${path}"
	else
		curl --silent --show-error \
			--request "$method" \
			--header "Content-Type: application/json" \
			--data "$body" \
			--write-out '\nHTTP %{http_code}\n' \
			"${base_url}${path}"
	fi
}

echo "GET /health"
curl --silent --show-error --write-out '\nHTTP %{http_code}\n' "${base_url}/health"

administrator_body="$(json_request POST /api/authentication/register "" "{\"email\":\"${administrator_email}\",\"password\":\"${password}\"}")"
administrator_id="$(printf '%s' "$administrator_body" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
if [ -z "$administrator_id" ]; then
	echo "administrator registration failed: ${administrator_body}" >&2
	exit 1
fi

if [ -n "${POSTGRES_CONTAINER:-}" ]; then
	docker exec -i "$POSTGRES_CONTAINER" psql \
		--username afraniocaires \
		--dbname ecommerce \
		--command "UPDATE customers SET roles = 'CUSTOMER,ADMIN' WHERE id = '${administrator_id}';" \
		>/dev/null
else
	docker compose exec -T postgresql psql \
		--username afraniocaires \
		--dbname ecommerce \
		--command "UPDATE customers SET roles = 'CUSTOMER,ADMIN' WHERE id = '${administrator_id}';" \
		>/dev/null
fi

administrator_login="$(json_request POST /api/authentication/login "" "{\"email\":\"${administrator_email}\",\"password\":\"${password}\"}")"
administrator_token="$(printf '%s' "$administrator_login" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')"

customer_body="$(json_request POST /api/authentication/register "" "{\"email\":\"${customer_email}\",\"password\":\"${password}\"}")"
customer_id="$(printf '%s' "$customer_body" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
customer_login="$(json_request POST /api/authentication/login "" "{\"email\":\"${customer_email}\",\"password\":\"${password}\"}")"
customer_token="$(printf '%s' "$customer_login" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')"

if [ -z "$administrator_token" ] || [ -z "$customer_id" ] || [ -z "$customer_token" ]; then
	echo "authentication setup failed" >&2
	exit 1
fi

product_body="$(json_request POST /api/products "$administrator_token" '{"name":"Mechanical Keyboard","description":"HTTP flow product","price_cents":10000}')"
product_id="$(printf '%s' "$product_body" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
if [ -z "$product_id" ]; then
	echo "product creation failed: ${product_body}" >&2
	exit 1
fi

show_request PUT "/api/inventory/${product_id}" "$administrator_token" '{"quantity":5}'
show_request POST /api/orders "$customer_token" "{\"items\":[{\"product_id\":\"${product_id}\",\"quantity\":2}]}"

echo
echo "GET /api/orders?limit=10&offset=0"
curl --silent --show-error \
	--header "Authorization: Bearer ${customer_token}" \
	--write-out '\nHTTP %{http_code}\n' \
	"${base_url}/api/orders?limit=10&offset=0"

show_request POST /api/products "" '{"name":"Unauthorized","price_cents":100}'
show_request POST /api/orders "$customer_token" '{'
show_request POST /api/orders "$customer_token" "{\"items\":[{\"product_id\":\"${product_id}\",\"quantity\":999}]}"
