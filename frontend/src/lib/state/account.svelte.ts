interface UserState {
    ID: number
    Username: string
}

export const userState = $state<UserState>({ID: -1, Username: ""})
