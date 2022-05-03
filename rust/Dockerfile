FROM rust:latest as builder

ENV CARGO_HOME=/tmp/cargo
RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y ca-certificates libssl1.1 libssl-dev musl musl-dev musl-tools
RUN rustup target add x86_64-unknown-linux-musl

WORKDIR /rust
COPY    ./Cargo.toml ./
COPY    ./Cargo.toml ./
COPY    src ./src
RUN     cargo fetch --verbose
RUN     cargo build --release --target x86_64-unknown-linux-musl

FROM alpine
WORKDIR /bin/
COPY --from=builder /rust/target/x86_64-unknown-linux-musl/release/bc_rust_service ./
ENTRYPOINT ["/bin/bc_rust_service"]
EXPOSE  9099
