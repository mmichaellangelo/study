export interface Card {
    front: string,
    back: string,
}

export interface Set {
    name: string,
    cards: Card[],
}

export interface StatusMessageData {
    loading: boolean
    success: boolean
    message: string
}