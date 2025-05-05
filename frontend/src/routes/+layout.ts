import { userState } from "$lib/state/account.svelte";
import type { LayoutLoad } from "./$types";

export const load: LayoutLoad = async () => {
    if (userState.ID == -1) {
        try {
            console.log("fetching")
            const res = await fetch(`http://localhost:8080/me`,
                {
                    method: "GET",
                    credentials: "include",
                }
            )
            console.log(res)
            const data = await res.json()
            const userID = data.userID
            console.log(userID)
            return {
                userID: userID
            }
        } catch (e) {
            return {
                userID: -1
            }
        }
    }
}