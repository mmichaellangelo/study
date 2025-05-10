import { API } from "$lib/api";

export async function createAccount(email: string, username: string, password: string) {
    try {
        const res = await fetch(`${API}/accounts`, {
            method: "POST",
            credentials: "include"
        })
        if (!res.ok) {

        } else {
            
        }
    } catch (err) {
        if (err instanceof Error) {

        }
    }
}