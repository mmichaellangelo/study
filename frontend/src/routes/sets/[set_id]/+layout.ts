import { API } from "$lib/api";
import type { Set } from "$lib/types/types";
import type { LayoutLoad } from "./$types";

export const load: LayoutLoad = async ({ params, fetch }) => {
    try {
        const res = await fetch(`http://${API}/sets/${params.set_id}`, {
            method: "GET",
            credentials: "include",
            
        })
        if (!res.ok) {
            return { error: await res.text() }
        }
        const data = await res.text()
        const set: Set = JSON.parse(data)
        return { set: set}
    } catch (e) {
        return { error: "unknown error" }
    }
}