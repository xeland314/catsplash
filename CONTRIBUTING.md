# Contributing to Catsplash

## Getting started
1. Fork and clone the repository.
2. Install the required dependencies: `sudo apt install hostapd dnsmasq iptables gcc make sqlite3 golang`.
3. Run the tests with `make test`.

## How to contribute
- Bugs: open an issue with steps to reproduce and the expected behavior.
- Features: open an issue first so the change can be discussed before implementation.
- Pull requests: branch from `main`, include tests when relevant, and describe the change clearly.

## Security guidelines
Catsplash handles personal data (MAC addresses, IP addresses) under LOPDP compliance. All contributors must follow these rules:

**Never:**
- Interpolate strings in `exec.Command` — always pass args as separate parameters
- Log personal data (MACs, IPs) in plaintext — use `maskMAC()` for MACs
- Accept MAC/IP as URL parameters for identity resolution — use cookie-based session tokens
- Skip input validation — always call `isValidMAC()` before using MAC addresses
- Commit secrets, passwords, or API keys to the repository

**Always:**
- Use `subtle.ConstantTimeCompare` for token/secret comparison
- Add tests for new endpoints that handle personal data
- Verify that CSRF protection is not broken by template changes
- Run `go vet ./...` and `go test ./...` before submitting

See `docs/security_guidelines.md` for the full security reference.

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
