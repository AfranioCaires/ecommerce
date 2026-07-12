# Test Maker Skill

Guidelines for Go unit and integration tests:

1. Group scenarios using `t.Run("it should ...", func(t *testing.T)...)`.
2. Keep `sut` (System Under Test) scoped locally inside the test functions.
3. Use fast in-memory mocks/adapters for the repositories and databases rather than calling live systems.
4. Ensure all variables use complete, descriptive names (no abbreviations like `ctx`, `c`, `err`, `u`).
