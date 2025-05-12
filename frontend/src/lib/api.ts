import { PUBLIC_API_ADDRESS } from '$env/static/public'
import { z } from 'zod'

export const API = PUBLIC_API_ADDRESS

////////////
// SCHEMA

const ErrorType = z.enum([""])

const RefreshResponseOKSchema = z.object({
    refresh: z.string().min(1),
    access: z.string().min(1),
})
type RefreshResponseOK = z.infer<typeof RefreshResponseOKSchema>

const RefreshResponseErrorSchema = z.object({
    error: z.object({
        type: z.string().,
        message: z.string()
    })
})
type RefreshResponseError = z.infer<typeof RefreshResponseErrorSchema>

function getAccessToken() {
    return localStorage.getItem("access")
}

function getRefreshToken() {
    return localStorage.getItem("refresh")
}

function setAccessToken(token: string) {
    localStorage.setItem("access", token)
}

function setRefreshToken(token: string) {
    localStorage.setItem("refresh", token)
}

export async function refreshAccess(): Promise<boolean> {
    const currentRefresh = getRefreshToken()
    try {
        const res = await fetch(`${API}/auth/refresh`, {
            method: "POST",
            headers: {
                "Authorization": `Bearer ${currentRefresh}`
            }
        })
        if (!res.ok) {
            return false
        }
        const data: unknown = await res.json()
        const valid = RefreshResponseSchema.safeParse(data)
        if (valid.success) {
            setAccessToken(valid.data.access)
            setRefreshToken(valid.data.refresh)
            return true
        }
        return false
    } catch(e) {
        return false
    }
}

export async function apiFetch(path: string, method: string) {
    try {
        await fetch(`${API}/${path}`, {
            method: method,
            headers: {
                "Authorization": `Bearer ${"<<TOKEN>>"}`
            }
        })
    } catch(e) {

    }
}