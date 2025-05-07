export async function createAccount(email: string, username: string, password: string) {
    try {
        const res = await fetch(`http://localhost:8080/accounts`, {
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