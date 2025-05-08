import type { Set } from "$lib/types/types";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
    try {
        const res = await fetch(`http://localhost:8080/sets/${params.set_id}`, {
            method: "GET",
            credentials: "include"
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