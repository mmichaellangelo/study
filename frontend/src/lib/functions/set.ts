import { API } from "$lib/api"
import type { Card, Set } from "$lib/types/types"

export interface CardUpdate {
    id?: number // new if no id
    front: string
    back: string
}

export interface SetUpdate {
    name?: string
    cards?: CardUpdate[]
}

export async function UpdateSet(setID: number, updateData: SetUpdate): Promise<Set> {
    try {
        const res = await fetch(`${API}/sets/${setID}`, {
            method: "PATCH",
            credentials: "include",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(updateData)
        })
        if (!res.ok) {
            return Promise.reject(await res.text())
        }
        const data = await res.json() as Set
        return data
    } catch (e) {
        if (e instanceof Error) {
            return Promise.reject(e.message)
        }
        return Promise.reject("unknown")
    }
}