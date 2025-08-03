export async function fetchCacheState() {
    const res = await fetch("http://localhost:8080/cache", { method: "GET" });
    if (!res.ok) throw new Error("Failed to fetch cache");
    return res.json();
}

export async function addToCache(key: string, value: any) {
    const res = await fetch("http://localhost:8080/add", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ key, value }),
    });
    if (!res.ok) throw new Error("Failed to add to cache");
    return;
}