import { API } from "$lib/api"
import { userState } from "$lib/state/account.svelte"
import type { LayoutLoad } from "./$types"

export const prerender = false

export const load: LayoutLoad = async ({fetch}) => {
    if (userState.ID == -1) {
        try {
            const res = await fetch(`${API}/me`,
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