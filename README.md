# LRU cache with TTL

This is a simple in-memory [LRU cache](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_Recently_Used_(LRU)) implemented in GO, complete with TTL expiration, observability and a visual demo.

## Why?

This is a portfolio project. Caching is essential for building web applications that scale, so I set out to implement one type of cache that would be useful in a real-world application. To make it easier to understand and demo, I also built a small visualizer.

### Design Musings

#### Architecture of a cache-heavy application

While an in-memory cache like this LRU implementation is a useful first layer for reducing latency and offloading repeated work, it is not sufficient on its own for applications handling heavy or distributed workloads.

In a production system, a multi-tiered caching architecture is often more appropriate. A typical setup might include:
- Local in-memory cache (like this one) for ultra-fast, process-local lookups
- Distributed cache (e.g. Memcached or Redis) as a shared layer across multiple app instances
- Persistent store (e.g. PostgreSQL, MongoDB) for source-of-truth data

This layered approach reduces database load, improves fault tolerance, and ensures horizontal scalability.

For even finer control, an additional layer can be added: a request-scoped cache, typically a lightweight, non-thread-safe in-memory cache that lives only during the lifecycle of a single request or function. This is useful for memoizing expensive computations or avoiding duplicate lookups in the same call stack.

#### Mutex

To make the LRU cache thread-safe, I wrapped its functions with a `sync.Mutex`. While it may seem like `sync.RWMutex` would be more appropriate since reads are typically more frequent, in this case, this approach offers minimal real benefit. Every Get() operation in the LRU cache modifies internal state by moving the accessed item to the front of the list to reflect recency. As a result, even read operations require a write lock.

Using RWMutex would therefore require either locking with a write lock for reads (negating the advantage), or duplicating logic to handle read-only access without reordering, which could lead to inconsistencies.

## Features
- ⚡ Thread-safe Go LRU cache
- ⏱️ Optional TTL support
- 📊 Prometheus metrics endpoint (/metrics)
- 🔍 Live cache state via /cache endpoint
- 🧩 Interactive frontend using React Flow
- 🎨 TailwindCSS + Shadcn styling
- 🔄 Drag-and-drop nodes to visualize recency ordering
- ➕ Add new entries via UI dialog

## How to Run

### Backend

```bash
cd visualizer/backend
go run main.go
```

### Frontend
```bash
cd visualizer/my-cache-ui
npm install
npm run dev
```
Make sure the backend is running at localhost:8080.

## Demo

You can drag nodes around or add new cache items via the visual interface. LRU eviction is reflected live.

![Demo](./assets/demo.gif)

## Tech Stack

- Go for backend
- React + Vite for frontend
- React Flow for graphs
- TailwindCSS for styles
- Prometheus for metrics