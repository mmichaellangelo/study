import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
    try {
        const response = await fetch(`http://localhost:8080/accounts/${params.account_id}`, {
            method: "GET",
            credentials: "include",
        })
        if (!response.ok) {
            return
        }
        const data = await response.json()
        console.log(data)
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