import type { PageLoad } from "./$types"

export const load: PageLoad = async () => {
    try {
        const res = await fetch(`http://localhost:8080/sets`, {
            method: "GET",
            credentials: "include",
        })
        if (!res.ok) {
            return { sets: null }
        }
        return {
            sets: await res.json()
        }
    } catch (e) {
        console.log(e)
    }
}