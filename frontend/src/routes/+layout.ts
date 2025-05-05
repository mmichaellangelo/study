import { userState } from "$lib/state/account.svelte";
import type { LayoutLoad } from "./$types";

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
            console.log(data)
            return {
                userID: data.userid,
                username: data.username,
            }
        } catch (e) {
            return {
                userID: -1
            }
        }
    }
}