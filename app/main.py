import os
import asyncio

from fastapi import FastAPI
from prometheus_fastapi_instrumentator import Instrumentator

from sqlalchemy.ext.asyncio import AsyncEngine, create_async_engine
from sqlalchemy.sql import text as sql_text


DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgresql+asyncpg://postgres:postgres@postgres/postgres",
)

engine: AsyncEngine = create_async_engine(
    DATABASE_URL,
    pool_size=100,
    max_overflow=100,
)

async def ensure_dummy():
    async with engine.begin() as conn:
        await conn.execute(
            sql_text(
                "CREATE TABLE IF NOT EXISTS dummy ("
                " id serial primary key, "
                " ts timestamptz default now()"
                ")"
            )
        )

asyncio.create_task(ensure_dummy())


app = FastAPI(title="o11y-demo-service")
Instrumentator().instrument(app).expose(app)


@app.get("/items")
async def items():
    return {"msg": "ok"}


@app.get("/slow")
async def slow():
    await asyncio.sleep(float(os.getenv("SLOW_DELAY_SEC", "0.7")))
    return {"status": "ok"}


@app.get("/db_tx")
async def db_tx():
    async with engine.begin() as conn:
        await conn.execute(sql_text("INSERT INTO dummy DEFAULT VALUES"))
    return {"status": "tx_ok"}


@app.get("/health")
async def health():
    return {"status": "healthy"}