import { API } from "$lib/api"
import type { PageLoad } from "./$types"

export const load: PageLoad = async ({fetch}) => {
    try {
        const res = await fetch(`http://${API}/sets`, {
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