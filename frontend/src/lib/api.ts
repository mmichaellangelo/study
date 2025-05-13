import { PUBLIC_API_ADDRESS } from '$env/static/public'
import { z } from 'zod'
import { jwtDecode } from 'jwt-decode'

export const API = PUBLIC_API_ADDRESS

////////////
// SCHEMA

// AppError codes
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
export type ErrCode = z.infer<typeof ErrCodeSchema>

// Schema for error response, should always include errcode
const ErrorResponseSchema = z.object({
    errcode: ErrCodeSchema,
})
type ErrorResponse = z.infer<typeof ErrorResponseSchema>

const RefreshResponseSchema = z.object({
    refresh: z.string().min(1),
    access: z.string().min(1),
})
type RefreshResponse = z.infer<typeof RefreshResponseSchema>

const ClaimsSchema = z.object({
    userid: z.number(),
    username: z.string(),
    exp: z.number(),
})
type Claims = z.infer<typeof ClaimsSchema>

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
        const data: unknown = await res.json()
        if (!res.ok) {
            const valid = ErrorResponseSchema.safeParse(data)
            if (valid.success) {
                console.log(valid.data.errcode)
            }
        }
        const valid = RefreshResponseSchema.safeParse(data)
        if (valid.success) {
            setAccessLocal(valid.data.access)
            setRefreshLocal(valid.data.refresh)
            return true
        }
        return false
    } catch(e) {
        return false
    }
}

export async function apiFetchWithRefresh(path: string, method: string, body: any): Promise<Response> {
    const res = await fetch(`${API}/${path}`, {
        method: method,
        headers: {
            "Authorization": `Bearer ${getAccessLocal()}`
        },
        body: JSON.stringify(body)
    }) 
    if (!res.ok) {
        const body = res.json() as unknown
        const parse = ErrorResponseSchema.safeParse(body)
        if (parse.success) {
            if (parse.data.errcode == "TOKEN_EXPIRED") {
                const refreshed = await refreshAccess()
                if (!refreshed) {
                    return res
                }
            } else {
                return res
            }
        }
    }
    return res
}

export function isAccessValid() {
    const access = getAccessLocal()
    if (!access || access == "") {
        return false
    }
    const d = jwtDecode(access) as unknown
    const c = ClaimsSchema.safeParse(d)
    if (c.success) {
        const timeDiff = Date.now() - c.data.exp
        if (timeDiff < 0) {
            return false
        }
        return true
    }
}