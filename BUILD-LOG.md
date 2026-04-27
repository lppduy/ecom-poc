# Build Log - E-Commerce POC

File nay dung de ghi daily log. Moi ngay ghi ngan gon nhung day du.

---

## Template (copy cho moi ngay)

### Date: YYYY-MM-DD

#### 1) Da lam duoc
- 
- 
- 

#### 2) Van de / blocker
- 
- 

#### 3) Cach xu ly
- 
- 

#### 4) Bai hoc system design hom nay
- 
- 

#### 5) Ke hoach ngay tiep theo
- 
- 

---

## Date: 2026-04-27

#### 1) Da lam duoc
- Tao bo tai lieu POC va roadmap tong.
- Tao plan 4 tuan de track tien do.
- Tao checklist ngay 1 theo session.
- Tao skeleton folders `services`, `infra`, `docs`, `scripts`.
- Tao 6 service Go voi `cmd/main.go`, `internal/`, endpoint `/health`.
- Tao `infra/docker-compose.yml` (postgres, redis, kafka, kafka-ui).
- Chuyen Kafka sang KRaft mode (khong can Zookeeper).
- Chay `docker compose up -d` thanh cong voi stack moi.
- Tao placeholders trong `docs/` cho architecture/sequence/tradeoffs/runbook.
- Implement `GET /products` trong `services/catalog`.
- Implement `POST /cart/items` trong `services/cart` (in-memory).
- Implement `POST /orders` trong `services/order` (tao order `PENDING`).
- Chay `go build` thanh cong cho 3 service catalog/cart/order.

#### 2) Van de / blocker
- Chua khoi tao `.env.example` tu dong do privacy hook chan file `.env*`.
- Chua verify ket noi tu service toi postgres/redis.
- `POST /orders` hien chua validate cart va chua luu postgres (dang mock de hoc flow).

#### 3) Cach xu ly
- Bat dau tu `DAY-1-CHECKLIST.md`, lam theo thu tu Session A -> D.
- Uu tien endpoint `POST /orders` tra `PENDING`.

#### 4) Bai hoc system design hom nay
- Co roadmap va scope guard se giam over-engineering.
- Nen uu tien critical path order/payment truoc cac tinh nang phu.

#### 5) Ke hoach ngay tiep theo
- Tao thu cong `.env.example` cho moi service.
- Implement API toi thieu: `GET /products`, `POST /cart/items`, `POST /orders`.
- Verify ket noi service -> postgres/redis.
- Nang `POST /orders`: validate cart + persist postgres.
