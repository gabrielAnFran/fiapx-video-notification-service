# fiapx-video-notification-service

Microsservico de notificacao por e-mail do sistema de processamento de
video FIAP X (upload de video -> extracao assincrona de frames -> zip para
download), um dos quatro microsservicos do hackathon. Este e o mais simples
dos quatro: um worker que consome um comando "notifique o usuario" e envia
um e-mail via SMTP, registrando o que foi enviado.

## Arquitetura

Este servico nao chama nenhum outro servico diretamente e nao publica nenhum
evento: a comunicacao acontece exclusivamente via um comando consumido do
RabbitMQ (exchange topic `video.events`). Por ser um consumidor terminal,
nao ha outbox nem `cmd/outbox-dispatcher` neste repositorio — nao ha nada
para este servico publicar de volta na saga.

Organizacao dos processos (2 binarios, 1 banco de dados, 1 broker AMQP
compartilhado):

- `cmd/server` — expoe apenas `/healthz` e `/readyz` para probes de
  Kubernetes; roda as migrations no boot. Nao ha rotas de negocio.
- `cmd/worker` — consome comandos do RabbitMQ (`notification-service.events.q`)
  e envia e-mails via SMTP (atualmente: `video.notify.requested`).

## Participacao na saga

- **Emite**: nada. Este servico e um consumidor terminal.
- **Consome**: `video.notify.requested` — comando com o payload
  `{video_id, user_id, recipient_email, notification_type, original_filename,
  error_message?}`, onde `notification_type` e `"COMPLETED"` ou `"FAILED"`.
  Ao processar, envia um e-mail ao usuario e grava um registro de auditoria
  na tabela `notification_log` (`status` `SENT` ou `FAILED`).
- **Idempotencia**: tabela `processed_events`, indexada por `event_id`,
  verificada antes do processamento. O evento so e marcado como processado
  **depois** de um envio de e-mail bem sucedido — nao antes. Isso e
  proposital: se o evento fosse marcado como processado independentemente do
  resultado do envio, uma falha transiente de SMTP nunca seria reentregue,
  porque a checagem de idempotencia bloquearia o reprocessamento no reenvio
  da mensagem. Por isso, em caso de falha de envio o registro de auditoria
  ainda e gravado (com `status = FAILED`, para fins de auditoria), mas o erro
  e retornado para que o mecanismo de retry/DLQ do `amqp.go` cuide da
  reentrega. Apos `MaxRetries` tentativas, a mensagem cai na DLQ — esse e o
  comportamento terminal correto aqui, ja que o status do video ja esta
  gravado de forma durável como `COMPLETED`/`FAILED` no upload-service
  independentemente do e-mail de notificacao ter sido entregue ou nao; esse
  caminho de retry/DLQ afeta apenas a notificacao "bonus", nunca a corretude
  do sistema principal.

## Envio de e-mail

`internal/infrastructure/email.Sender` e configurado via variaveis de
ambiente `SMTP_*`. Comportamento:

- Se `SMTP_HOST` ou `SMTP_PORT` estiverem vazios, o envio vira um no-op que
  apenas loga "SMTP not configured, skipping" e retorna sucesso — util para
  rodar o worker localmente sem credenciais SMTP reais.
