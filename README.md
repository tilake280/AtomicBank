# ⚛️ AtomicBank

`AtomicBank` is a modern, high-performance fintech backend engineered to handle financial data with absolute consistency, lightning speed, and zero compromise on security. 

While traditional banking systems rely on nightly batch processing and fragile balance-overwrite patterns, `AtomicBank` utilizes a real-time, event-ready architecture backed by an immutable ledger. Every monetary movement is strictly atomic—it either succeeds perfectly or fails without a trace.

### 🛠️ Core Architecture Highlights

* **Immutable Double-Entry Ledger:** Balance is computed dynamically from cryptographic-style debits and credits. Account balances are never directly updated, creating a permanent, bulletproof audit trail.
* **Race-Condition Immunity:** Utilizes row-level `FOR UPDATE` locking and strict database serialization to guarantee absolute protection against double-spending attacks.
* **Zero-Trust REST API:** Built using the high-performance `Gin` framework in Go. Features secure context-injection middleware to entirely neutralize Insecure Direct Object Reference (IDOR) exploits.
* **Context-Aware Lifecycle:** Leverages Go contexts to propagate cancellation signals. If a client drops their connection mid-request, the backend automatically aborts and rolls back the database transaction in-flight.

### 🧰 Tech Stack

* **Language:** Go (Golang) for high-concurrency and sub-millisecond execution times.
* **Database:** PostgreSQL / CockroachDB (Distributed, ACID-compliant SQL).
* **API Framework:** Gin Gonic.