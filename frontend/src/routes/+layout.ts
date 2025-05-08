import { userState } from "$lib/state/account.svelte"
import type { LayoutLoad } from "./$types"

export const prerender = false

export const load: LayoutLoad = async () => {
    if (userState.ID == -1) {
        try {
            const res = await fetch(`http://localhost:8080/me`,
                {
                    method: "GET",
                    credentials: "include",
                }
            )
            const data = await res.json()            
            userState.ID = data.userid
            userState.Username = data.username
            return {
                account: {
                    id: data.userid,
                    username: data.username
                }
            }
            
        } catch (e) {
            userState.ID = -1
            userState.Username = ""
        }
    }
}