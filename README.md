# o11y 🚀  
  
На вм в облаке развёрнуты сервис, СУБД, нагрузочный генератор, стек наблюдаемости и алерты в тг канал.  
  
> VM IP: **51.250.86.234**  
  
---  
  
## Состав  
| Компонент           | URL                         | Учётные данные          |     |
| ------------------- | --------------------------- | ----------------------- | --- |
| FastAPI service     | `http://51.250.86.234:8000` | `/docs` – Swagger       |     |
| Load Generator (Go) | `http://51.250.86.234:8080` | UI для запуска нагрузки |     |
| Prometheus          | `http://51.250.86.234:9090` | —                       |     |
| Alertmanager        | `http://51.250.86.234:9093` | —                       |     |
| Grafana             | `http://51.250.86.234:3000` | admin / admin           |     |
| Telegram канал      | t.me/pypytyty               | публичные алерты        |     |
  
Все сервисы стартуют одной командой Docker Compose, конфиги лежат в каталоге `infra/`.  
  
---  
  
## Как проверить алерты  
1. Откройте генератор нагрузки: `http://51.250.86.234:8080`  
2. Заполните поля:  
   * **Path** – `/db_tx,/slow`  
   * **RPS** – `220`  
   * **Duration** – `200`  
3. Нажмите *Start*. Через ~30 с метрики превысят пороги:  
   * `pg_stat_database_xact_commit` > **100 TPS**  
   * p99 latency > **500 ms**  
4. В тг канале появятся два отдельных алерта.

  `Если будешь проверять алерты по отдельности, то в пути оставь что-то одно и выставь значение RPS ~110-120, так как если 2 сразу, то он делит рпс пополам на каждый путь`
---  
  
## Что где лежит  
| Путь | Назначение |  
|------|-------------|  
| `app/` | FastAPI-приложение + эндпоинты `/db_tx`, `/slow`, `/health`, `/docs` |  
| `load-go/` | Go-сервис с UI для генерации HTTP-нагрузки |  
| `infra/prometheus/` | `prometheus.yml` + правила оповещений |  
| `infra/alertmanager/` | `config.yml` для доставки в Telegram |  
| `infra/grafana/provisioning/` | автоматически загружаемые datasources & dashboards |  
| `data/` | CSV-файл (≈1500 строк) для первоначального импорта в Postgres |  
  
---  
  
## Основные метрики и правила  
| Alert | Выражение | Порог / for |  
|-------|-----------|-------------|  
| **HighDBRPS** | `rate(pg_stat_database_xact_commit{datname="postgres"}[15s])` | >100 TPS, 30 s |  
| **HighP99Latency** | `1000*histogram_quantile(0.99, sum by(le)(rate(http_request_duration_seconds_bucket[5m])))` | >500 ms, 30 s |  
  
Все правила находятся в `infra/prometheus/rules/app_alerts.yml`.  
  
---  
  
## Инфраструктура как код (IaC)  
* **Prometheus** – конфиг + rules  
* **Alertmanager** – конфиг с Telegram receiver  
* **Grafana** – datasources и dashboards provisioning  
  
Изменения применяются просто: правим файлы → `docker compose up -d --build prometheus alertmanager grafana`.  
  
---  

## Полезные команды  
```bash  
# логи сервисов  
docker compose logs -f app         # FastAPI  
docker compose logs -f load_go     # генератор  
  
# доступ в psql  
docker compose exec postgres psql -U postgres  
  
# проверка состояния алертов  
curl -s http://localhost:9090/api/v1/alerts | jq  
```  
