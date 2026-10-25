# Arquitetura

Cada domínio vive em `internal/<dominio>` e é organizado por responsabilidade:

- `domain`: entidades, validações e contratos; não importa HTTP, SQL ou Gin.
- `application`: casos de uso; recebe e repassa `context.Context`.
- `infrastructure`: implementações de portas externas, como PostgreSQL.
- `transport`: adaptadores de entrada, como HTTP; converte request/response e não contém regra de negócio.
- `module.go`: ponto de composição do domínio. O `cmd/api` apenas configura dependências globais e monta módulos.

Para adicionar um domínio, replique essa estrutura e exponha somente o método de montagem no `module.go`. Domínios não devem importar a infraestrutura ou o transporte de outros domínios; colaboração entre eles deve ocorrer por contratos definidos na camada de aplicação.

## Contexto, timeout e autorização

O middleware global cria um prazo em `HTTP_REQUEST_TIMEOUT` e o handler entrega `c.Request.Context()` aos casos de uso. Qualquer consulta SQL ou cliente HTTP novo deve aceitar esse contexto e usar sua variante `...Context`, para cancelar a operação quando a requisição expirar ou o cliente desconectar.

Autorização deve ser middleware, aplicada no grupo de rotas que requer a política. Esse middleware autentica a requisição, coloca uma identidade mínima e tipada no `context.Context`, e a camada de aplicação decide se aquela identidade pode executar o caso de uso. Não use o contexto para dados de negócio nem faça verificações de permissão dentro de repositórios.

## Multi-tenancy

As rotas da API exigem `X-Tenant` e `X-Tenant-Key`. A chave nunca é armazenada em texto puro: o registry central `control.tenants` guarda somente seu SHA-256. O middleware resolve o tenant e os repositórios executam cada operação em uma transação com `SET LOCAL search_path`; assim uma conexão do pool não pode reutilizar o schema de outro tenant.

Para criar o primeiro tenant (ou rotacionar sua chave), defina `TENANT_SLUG`, `TENANT_API_KEY` com ao menos 32 caracteres e `MIGRATIONS_PATH`, então execute `go run ./cmd/tenant`. O comando cria `tenant_<slug>`, registra a credencial e aplica todas as migrações `*_up.sql` ainda pendentes naquele schema.
