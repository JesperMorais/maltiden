const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080"

export async function checkHealth(): Promise<boolean> {
    try {
        const res = await fetch(`${API_URL}/health`)
        return res.ok
    } catch {
        return false
    }
}