import { API } from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params, fetch }) => {
    try {
        const response = await fetch(`${API}/accounts/${params.account_id}`, {
            method: "GET",
            credentials: "include",
        })
        if (!response.ok) {
            return
        }
        const data = await response.json()
        const createdObj = new Date(data.created)
        return {
            account: {
                id: data.id,
                username: data.username,
                email: data.email,
                created: createdObj
            }
        }
    } catch (e) {
        return
    }
}