# Contributing to Catsplash

## Getting started
1. Fork and clone the repository.
2. Install the required dependencies: `sudo apt install hostapd dnsmasq iptables gcc make sqlite3 golang`.
3. Run the tests with `make test`.

## How to contribute
- Bugs: open an issue with steps to reproduce and the expected behavior.
- Features: open an issue first so the change can be discussed before implementation.
- Pull requests: branch from `main`, include tests when relevant, and describe the change clearly.

## Code style
- Keep Go code idiomatic and gofmt-ed.
- Prefer small, focused changes over broad refactors.
- Use clear names and keep comments near the behavior they explain.

## Commit messages
Use the format: `tipo: descripción corta`

Types:
- `feat`: new functionality
- `fix`: bug fixes
- `docs`: documentation updates
- `refactor`: internal restructuring
- `test`: test-related changes

## Questions
Open an issue or contact christopher.villamarin@protonmail.com.
