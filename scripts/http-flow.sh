#!/bin/sh
set -eu

base_url="${BASE_URL:-http://localhost:3000}"
payment_base_url="${PAYMENT_BASE_URL:-http://localhost:3001}"
demo_suffix="$(date +%s)"
administrator_email="administrator-${demo_suffix}@example.com"
customer_email="customer-${demo_suffix}@example.com"
password="password"
response_file="$(mktemp)"
trap 'rm -f "$response_file"' EXIT

perform_request() {
	request_method="$1"
	request_path="$2"
	request_token="$3"
	request_body="$4"
	request_url="${base_url}${request_path}"
	if [ -n "$request_token" ] && [ -n "$request_body" ]; then
		response_code="$(curl --silent --show-error --output "$response_file" --write-out '%{http_code}' --request "$request_method" --header 'Content-Type: application/json' --header "Authorization: Bearer ${request_token}" --data "$request_body" "$request_url")"
	elif [ -n "$request_token" ]; then
		response_code="$(curl --silent --show-error --output "$response_file" --write-out '%{http_code}' --request "$request_method" --header "Authorization: Bearer ${request_token}" "$request_url")"
	elif [ -n "$request_body" ]; then
		response_code="$(curl --silent --show-error --output "$response_file" --write-out '%{http_code}' --request "$request_method" --header 'Content-Type: application/json' --data "$request_body" "$request_url")"
	else
		response_code="$(curl --silent --show-error --output "$response_file" --write-out '%{http_code}' --request "$request_method" "$request_url")"
	fi
	response_body="$(tr -d '\r\n' < "$response_file")"
}

expect_request() {
	expected_code="$1"
	shift
	perform_request "$@"
	printf '\n%s %s\n%s\nHTTP %s\n' "$1" "$2" "$response_body" "$response_code"
	if [ "$response_code" != "$expected_code" ]; then
		echo "expected HTTP ${expected_code}, received ${response_code}" >&2
		exit 1
	fi
}

json_string() {
	printf '%s' "$1" | sed -n "s/.*\"$2\":\"\([^\"]*\)\".*/\1/p"
}

promote_administrator() {
	administrator_id="$1"
	if [ -n "${POSTGRES_CONTAINER:-}" ]; then
		docker exec -i "$POSTGRES_CONTAINER" psql --username afraniocaires --dbname ecommerce --command "UPDATE customers SET roles = 'CUSTOMER,ADMIN' WHERE id = '${administrator_id}';" >/dev/null
	else
		docker compose exec -T ecommerce-postgresql psql --username afraniocaires --dbname ecommerce --command "UPDATE customers SET roles = 'CUSTOMER,ADMIN' WHERE id = '${administrator_id}';" >/dev/null
	fi
}

stock_quantity() {
	product_id="$1"
	if [ -n "${POSTGRES_CONTAINER:-}" ]; then
		docker exec -i "$POSTGRES_CONTAINER" psql --tuples-only --no-align --username afraniocaires --dbname ecommerce --command "SELECT quantity FROM stocks WHERE product_id = '${product_id}';" | tr -d '[:space:]'
	else
		docker compose exec -T ecommerce-postgresql psql --tuples-only --no-align --username afraniocaires --dbname ecommerce --command "SELECT quantity FROM stocks WHERE product_id = '${product_id}';" | tr -d '[:space:]'
	fi
}

wait_for_order() {
	order_id="$1"
	expected_status="$2"
	attempt=1
	while [ "$attempt" -le 30 ]; do
		perform_request GET "/pedidos/${order_id}" "" ""
		polled_status="$(json_string "$response_body" status)"
		if [ "$polled_status" = "$expected_status" ]; then
			printf 'Pedido %s alcançou %s na tentativa %s.\n' "$order_id" "$polled_status" "$attempt"
			return 0
		fi
		sleep 1
		attempt=$((attempt + 1))
	done
	echo "pedido ${order_id} não alcançou ${expected_status}; último status: ${polled_status:-desconhecido}" >&2
	exit 1
}

