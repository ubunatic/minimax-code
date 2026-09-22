# System observation and data boundaries

This is a source-level review of places where MiniMax Code can inspect the host,
workspace, browser, credentials, or network as of 2026-09-22. “Can inspect” does
not mean that every path runs automatically: most are tool capabilities invoked by
the model, the user, or an explicitly enabled feature.

## Verdict

The application is not a passive local chat client. Its local runtime is designed
to read and write the selected workspace, run subprocesses, inspect Git metadata,
read attachments, and optionally control a browser. Those capabilities are broad
but generally action-driven.

The highest-risk boundaries are:

1. shell/process execution;
2. arbitrary workspace/file reads and media extraction;
3. sending prompts, selected file contents, images, or tool results to the
   configured model/provider;
4. cloud/MCP tools, image reverse search, website deployment, and remote control;
5. browser DOM/screenshot/clipboard/file-upload access;
6. diagnostic and telemetry uploads.

The source does not show a general background “scan my whole home directory and
upload it” loop. It does show enough capability to do that if a model/tool call is
authorized and given a broad path or shell command. The effective security boundary
is therefore tool policy plus the host process permissions.

## Local filesystem and workspace inspection

The main implementation is `packages/local-runtime/src/files/api.ts`. It exposes
workspace file listing, search, reads, writes, previews/thumbnails, and media
handling. Related helpers include:

- `packages/agent-tools/src/desktop/local-rg-runner.ts` for local ripgrep searches;
- `packages/agent-tools/src/shared/read-*` for PDF, notebook, video, and other
  file extraction;
- `packages/local-runtime/src/files/worktrees.ts` and `managed-worktrees.ts` for
  Git root, worktree, status, reflog, and worktree operations;
- `packages/local-runtime/src/files/git-process.ts`, which runs `git` with the
  selected workspace as `cwd`;
- `packages/local-runtime/src/messages/user-media.ts` and image/thumbnail helpers,
  which read attachment bytes from local paths.

These paths can see more than source text: `.git` metadata, untracked files,
logs, images, documents, notebooks, media, and any file reachable by an allowed
path. Search helpers apply sensitive-file and root policies in several places, but
those are guardrails—not a guarantee that a user-authorized shell command or a
different tool cannot access the same data.

## Shell and subprocess execution

The TUI recognizes bash input in `packages/tui/src/tui/commands/bash-input.ts`.
The runtime and desktop tool contracts expose command execution and background
tasks; `packages/agent-tools/src/desktop/types.ts` documents the command/background
interfaces. Git and updater code also spawn local processes.

An authorized shell command can inspect environment variables, processes, network
configuration, home-directory files, SSH material, credential stores, and anything
the operating-system account can read. It can also modify or delete data unless the
host sandbox/policy prevents it. Treat shell access as full local-user authority,
not as a read-only project helper.

The canonical role policy in
`packages/agent-tools/src/desktop/canonical-tool-policy.ts` removes write,
delegation, memory, and computer-use tools from certain read-only subagent roles,
but the normal interactive runtime still has broader configured capabilities.

## Environment, identity, and local metadata

The process reads `process.env` for provider keys, region/build selection, data
directories, proxy settings, telemetry flags, and update behavior. Provider setup
also reads API-key environment variables in `packages/tui/src/cli/provider-command.ts`.
Data-directory and path helpers read the user home directory and configured
`MINIMAX_DATA_DIR`/related locations.

The TUI persists and reads MiniMax OAuth state, region preferences, account
identity, sessions, caches, and diagnostics under the configured data directory.
That is local state access, not by itself network disclosure, but credentials and
session files should be treated as sensitive and never placed inside a workspace
that might be sent to a model.

## Browser and desktop observation

`packages/browser-core/src/` provides a provider-neutral browser boundary. Its
contracts include DOM/semantic snapshots, JavaScript evaluation, screenshots,
keyboard/mouse actions, downloads, clipboard abstraction, and file upload. The
browser transport can use an attached Chrome DevTools target/WebSocket.

Important consequences:

- a browser-capable tool can inspect the current page and rendered content;
- page state may include logged-in web applications and private data;
- screenshots and DOM state can become model input;
- `cdp-file-upload.ts` can select local files for a browser page;
- the clipboard interface is abstracted and the in-memory implementation is safe,
  but a host provider may connect it to a real clipboard.

The source does not establish that browser access is always active. It is a
capability exposed by the runtime/host integration and should be disabled at the
host policy boundary when not needed.

## Network and model data flow

Even with MiniMax telemetry disabled, ordinary model use sends the prompt and any
context selected for the turn to the configured provider. That can include file
contents, command output, Git information, images, browser state, and tool
results. The provider is determined by the selected model/provider configuration;
local models avoid external model transport only when their endpoint is genuinely
local.

Explicit network-capable paths include:

- local web fetch in `packages/local-runtime/src/web-fetch/`;
- web search and Matrix/MCP tools in `packages/agent-tools/src/cloud/matrix-tools/`;
- image search/download and image reverse search, where local images may be sent
  to a remote service;
- website deployment/upload under `packages/local-runtime/src/website-deploy/`;
- channel adapters such as Telegram, Feishu, WeChat, or other configured hosts.

These are feature/tool paths, not evidence of an unconditional background upload.
They should nevertheless be considered data egress whenever enabled.

## Diagnostics, telemetry, and updates

Business telemetry is implemented in
`packages/tui/src/analytics/business-telemetry.ts`; diagnostic incident handling is
in `packages/tui/src/observability/incident-reporter.ts`. Both are opt-in through
config and blocked by `MCODE_DISABLE_TELEMETRY=1` or `DO_NOT_TRACK=1`, as documented
in [local-providers-without-minimax.md](local-providers-without-minimax.md).

The updater reads installation metadata, contacts release endpoints, downloads
artifacts, verifies them, and can spawn the package manager/runtime. This is not
workspace reconnaissance, but it is an automatic network/process boundary worth
including in a locked-down deployment. Inspect `packages/tui/src/update/` if
updates must be disabled or pinned.

## Hardening recommendations

For a local-model/private-workspace deployment:

1. Run with a dedicated OS user/container whose filesystem contains only the
   intended workspace and provider configuration.
2. Disable telemetry and diagnostics with both environment flags and config.
3. Remove or deny shell, browser, cloud/MCP, website-deploy, channel, and remote-
   control capabilities unless needed.
4. Use a local model endpoint and verify that no provider fallback points to a
   hosted API.
5. Block outbound network by default; allow-list only the local model endpoint
   and any explicitly required provider.
6. Keep OAuth/API credentials outside the workspace and outside model-readable
   paths; rotate any credential that may have been exposed to a prompt, command,
   log, or diagnostic bundle.
7. Review `packages/agent-tools/src/desktop/canonical-tool-policy.ts` and the host
   tool catalog when changing policy. Removing a UI command alone does not remove
   the underlying runtime capability.

## Bottom line

The app’s ordinary local inspection is intentional and broad: workspace files,
Git, attachments, subprocesses, and optionally browser state. It is not shown as
an unconditional background snooping mechanism. Privacy depends on restricting
the tool catalog and OS/network permissions, not only on turning off telemetry.
