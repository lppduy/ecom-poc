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
