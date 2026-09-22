FROM node:24-bookworm-slim

ENV CI=1 \
    MCODE_DISABLE_TELEMETRY=1 \
    DO_NOT_TRACK=1

WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@9.12.0 --activate
RUN apt-get update \
    && apt-get install --no-install-recommends --yes g++ make python3 \
    && rm -rf /var/lib/apt/lists/*

COPY --chown=node:node package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY --chown=node:node packages ./packages
COPY --chown=node:node third_party ./third_party
COPY --chown=node:node scripts ./scripts
COPY --chown=node:node test ./test
COPY --chown=node:node release ./release
COPY --chown=node:node tsconfig*.json vitest.oss.config.mjs ./

RUN chown -R node:node /workspace
USER node
RUN pnpm install --frozen-lockfile
RUN pnpm build

ENTRYPOINT ["node", "dist/cli.js"]
