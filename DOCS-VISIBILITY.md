# Documentation Visibility Policy

Keep this repository public-friendly while preserving private working notes locally.

## Public (safe to push)

- `README.md`
- `docs/`
- `services/`
- `infra/`
- `scripts/`

## Local only (do not push)

- `local/`
- `*.local.md`
- `.env*` (except `.env.example`)
- Any file containing secrets, private URLs, PII, or production dumps

## Pre-push checklist

- [ ] No `.env` or secret files are staged
- [ ] No tokens/passwords in code or docs
- [ ] No private internal URLs
- [ ] No personal/customer data
# Documentation Visibility Policy

Use this policy to keep the public repository clean while protecting local-only notes.

## Public (safe to push)

- `README.md`
- `docs/architecture.md`
- `docs/sequence-checkout.md`
- `docs/tradeoffs.md`
- `docs/runbook.md`
- Source code under `services/`, `infra/`, `scripts/`

## Local only (do not push)

- Learning journals and daily progress logs
- Team-only planning drafts
- Secrets, credentials, internal-only URLs
- Any personal or production data dumps

## Local-only location

- `local/`
- `*.local.md`

All local-only paths above are ignored by `.gitignore`.

## Pre-push checklist

- [ ] No `.env` or secret files are staged
- [ ] No tokens/passwords in code or docs
- [ ] No private internal URLs
- [ ] No customer or personal data
# Documentation Visibility Policy

Use this policy to keep the public repository clean while protecting local/internal notes.

## Public (safe to push)

- `README.md`
- `docs/architecture.md`
- `docs/sequence-checkout.md`
- `docs/tradeoffs.md`
- `docs/runbook.md`
- Source code under `services/`, `infra/`, `scripts/`

## Local only (do not push)

- Learning journals and daily progress logs
- Team-only notes and planning drafts
- Secrets, private endpoints, credentials, internal URLs
- Any personal data (PII) or production debug dumps

## Local-only locations

- `docs-private/`
- `notes-private/`
- `*.local.md`

All locations above are ignored by `.gitignore`.

## Pre-push checklist

- [ ] No `.env` or secret file included
- [ ] No tokens/passwords in code or docs
- [ ] No private internal URLs
- [ ] No customer or personal data
# Documentation Visibility Policy

Use this policy to keep the public repository clean while protecting local/internal notes.

## Public (safe to push)

- `README.md`
- `docs/architecture.md`
- `docs/sequence-checkout.md`
- `docs/tradeoffs.md`
- `docs/runbook.md`
- Source code under `services/`, `infra/`, `scripts/`

## Local only (do not push)

- Learning journals and daily progress logs
- Team-only notes and planning drafts
- Secrets, private endpoints, credentials, internal URLs
- Any personal data (PII) or production debug dumps

## Local-only locations

- `docs-private/`
- `notes-private/`
- `*.local.md`

All locations above are ignored by `.gitignore`.

## Pre-push checklist

- [ ] No `.env` or secret file included
- [ ] No tokens/passwords in code or docs
- [ ] No private internal URLs
- [ ] No customer or personal data
# Docs Visibility Policy

Tai lieu nay giup ban quyet dinh cai gi duoc push public, cai gi nen giu local.

## Public (nen push len GitHub)

- `README.md`
- `PLAN-4-WEEKS.md`
- `DAY-1-CHECKLIST.md`
- `BUILD-LOG.md` (chi giu ghi chu ky thuat, khong ghi secret/noi bo nhay cam)
- `docs/architecture.md`
- `docs/sequence-checkout.md`
- `docs/tradeoffs.md`
- `docs/runbook.md`
- Toan bo source code trong `services/`, `infra/`, `scripts/`

## Private (chi de tren may)

- Ghi chu thong tin nhay cam noi bo team/cong ty
- Token, key, webhook secret, endpoint private
- Tai lieu chua PII (du lieu nguoi dung that)
- Noi dung retrospective thuan noi bo

## Quy uoc luu private

- Dat trong `docs-private/` hoac `notes-private/`
- Dung hau to `*.local.md` cho file local
- Tat ca cac muc tren da duoc ignore trong `.gitignore`

## Checklist truoc khi push public

- [ ] Khong co file `.env` that
- [ ] Khong co token/API key trong code/doc
- [ ] Khong co URL private cua he thong that
- [ ] Khong co thong tin nhan than/khach hang that (PII)
