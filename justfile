set shell := ["bash", "-euo", "pipefail", "-c"]

suite_root := justfile_directory()
bin_dir := suite_root / ".bin"
build_dir := suite_root / ".build"
harness := bin_dir / "agent-framework-suite"
rust_driver := bin_dir / "agent-framework-rs-driver"
python_driver := suite_root / "drivers/python/agent-framework-py-driver"

default:
    @just --list

prepare:
    mkdir -p "{{ bin_dir }}" "{{ build_dir }}"

build-suite: prepare
    go build -o "{{ harness }}" ./cmd/agent-framework-suite

build-rust-driver: prepare
    CARGO_TARGET_DIR="{{ build_dir }}/rust-driver" cargo build --release --manifest-path drivers/rust/Cargo.toml
    cp "{{ build_dir }}/rust-driver/release/agent-framework-rs-driver" "{{ rust_driver }}"

build: build-suite build-rust-driver

# Harness-only tests. These use fake drivers and never invoke either framework.
test:
    go test ./...

vet:
    go vet ./...

fmt:
    rg --files -g '*.go' cmd internal | xargs gofmt -w
    cargo fmt --manifest-path drivers/rust/Cargo.toml

fmt-check:
    test -z "$(rg --files -g '*.go' cmd internal | xargs gofmt -l)"
    cargo fmt --manifest-path drivers/rust/Cargo.toml -- --check

doctor *args: build
    "{{ harness }}" --python-driver "{{ python_driver }}" --rust-driver "{{ rust_driver }}" doctor {{ args }}

list *args: build-suite
    "{{ harness }}" list {{ args }}

run *args: build
    "{{ harness }}" --python-driver "{{ python_driver }}" --rust-driver "{{ rust_driver }}" run {{ args }}

parity: build
    "{{ harness }}" --python-driver "{{ python_driver }}" --rust-driver "{{ rust_driver }}" run --tag parity --stream

# Mandatory hermetic acceptance: both public-library drivers and every v1 case.
check: test vet fmt-check build
    "{{ harness }}" --python-driver "{{ python_driver }}" --rust-driver "{{ rust_driver }}" doctor
    "{{ harness }}" --python-driver "{{ python_driver }}" --rust-driver "{{ rust_driver }}" run --stream

clean:
    rm -rf "{{ bin_dir }}" "{{ build_dir }}" reports
