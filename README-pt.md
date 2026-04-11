[English](README.md)

# Wallet Service — Visão Geral

O Wallet Service é um sistema de carteiras digitais que permite criar carteiras vinculadas a titulares e realizar transferências de saldo entre elas. Todos os valores monetários são representados em centavos (inteiros).

O projeto é composto por dois serviços independentes: `wallet-service` (backend transacional) e `wallet-analytics` (camada analítica).

## Arquitetura

O sistema adota os padrões **Event Sourcing** e **CQRS** (Command Query Responsibility Segregation).

**Lado de escrita (Command):** o estado da carteira não é armazenado diretamente em banco de dados relacional. Em vez disso, cada operação gera um evento de domínio que é persistido no KurrentDB (event store). O estado atual de uma carteira é reconstruído a cada operação replaying todos os eventos do seu stream.

**Lado de leitura (Query):** as consultas são servidas a partir de uma projeção no MySQL, mantida atualizada por um consumer assíncrono que consome os eventos do KurrentDB.

**Lado analítico:** uma segunda projeção em PostgreSQL é consumida pelo `wallet-analytics` via dbt para geração de marts analíticos.

```
Requisição HTTP (escrita)
  └─► WalletCommandController
        └─► WalletKurrentDBESHandler (replay do stream → agregado Wallet)
              └─► Wallet.TransferFunds() / NewWallet()
                    └─► Append de eventos no KurrentDB

KurrentDB (persistent subscriptions)
  ├─► WalletKurrentDBConsumer                       → processa recebimento de transferências
  ├─► WalletKurrentDBProjectorConsumer (MySQL)      → wallet_projections / transfer_projections
  └─► WalletKurrentDBProjectorConsumer (PostgreSQL) → wallet_projections / transfer_projections

Requisição HTTP (leitura)
  └─► WalletQueryController → consulta MySQL

PostgreSQL
  └─► wallet-analytics (dbt) → views de staging → mart_wallets / mart_transfers
```

## Stack Tecnológica

| Camada           | Tecnologia    |
| ---------------- | ------------- |
| Analytics        | dbt           |
| Event Store      | KurrentDB     |
| Banco de leitura | MySQL 8.0     |
| Banco analítico  | PostgreSQL 17 |

## wallet-service

Implantado como um binário monolítico (`cmd/monolith`). Na inicialização, cria três persistent subscriptions no KurrentDB e sobe o servidor HTTP.

### Modelo de Domínio

**Agregado `Wallet`:** encapsula o saldo e o histórico de transferências. Aplica as seguintes regras de negócio:

- Recebimento que ultrapasse o limite máximo: `ErrBalanceLimitExceeded`
- Saldo insuficiente para transferência: `ErrInsufficientBalance`
- Transferência duplicada (mesmo ID): `ErrDuplicateTransfer`
- Categoria não informada: assume `unclassified`

Categorias válidas: `food`, `fuel`, `sports`, `health`, `travel`, `essentials`, `entertainment`, `unclassified`.

**Eventos de domínio:**

| Evento                       | Gerado por                  |
| ---------------------------- | --------------------------- |
| `WalletCreatedEvent`         | criação de carteira         |
| `FundsTransferredEvent`      | débito (lado remetente)     |
| `FundsTransferReceivedEvent` | crédito (lado destinatário) |

O evento `FundsTransferredEvent` é consumido assincronamente pelo `WalletKurrentDBConsumer`, que carrega a carteira destinatária e emite o comando `ReceiveFundsTransfer`, fechando o ciclo da transferência.

### API HTTP

| Método | Rota                          | Descrição                                 |
| ------ | ----------------------------- | ----------------------------------------- |
| `POST` | `/wallets`                    | Cria uma carteira                         |
| `POST` | `/wallets/{id}/transfer`      | Inicia uma transferência                  |
| `POST` | `/wallets/{id}/mock-transfer` | Simula recebimento de transferência       |
| `GET`  | `/wallets`                    | Lista carteiras por titular (`holder_id`) |
| `GET`  | `/wallets/{id}`               | Consulta carteira por ID                  |
| `GET`  | `/wallets/{id}/transfers`     | Lista transferências de uma carteira      |

## wallet-analytics

Projeto dbt que lê as projeções do PostgreSQL e produz dois marts:

**`mart_wallets`:** resumo por carteira com saldo atual, categoria favorita, status (`active` ou `suspended` com base na última movimentação nos últimos 90 dias), contagens de transferências e valor total movimentado.

**`mart_transfers`:** histórico de transferências enriquecido com saldo antes e depois de cada transação, calculado via window functions SQL.

Testes customizados validam integridade dos dados: saldo positivo, matemática de saldo correta e consistência entre contagens.

## Ambiente Local

O ambiente de desenvolvimento é provisionado via Docker Compose:

```sh
docker compose -f sandbox/docker-compose.yaml up -d
```

| Container           | Porta       | Finalidade                           |
| ------------------- | ----------- | ------------------------------------ |
| `wallet-mysql`      | 3306        | Projeção de leitura                  |
| `wallet-postgresql` | 5432        | Projeção analítica                   |
| `wallet-kurrentdb`  | 2113 / 1113 | Event store (UI em `localhost:2113`) |

Os schemas das tabelas de projeção estão em `sandbox/create-wallet-tables-mysql.sql` e `sandbox/create-wallet-tables-postgresql.sql`. O script `sandbox/populate-wallet.js` pode ser usado para seed de dados.
