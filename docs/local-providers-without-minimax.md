# Using local providers without MiniMax login or telemetry

This report traces the public CLI/TUI code as of 2026-09-22. It describes what is
conditional, what is still reachable from an explicitly selected MiniMax feature,
and the smallest configuration or code changes for a local-provider-only build.

## Short answer

The shipped CLI now applies both opt-outs at startup. They are also useful when
embedding the runtime or launching another entry point:

```sh
export MCODE_DISABLE_TELEMETRY=1
export DO_NOT_TRACK=1
```

The CLI preserves an explicit environment override. To intentionally opt in,
unset both variables before launch and enable the telemetry channels in config.

Use a provider whose configuration resolves to `authMode: api-key` or
`authMode: oauth`, not `authMode: managed-login`. Do not select MiniMax Token
Plan/managed models, `/checkin`, account management, feedback submission, or
other MiniMax-managed features. BYOK is explicitly supported without a MiniMax
login; the README's provider examples use `custom_provider` and say that
`minimax_api` is reserved for the official API.

The environment flags are the strongest operational switch: they prevent both
business telemetry construction and diagnostic uploads, even if config enables a
channel. Confirm with `mcode telemetry status`; it should report `enabled: false`
and identify `MCODE_DISABLE_TELEMETRY` or `DO_NOT_TRACK`.

## MiniMax login and OAuth

The shared OAuth implementation is in `packages/oauth-core/src/`. The TUI creates
it during startup in `packages/tui/src/tui/launcher.ts`, using endpoints resolved
by `packages/oauth-core/src/endpoint-config.ts`. The actual device authorization,
token, refresh, and revoke requests are in `oauth-client.ts`; these are the
MiniMax account endpoints, not Claude/Codex/local-model authentication.

The user-facing `mcode login` path is
`packages/tui/src/cli/auth-command.ts` -> `packages/tui/src/auth/application.ts`.
It starts the shared device flow, opens or prints the authorization URL, and
records `login_click`/`login_result` only when a telemetry client was supplied.
The interactive `/login` command uses the same application.

Login is not required merely to start a BYOK conversation. The gate is
`packages/tui/src/application/login-gate.ts`: it requires login when the account
is `managed-login`, or when the selected model is Token Plan, and otherwise lets
the selected BYOK route proceed. The runtime adapter can synchronize the shared
MiniMax lease in `packages/tui/src/runtime/adapter.ts`; this is the path that can
refresh a MiniMax token when a managed route is selected.

The practical forced-login switch is therefore provider selection/configuration:

- Keep local/custom models on `api-key` (or the provider's documented OAuth mode).
- Keep Claude/Codex on their own provider OAuth implementations.
- Do not mark a local/custom provider `managed-login` and do not select a
  `token-plan` model.
- If a feature still demands login, inspect its caller of
  `requireTuiInteractiveAgentAccess` or `requireTuiInteractiveAccountLogin`.
  Chat uses the former; feedback uses the latter. Removing those gates is a
  product/code change and will disable the corresponding MiniMax capability, not
  just the account prompt.

## Telemetry and callbacks to MiniMax

### Business/usage telemetry

`packages/tui/src/analytics/business-telemetry.ts` defines the event set, payload,
and POST endpoint. It includes launch, chat, command, session-lifecycle, and
MiniMax login/logout events. The endpoint is a MiniMax/Hailuo/BigData reporter
depending on region and build environment.

The TUI constructs this client in `packages/tui/src/tui/launcher.ts`, but only when
`config.telemetry.enabled === true` and neither `MCODE_DISABLE_TELEMETRY` nor
`DO_NOT_TRACK` is enabled. `packages/config/src/telemetry-policy.ts` applies the
same fail-closed rule to telemetry channels. The config fields are
`telemetry.enabled`, `telemetry.metrics`, and `telemetry.diagnostics`; an explicit
false value is the safe persistent setting.

### Diagnostics / incident reporting

`packages/tui/src/observability/incident-reporter.ts` stores bounded local incident
records and can upload batches to the MiniMax matrix observability endpoint. It
requires the diagnostics channel to be explicitly enabled, and the same two
environment opt-outs take precedence. It may resolve a MiniMax access token and
real user ID when an upload is authorized. With both environment flags above set,
this upload path is disabled as well.

### Account-dependent callbacks

`packages/tui/src/checkin/http-gateway.ts` calls MiniMax daily check-in status/claim
endpoints and requires a MiniMax access token plus real user ID. Feedback is gated
by `requireTuiInteractiveAccountLogin` and uses the runtime's MiniMax identity.
These are not background telemetry, but they are MiniMax callbacks and should be
avoided or removed from a strict local-only product surface.

## Recommended local-only policy

1. Set `MCODE_DISABLE_TELEMETRY=1` in the launcher/service environment, not only
   in an interactive shell.
2. Set `telemetry.enabled: false`, `telemetry.metrics: false`, and
   `telemetry.diagnostics: false` in the active config as defense in depth.
3. Configure and select only BYOK/OAuth/local providers. Verify the resolved
   account/provider inspection says `api-key` or `oauth`, never `managed-login`.
4. Avoid `/login`, `/checkin`, feedback, Token Plan, managed tools, and MiniMax
   account commands.
5. Run `mcode telemetry status` and inspect network traffic if the deployment
   must enforce a hard boundary. A network deny rule for MiniMax/Hailuo domains is
   an additional defense, but may also disable updates or other explicitly used
   features.

## If a code-level guarantee is required

The narrowest change is in `createConfiguredTuiBusinessTelemetry` in
`packages/tui/src/tui/launcher.ts`: return `undefined` unconditionally for a
local-only build, and omit the incident reporter construction or inject the noop
reporter. Preserve the existing login gates so MiniMax-only features fail clearly;
instead, make the local-only provider configuration impossible to resolve as
`managed-login`. If the requirement is that those features must never even attempt
MiniMax auth, add an explicit local-only capability flag at the runtime boundary
and make `resolveAccessTokenLease`, account synchronization, check-in, and feedback
reject before creating or refreshing the shared OAuth session.

Do not replace the shared OAuth endpoints with local-provider endpoints: Claude,
Codex, and local models have separate provider implementations and credentials.
