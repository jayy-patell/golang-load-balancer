# ⚖️ Go Load Balancer

A lightweight load balancer built in Go, supporting multiple load balancing strategies, rate limiting mechanisms, dynamic backend management, and health checking.

---

## 🚀 Features

- 🔁 Load Balancing Algorithms:
  - Round Robin (`rr`)
  - Weighted Round Robin (`wrr`)
  - Least Connections (`lc`)
  - IP Hashing (`ip`)
  
- 🛡️ Rate Limiting:
  - Token Bucket
  - Leaky Bucket
  - Fixed Window
  - Per-client IP limiting (`X-Forwarded-For`)

- ⚙️ Runtime Features:
  - Health checking of backend servers
  - Dynamic add/remove of backends via admin endpoints

---

## 🛠️ Getting Started

### Clone the repository

```bash
git clone https://github.com/jayy-patell/golang-load-balancer.git
cd golang-load-balancer
```

---

## 🧪 Run the Load Balancer

### ✅ Without Rate Limiting

```bash
go run main.go -algo=rr -n=5                          # Round Robin with 5 backends
go run main.go --algo=wrr --n=3 --weights=5,1,1       # Weighted Round Robin
go run main.go --algo=lc --n=3                        # Least Connections
go run main.go --algo=ip --n=3                        # IP Hash
```

### 🚦 With Rate Limiting

```bash
go run main.go -algo=rr -n=3 -limiter=token -rate=10 -burst=5      # Token Bucket
go run main.go -algo=rr -n=3 -limiter=leaky -rate=8 -burst=3       # Leaky Bucket
go run main.go -algo=rr -n=3 -limiter=fixed -rate=5                # Fixed Window
```

### 🔐 Per-client IP Rate Limiting
This is useful for limiting requests from the same client IP. It prevents a single abusive client from leading to a denial of service for others.

```bash
go run main.go -algo=rr -n=3 -limiter=fixed -rate=2 -burst=2
```

---

## 🌐 Load Balancer Endpoint

```http
GET http://localhost:8090/loadbalancer
```

Each request is forwarded to a healthy backend based on the selected algorithm.

---

## 🔬 Testing Features

### Test Least Connections

```bash
for /L %i in (1,1,20) do start /B curl http://localhost:8090/loadbalancer
```
### Test with same client IP:

```bash
for /L %i in (1,1,6) do start /B curl -H "X-Forwarded-For: 1.2.3.4" http://localhost:8090/loadbalancer
```

---

## ⚙️ Admin API (Dynamic Backend Management)

### ➕ Add Backend

```bash
curl -X POST "http://localhost:8090/admin/addBackend?url=http://localhost:8083&weight=2"
```

### ➖ Remove Backend

```bash
curl -X POST "http://localhost:8090/admin/removeBackend?url=http://localhost:8082"
```

---

## ❤️ Health Checks

Each backend exposes:

```http
GET http://localhost:808X/health
```

---

## 📊 (Coming Soon)

- Prometheus Metrics & Monitoring: Request count per backend, Error rate, Rate limiter rejects, Live backend health metrics
- More load balancing algorithms and rate limiting strategies

---

## 👨‍💻 Author

Built with 💡 and Go by [@jayy-patell](https://github.com/jayy-patell)
Feel free to reach out for any questions or suggestions!