- Se `SMTP_USER` e `SMTP_PASSWORD` estiverem **ambos** vazios (mas host/porta
  configurados), o envio e feito sem autenticacao (`auth = nil`) — o modo
  esperado para testar localmente contra o [Mailhog](https://github.com/mailhog/MailHog)
  (`SMTP_HOST=localhost`, `SMTP_PORT=1025`), que aceita SMTP sem autenticacao.
- Caso contrario, usa `smtp.PlainAuth` normalmente (SMTP real, ex.: um
  provedor transacional de e-mail).

## Rodando localmente (standalone)

Requer Postgres e RabbitMQ acessiveis pelas variaveis de ambiente abaixo.
Exemplo usando containers Docker locais (e opcionalmente Mailhog para
capturar e-mails sem enviar de verdade):

```bash
docker run -d --name notification-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=notification_service -p 5432:5432 postgres:16-alpine
docker run -d --name notification-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine
docker run -d --name notification-mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog

export NOTIFICATION_DB_DSN="host=localhost user=postgres password=postgres dbname=notification_service port=5432 sslmode=disable"
export NOTIFICATION_AMQP_URL="amqp://guest:guest@localhost:5672/"
export SMTP_HOST=localhost
export SMTP_PORT=1025
export SMTP_FROM_EMAIL=no-reply@fiapx.local

go run ./cmd/server   # apenas /healthz e /readyz na porta :8083, roda migrations no boot
go run ./cmd/worker   # processo separado, consome video.notify.requested e envia e-mails
```

A UI web do Mailhog (`http://localhost:8025`) mostra os e-mails "enviados"
sem sair da maquina local.

Ou usando as imagens de container geradas a partir do `Dockerfile` deste
repositorio:

```bash
docker build --build-arg TARGET=worker -t fiapx-video-notification-service:worker .
docker run --rm \
  -e NOTIFICATION_DB_DSN="host.docker.internal:..." \
  -e NOTIFICATION_AMQP_URL="amqp://guest:guest@host.docker.internal:5672/" \
  -e SMTP_HOST=host.docker.internal -e SMTP_PORT=1025 \
  fiapx-video-notification-service:worker
```

## Rodando como parte da stack completa

Este servico e composto como parte da stack completa do sistema FIAP X
junto com os outros tres microsservicos irmaos (upload, processamento de
frames e o orquestrador/saga), compartilhando Postgres e RabbitMQ.

## Variaveis de ambiente

| Variavel                | Padrao                                                                                                | Descricao                                     |
|--------------------------|--------------------------------------------------------------------------------------------------------|------------------------------------------------|
| `NOTIFICATION_PORT`      | `8083`                                                                                                  | Porta HTTP do `cmd/server` (apenas health checks) |
| `NOTIFICATION_DB_DSN`    | `host=localhost user=postgres password=postgres dbname=notification_service port=5432 sslmode=disable` | DSN do Postgres (formato GORM/`lib/pq`)         |
| `NOTIFICATION_AMQP_URL`  | `amqp://guest:guest@localhost:5672/`                                                                    | URL de conexao com o RabbitMQ                   |
| `SMTP_HOST`              | *(vazio)*                                                                                                | Host do servidor SMTP; vazio = modo no-op (loga e nao envia) |
| `SMTP_PORT`              | *(vazio)*                                                                                                | Porta do servidor SMTP                          |
| `SMTP_USER`              | *(vazio)*                                                                                                | Usuario SMTP; se vazio junto com `SMTP_PASSWORD`, envia sem autenticacao (modo Mailhog) |
| `SMTP_PASSWORD`          | *(vazio)*                                                                                                | Senha SMTP                                      |
| `SMTP_FROM_EMAIL`        | `no-reply@fiapx.local`                                                                                  | Endereco de remetente usado no cabecalho `From` |

## Rotas HTTP

- `GET /healthz` — liveness, sempre `200`.
- `GET /readyz` — readiness, faz ping no Postgres; `503` se indisponivel.

Este servico nao expoe nenhuma rota de negocio: toda a logica roda no
`cmd/worker`, disparada por eventos.

## Testes

```bash
go test ./...                                              # apenas testes unitarios, sem dependencias externas, rapido
go test -tags=integration ./...                            # unitarios + integracao, precisa de Docker (deferred — placeholder ainda vazio)
```

O teste de integracao em `tests/integration/` fica atras da build tag
`integration` justamente para que `go test ./...` continue rapido e sem
dependencias externas. Testes reais com testcontainers-go (Postgres +
RabbitMQ, e idealmente Mailhog para validar o envio de e-mail ponta a ponta)
ficaram como proximo passo (stretch goal), nao implementados neste momento —
o arquivo hoje e apenas um placeholder documentando a intencao.

### Cobertura

Cobertura por pacote, via `go test ./... -coverprofile=coverage.out`:

| Pacote                                | Como                                 |
|-----------------------------------------|----------------------------------------|
| `internal/domain/entities`              | trivial (struct pura, sem logica)      |
| `internal/application/usecases`         | unit (fakes em memoria)                |
| `internal/infrastructure/email`         | unit (subject/body extraidos em funcoes puras, testados sem rede) |
| `internal/infrastructure/config`        | unit                                   |
| `internal/infrastructure/db`            | fora de escopo aqui — precisa de integracao (nao implementada ainda) |
| `internal/infrastructure/messaging`     | fora de escopo aqui — precisa de integracao (nao implementada ainda) |
| `cmd/*`                                 | fora de escopo (wiring de `main()`)    |

`internal/application/usecases` cobre `SendNotificationUseCase` nos tres
caminhos que importam: sucesso (grava `SENT`, marca processado), falha de
envio (grava `FAILED` com a mensagem de erro, retorna erro, **nao** marca
processado — para permitir retry), e replay idempotente (evento ja
processado e um no-op completo, sem chamar o sender nem gravar log de novo).
`internal/infrastructure/email` cobre o caminho "SMTP nao configurado" (deve
retornar sucesso sem tentar enviar) e a construcao pura de assunto/corpo para
os tipos `COMPLETED` e `FAILED`.

## Notas de desenvolvimento / desvios da especificacao

- Segue as mesmas convencoes de `pos-os-service`/`project1`: modulo Go 1.23,
  Clean Architecture (`domain` / `application` / `infrastructure` /
  `presentation`), GORM + `golang-migrate` para persistencia, e o mesmo
  helper `internal/infrastructure/messaging/amqp.go` (copiado byte-a-byte)
  compartilhado entre os quatro microsservicos do sistema, para garantir que
  o contrato de eventos (exchanges, retry, DLQ) seja identico entre eles.
- Sem `cmd/outbox-dispatcher`: este servico nunca publica eventos, entao o
  padrao outbox (usado em `pos-os-service` para escritas HTTP) nao se
  aplica aqui.
- Sem HPA no chart Helm: e um worker de baixo throughput (um e-mail por
  video processado), nao ha necessidade de autoscaling horizontal para o
  escopo do hackathon.
