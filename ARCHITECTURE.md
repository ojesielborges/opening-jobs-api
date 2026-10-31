# Arquitetura

Os domínios de negócio vivem em `internal/business/<dominio>` e são organizados por responsabilidade:

- `domain`: entidades, validações e contratos; não importa HTTP, SQL ou Gin.
- `application`: casos de uso; recebe e repassa `context.Context`.
- `infrastructure`: implementações de portas externas, como PostgreSQL.
- `transport`: adaptadores de entrada, como HTTP; converte request/response e não contém regra de negócio.
- `module.go`: ponto de composição do domínio. O `cmd/api` apenas configura dependências globais e monta módulos.

`internal/controlplane` reúne capacidades administrativas e de configuração da aplicação, como tenants. `internal/operations` contém ferramentas operacionais, como migrations. `internal/bootstrap` inicializa configurações e dependências compartilhadas. `internal/platform` contém adaptadores técnicos reutilizáveis, como HTTP e PostgreSQL.

Para adicionar um domínio de negócio, replique a estrutura em `internal/business` e exponha somente o método de montagem no `module.go`. Domínios não devem importar a infraestrutura ou o transporte de outros domínios; colaboração entre eles deve ocorrer por contratos definidos na camada de aplicação.

## Contexto, timeout e autorização

O middleware global cria um prazo em `HTTP_REQUEST_TIMEOUT` e o handler entrega `c.Request.Context()` aos casos de uso. Qualquer consulta SQL ou cliente HTTP novo deve aceitar esse contexto e usar sua variante `...Context`, para cancelar a operação quando a requisição expirar ou o cliente desconectar.

Autorização deve ser middleware, aplicada no grupo de rotas que requer a política. Esse middleware autentica a requisição, coloca uma identidade mínima e tipada no `context.Context`, e a camada de aplicação decide se aquela identidade pode executar o caso de uso. Não use o contexto para dados de negócio nem faça verificações de permissão dentro de repositórios.

## Multi-tenancy

As rotas da API exigem `X-Tenant` e `X-Tenant-Key`. A chave nunca é armazenada em texto puro: o registry central `control.tenants` guarda somente seu SHA-256. O middleware resolve o tenant e os repositórios executam cada operação em uma transação com `SET LOCAL search_path`; assim uma conexão do pool não pode reutilizar o schema de outro tenant.

Para criar um tenant, execute `go run ./cmd/tenant --slug acme`. O comando cria `tenant_<slug>`, gera e exibe uma chave de API segura e registra somente o hash da chave; não executa migrations. Para definir uma chave específica, use `--api-key`; as variáveis `TENANT_SLUG` e `TENANT_API_KEY` também continuam aceitas como alternativa. Execute as migrations manualmente com `go run ./cmd/migration --tenant acme`.
