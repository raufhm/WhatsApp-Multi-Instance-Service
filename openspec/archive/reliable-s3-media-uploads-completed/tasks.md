# Tasks

- [x] Define the upload-job domain model and repository interface.
- [x] Add the PostgreSQL migration for upload jobs, indexes, status values, and retry metadata.
- [x] Implement tenant-scoped job creation, claiming, completion, retry, and permanent-failure updates.
- [x] Refactor S3 upload handling to return and persist failures instead of discarding them.
- [x] Implement the bounded-backoff worker and graceful shutdown.
- [x] Add configuration and structured logs/metrics for the worker lifecycle.
- [x] Add unit tests for retry classification, backoff, idempotent keys, and terminal failures.
- [x] Add integration coverage for concurrent claims and recovery after a process restart.
- [x] Update the README roadmap and deployment/configuration documentation.
