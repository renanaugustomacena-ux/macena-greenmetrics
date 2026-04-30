# syntax=docker/dockerfile:1.7
#
# Doctrine refs: Rule 53 (supply chain), Rule 19 (security as structural), Rule 42 (resource lifecycle).
# Base image policy: digest-pin via Renovate (`renovate.json`), refreshed weekly.
# Until Renovate migration lands, the floating `node:20-alpine` tag is documented
# under RISK-002 (Alpine rolling apk) in `docs/RISK-REGISTER.md` with a Trivy
# image-scan diff alert as compensating control.
# TODO renan: replace each `FROM node:20-alpine` with `FROM node:20-alpine@sha256:<digest>`
# once Renovate first-run PR completes.

FROM node:25-alpine AS deps
WORKDIR /app
COPY package.json ./
# package-lock.json is optional in template; developers run `npm install` locally.
RUN npm install --no-audit --no-fund

FROM node:25-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:25-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup -S greenmetrics && adduser -S greenmetrics -G greenmetrics
COPY --from=builder --chown=greenmetrics:greenmetrics /app/build ./build
COPY --from=builder --chown=greenmetrics:greenmetrics /app/node_modules ./node_modules
COPY --from=builder --chown=greenmetrics:greenmetrics /app/package.json ./package.json

USER greenmetrics
EXPOSE 3000
ENV PORT=3000 HOST=0.0.0.0

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --quiet --spider http://127.0.0.1:3000 || exit 1

CMD ["node", "build"]
