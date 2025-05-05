interface UserState {
    ID: number
}

export const userState = $state<UserState>({ID: -1})

export function getUserState() {
    const res = fetch(`http://localhost:8080/me`)
}