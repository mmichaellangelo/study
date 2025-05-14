import { page } from "$app/state"
import { API } from "$lib/api"
import type { Set } from "$lib/types/types"
import type { PageLoad } from "./$types"

export const load: PageLoad = async ({fetch}) => {
    try {
        const res = await fetch(`${API}/sets`, {
            method: "GET",
            credentials: "include",
        })
        if (!res.ok) {
            return { sets: null }
        }
        return {
            sets: await res.json() as Set[]
        }
    } catch (e) {
        console.log(e)
    }
}