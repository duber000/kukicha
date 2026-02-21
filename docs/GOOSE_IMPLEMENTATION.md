# Goose DevOps Implementation Strategy

This document outlines the architecture and execution strategy for deploying Goose as an autonomic DevOps actor using existing Model Context Protocol (MCP) servers and custom Kukicha (`ku-*`) binaries.

## 1. Security Posture: The "Read-Only Context, Scoped Write Execution" Model

To prevent catastrophic autonomous errors, the architecture strictly separates **Observability (Reads)** from **Mutations (Writes)**.

- **Reads (MCP Servers)**: All generic, wide-reaching information gathering is performed by community-standard MCP servers.
- **Service Accounts/Roles**: These MCP servers **must** be configured with strictly read-only credentials:
  - **k8s**: Retrieve cluster metadata, pod status, and logs using an `mcp-readonly` `ClusterRole`.
  - **Linux Servers**: Run the **Red Hat linux-mcp-server** under a tightly restricted, unprivileged `mcp-auditor` user account. It can read `/var/log`, `journalctl`, and `top`, but cannot execute `sudo` or modify config files.
  - **PostgreSQL**: Connect the **Google MCP Database Toolbox** using a database role limited to `CONNECT`, `SELECT`, and `pg_stat_activity` / `pg_stat_statements` views.
  - **Slurm**: Allow read-only execution of `sinfo`, `squeue`, and `sacct`.

By starving the generic MCP servers of write privileges, Goose cannot accidentally break infrastructure even if it hallucinates a destructive shell command or generic SQL `DROP` statement.

## 2. Integrating `ku-*` Binaries as the Mutation Layer

Goose reads contextual data via the read-only MCP servers. When Goose determines a fix is required, it **must** use the specialized Kukicha binaries (`ku-opts`, `ku-k8s`, `ku-pg`, etc.). These binaries form the "Write API." 

They parse strict arguments, perform input validation, handle exact state transitions natively using Kukicha's `onerr`, and prevent the AI from having arbitrary shell write access.

- **`ku-diag` (Remote Diagnostics)**: Goose uses Red Hat's Linux MCP to read a broken `/etc/nginx/nginx.conf`. To troubleshoot, Goose executes `ku-diag service-status --name nginx` or `ku-diag journal --priority err` to get structured diagnostic data.
- **`ku-k8s`**: Goose identifies a CrashLoopBackOff via MCP, and executes `ku-k8s get-unready-logs deployment_name` to safely extract log streams for analysis, without needing ArgoCD admin access.
- **`ku-slurm`**: Goose analyzes a `FAILED` job via read-only `sacct`, and executes `ku-slurm resubmit $JOB_ID --add-flags "--requeue --mem=32G"`.
- **`ku-pg`**: Goose spots a massive sequential scan via the Database Toolbox, and executes `ku-pg apply-index --table users --column email` (which Kukicha runs `CONCURRENTLY` and handles locking timeout errors).
- **`ku-sys`**: Used for structured systemd service queries (`ku-sys service-status nginx`) and scoped restarts (`ku-sys restart-service nginx`), outputting JSON status payloads for Goose analysis.

## 3. Proposal: A Safe "Ralph Loop" Scenario

### Scenario: The Auto-Tuning PostgreSQL Indexer

One of the safest, highest-ROI, autonomic processes you can hand over to Goose via a Ralph Loop is **Slow Query Analysis and Index Generation**. 

Because `CREATE INDEX CONCURRENTLY` in PostgreSQL does not block writes, it is a safe mutation that can instantly resolve application-level performance degradations.

**The Ralph Loop Architecture:**

1. **Initial Trigger**: A cron job or an alert fires off `goose run` because DB load has spiked.
2. **Worker AI (e.g., GPT-4o)**:
   - Uses the **Google MCP Database Toolbox** (read-only) to query `pg_stat_statements`.
   - Identifies the query causing the most DB time.
   - Uses the MCP to `EXPLAIN` the query and inspect table schemas to prove a missing index.
   - Proposes an index.
   - **Mutation**: Executes `ku-pg apply-index --table <name> --column <col>` (the Kukicha binary ensures the index is built concurrently and safely reports success/failure).
   - Writes the explanation to the Ralph Loop `work-summary.txt`.
3. **Reviewer AI (e.g., Claude 3.5 Sonnet)**:
   - Reads the task and the worker's summary.
   - Queries the **Google MCP Database Toolbox** to run the same `EXPLAIN` and validates that the query plan now utilizes the new index and the cost has dropped.
   - Checks `pg_stat_activity` to ensure there are no stuck advisory locks or abandoned index builds.
   - **Ship/Revise**: If the cost is reduced, it returns `SHIP`. If the index didn't help (or failed to build), it returns `REVISE`, telling the Worker to drop the useless index via `ku-pg drop-index` and try a composite index instead.

**Why this is safe:**
- Goose cannot drop tables or execute generic DML (`UPDATE`/`DELETE`) because the Google MCP Toolbox connection is read-only.
- Goose cannot execute arbitrary DDL because it has no raw database write access; it can only invoke `ku-pg apply-index` and `ku-pg drop-index`, which Kukicha severely restricts and validates.
- The Reviewer uses an independent model to verify actual performance improvements before declaring the loop resolved.
