---
name: Bug report
about: Report a bug or unintended behavior in Catsplash or catsctl
title: ''
labels: ''
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is. (e.g., "Sessions do not expire after the timeout interval" or "catsctl crashes on 32-bit ARM").

**To Reproduce**
Steps to reproduce the behavior:
1. Run Catsplash with command / flags: `sudo ./bin/catsplash ...`
2. Authenticate a client using namespace or physical device by doing `...`
3. Check `catsctl status` or iptables rules.
4. See error: (Paste the unexpected behavior or crash log here).

**Expected behavior**
A clear and concise description of what you expected to happen.

**Environment & System Context (please complete the following information):**
- **Catsplash Version / Commit Hash:** [e.g., v0.1.0 or commit ec45b1]
- **OS / Linux Distribution:** [e.g., Debian 12, Alpine Linux]
- **Kernel Version:** [e.g., output of `uname -r`]
- **Architecture:** [e.g., amd64-v1, amd64-v2, armv7, 386]
- **Go Version (if building from source):** [e.g., go1.23]
- **Firewall Backend:** [e.g., iptables-legacy, iptables-nft]

**Logs & Command Output**
If applicable, please provide relevant outputs to speed up debugging:
<details>
<summary>Catsplash Daemon Logs</summary>

```bash
# Paste logs here

```

```bash
# Paste iptables output here

```

```bash
# Paste catsctl status output here

```

**Additional context**
Add any other context about the problem here (e.g., custom network namespace setups, hardware specifications of the router).
