# Runtime Cutover Decision

The production composition root now creates repository-backed bounded-context services before registering HTTP routes. HTTP adapters receive application services rather than persistence repositories, keeping repository construction and PostgreSQL details out of route registration while preserving the existing API surface.
