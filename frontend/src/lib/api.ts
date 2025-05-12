import { PUBLIC_API_ADDRESS } from '$env/static/public'
import { z } from 'zod'

export const API = PUBLIC_API_ADDRESS

////////////
// SCHEMA

const ErrCodeSchema = z.enum([
    "ACCOUNT_EMAIL_EXISTS", 
    "ACCOUNT_USERNAME_EXISTS", 
    "BAD_REGISTRATION_INFO", 
    "BAD_EMAIL", 
    "NO_ACCESS_TOKEN", 
    "NO_REFRESH_TOKEN", 
    "TOKEN_EXPIRED", 
    "TOKEN_INVALID", 
    "REFRESH_TOKEN_INVALIDATED", 
    "BAD_AUTH_HEADER", 
    "BAD_CLAIMS", 
    "PASSWORD_INCORRECT", 
    "NOT_FOUND", 
    "ILLEGAL_ARGUMENT", 
    "INTERNAL_ERROR", 
    "DATABASE_ERROR"
])
type ErrCode = z.infer<typeof ErrCodeSchema>

const RefreshResponseOKSchema = z.object({
    refresh: z.string().min(1),
    access: z.string().min(1),
})
type RefreshResponseOK = z.infer<typeof RefreshResponseOKSchema>

const ErrorResponseSchema = z.object({
    errcode: ErrCodeSchema,
})
type ErrorResponse = z.infer<typeof ErrorResponseSchema>

function getAccessLocal() {
    return localStorage.getItem("access")
}

function getRefreshLocal() {
    return localStorage.getItem("refresh")
}

function setAccessLocal(token: string) {
    localStorage.setItem("access", token)
}

function setRefreshLocal(token: string) {
    localStorage.setItem("refresh", token)
}

export async function refreshAccess(): Promise<boolean> {
    const currentRefresh = getRefreshLocal()
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
        const valid = RefreshResponseOKSchema.safeParse(data)
        if (valid.success) {
            setAccessLocal(valid.data.access)
            setRefreshLocal(valid.data.refresh)
            return true
        } else {
            const valid = ErrorResponseSchema.safeParse(data)
            if (valid.success) {
                console.log(valid.data.errcode)
            }
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
                "Authorization": `Bearer ${getAccessLocal()}`
            }
        })
    } catch(e) {

    }
}