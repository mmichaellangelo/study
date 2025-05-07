export interface Card {
    id: number,
    set_id: number,
    front?: string,
    back?: string,
    created: Date
}

export interface Set {
    id: number,
    account_id: number,
    name: string,
    description: string,
    created: Date,
    cards?: Card[],
}

export interface StatusMessageData {
    loading: boolean
    success: boolean
    message: string
}