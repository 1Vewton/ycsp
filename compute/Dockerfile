FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    UV_LINK_MODE=copy

WORKDIR /app

RUN pip install --no-cache-dir uv

COPY pyproject.toml uv.lock ./
RUN uv sync --frozen

COPY . .

EXPOSE 8001

CMD ["sh", "-c", "uv sync --frozen && exec uv run uvicorn app.main:app --host 0.0.0.0 --port 8001 --reload"]
