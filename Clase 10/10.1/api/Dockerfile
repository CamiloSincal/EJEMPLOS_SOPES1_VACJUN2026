# ── Etapa 1: compilar ────────────────────────────────────────────────────────
FROM rust:1.78-slim AS builder

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY src ./src

RUN cargo build --release

# ── Etapa 2: imagen final (mínima) ───────────────────────────────────────────
FROM debian:bookworm-slim

WORKDIR /app
COPY --from=builder /app/target/release/api .

EXPOSE 8080
CMD ["./api"]