expect_request 200 GET /health "" ""
payment_health_code="$(curl --silent --show-error --output "$response_file" --write-out '%{http_code}' "${payment_base_url}/health")"
if [ "$payment_health_code" != "200" ]; then
	echo "payment-service health returned HTTP ${payment_health_code}" >&2
	exit 1
fi
printf '\nGET %s/health\n%s\nHTTP %s\n' "$payment_base_url" "$(tr -d '\r\n' < "$response_file")" "$payment_health_code"

expect_request 201 POST /clientes "" "{\"name\":\"Administrator\",\"email\":\"${administrator_email}\",\"password\":\"${password}\"}"
administrator_id="$(json_string "$response_body" id)"
promote_administrator "$administrator_id"
perform_request POST /api/authentication/login "" "{\"email\":\"${administrator_email}\",\"password\":\"${password}\"}"
administrator_token="$(json_string "$response_body" access_token)"

expect_request 201 POST /clientes "" "{\"name\":\"Saga Customer\",\"email\":\"${customer_email}\",\"passwordHash\":\"${password}\"}"
customer_id="$(json_string "$response_body" id)"
if [ -z "$administrator_id" ] || [ -z "$administrator_token" ] || [ -z "$customer_id" ]; then
	echo "authentication setup failed" >&2
	exit 1
fi

expect_request 201 POST /produtos "" '{"name":"Approved product","description":"Saga approval demo","price_cents":1000}'
approved_product_id="$(json_string "$response_body" id)"
expect_request 201 POST /produtos "" '{"name":"Declined product","description":"Saga compensation demo","price_cents":1013}'
declined_product_id="$(json_string "$response_body" id)"

expect_request 200 PUT "/api/inventory/${approved_product_id}" "$administrator_token" '{"quantity":5}'
expect_request 200 PUT "/api/inventory/${declined_product_id}" "$administrator_token" '{"quantity":5}'

expect_request 201 POST /pedidos "" "{\"clienteId\":\"${customer_id}\",\"itens\":[{\"produtoId\":\"${approved_product_id}\",\"quantidade\":1}]}"
approved_order_id="$(json_string "$response_body" order_id)"
expect_request 201 POST /pedidos "" "{\"clienteId\":\"${customer_id}\",\"itens\":[{\"produtoId\":\"${declined_product_id}\",\"quantidade\":1}]}"
declined_order_id="$(json_string "$response_body" order_id)"

reserved_declined_stock="$(stock_quantity "$declined_product_id")"
if [ "$reserved_declined_stock" != "4" ]; then
	echo "expected declined-product stock 4 after reservation, received ${reserved_declined_stock}" >&2
	exit 1
fi

expect_request 202 POST "/pedidos/${approved_order_id}/pagar" "" ""
approved_correlation_id="$(json_string "$response_body" correlationId)"
wait_for_order "$approved_order_id" PAID

expect_request 202 POST "/pedidos/${declined_order_id}/pagar" "" ""
declined_correlation_id="$(json_string "$response_body" correlationId)"
wait_for_order "$declined_order_id" CANCELED

approved_stock="$(stock_quantity "$approved_product_id")"
compensated_stock="$(stock_quantity "$declined_product_id")"
if [ "$approved_stock" != "4" ] || [ "$compensated_stock" != "5" ]; then
	echo "unexpected final stock: approved=${approved_stock}, compensated=${compensated_stock}" >&2
	exit 1
fi
printf '\nCorrelação aprovada: %s; estoque final: %s\n' "$approved_correlation_id" "$approved_stock"
printf 'Correlação recusada: %s; estoque restaurado: %s\n' "$declined_correlation_id" "$compensated_stock"

expect_request 409 POST "/pedidos/${approved_order_id}/pagar" "" ""
expect_request 409 POST "/pedidos/${declined_order_id}/cancelar" "" ""
expect_request 404 GET /pedidos/order-does-not-exist "" ""
expect_request 400 POST /pedidos "" '{'
expect_request 200 GET '/pedidos?limit=2&offset=0' "" ""

echo "Demonstração concluída: aprovação, recusa, compensação, idempotência HTTP e erros validados."
