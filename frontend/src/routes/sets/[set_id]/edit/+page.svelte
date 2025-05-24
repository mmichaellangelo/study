<script lang="ts">
    import { goto, invalidate } from "$app/navigation";
    import { API } from "$lib/api.js";
    import Loader from "$lib/components/Loader.svelte";
    import type { Card, Set } from "$lib/types/types";
    import { onMount } from "svelte";

    let {data} = $props()

    let setLocal = $state<Set|undefined>(undefined)
    let setRemote = $state<Set|undefined>(undefined)
    let isLoading = $state(true)

    let synced = $derived.by(() => {
        if (setLocal === undefined || setRemote === undefined) {
            return false
        }
        if (setLocal.name != setRemote.name) {
            return false
        }
        if (setLocal.cards?.length != setRemote.cards?.length) {
            return false
        }
        if (setLocal.cards && setRemote.cards) {
            const localCards = setLocal.cards
            const remoteCards = setRemote.cards
            for (let i = 0; i < localCards.length; i++) {
                if (localCards[i].front != remoteCards[i].front ||
                    localCards[i].back != remoteCards[i].back
                ) {
                    return false
                }
            }
        }
        return true
    })

    var blankCard: Card = {
        id: -1,
        set_id: -1,
        created: new Date(),
        front: "",
        back: ""
    }

    onMount(async () => {
        isLoading = false
        if (data.set) {
            setRemote = JSON.parse(JSON.stringify(data.set))
            setLocal = JSON.parse(JSON.stringify(data.set))
        }
    })

    let localNewCardIndex = $state(-1)

    function addCard() {
        if (setLocal) {
            const newCard: Card = {
                id: localNewCardIndex,
                front: "",
                back: "",
                created: new Date(),
                set_id: setLocal.id
            }
            if (setLocal.cards) {
                setLocal.cards.push(newCard)
            cardsToUpdate.push(newCard.id)
            } else {
                setLocal.cards = [newCard]
                cardsToUpdate.push(newCard.id)
            }
            // timeout pushes execution to after DOM is updated
            setTimeout(() => {
                const newCardElement = document.querySelector(`#card_${newCard.id}`);
                const frontInput = newCardElement?.querySelector<HTMLInputElement>(".front");
                frontInput?.focus();
                frontInput?.scrollIntoView()
            }, 0);
            localNewCardIndex--
        }
    }

    function debounce(func: () => void, delay: number): () => void {
        let timeoutId: NodeJS.Timeout | null = null;

        return function (): void {
            if (timeoutId !== null) {
                clearTimeout(timeoutId);
            }
            timeoutId = setTimeout(() => {
                func();
                timeoutId = null;
            }, delay);
        };
    }

    let cardsToUpdate = $state<number[]>([])
    let cardsToDelete = $state<number[]>([])
    let nameUpdate = $state(false)

    // Local card has been modified >> add it to cardsToUpdate if needed
    function updateCard(id: number) {
        if (setLocal && setRemote && setLocal.cards && setRemote.cards) {
            if (!cardsToUpdate.includes(id)) {
                const cardLocal = setLocal.cards.find(card => card.id == id)
                const cardRemote = setRemote.cards.find(card => card.id == id)
                if (!cardLocal) {
                    // BAD
                    console.log("VERY BAD card not found idk you're on your own")
                    return
                } else if (!cardRemote) {
                    // new card
                    cardsToUpdate.push(id)
                } else if (cardLocal.front == cardRemote.front &&
                    cardLocal.back == cardRemote.back) {
                        // card synced >> remove from update list
                        cardsToUpdate = cardsToUpdate.filter((i) => i !== id)
                } else {
                    // not synced
                    cardsToUpdate.push(id)
                }
            }
        }
    }

    function deleteCard(id: number) {
        if (setLocal && setRemote) {
            // delete from local and cardsToUpdate
            cardsToUpdate = cardsToUpdate.filter(cardID => cardID !== id)
            setLocal.cards = setLocal?.cards?.filter(card => card.id !== id)
            if (setRemote?.cards?.find(card => card.id == id)) {
                // Is in remote >> add to cardsToDelete
                if (!cardsToDelete.includes(id)) {
                    cardsToDelete.push(id)
                }
            }
        }
        
    }

    function updateName() {
        nameUpdate = true
    }

    interface CardCreateRequest {
        id: number
        front: string
        back: string
    }

    interface CardCreateResponse {
        old_id: number
        new_id: number
        front: string
        back: string
    }

    interface CardUpdateRequest {
        id: number
        front?: string
        back?: string
    }

    interface CardUpdateResponse {
        id: number
        front: string
        back: string
    }

    interface SetUpdateRequest {
        name?: string
        description?: string
        cards_created?: CardCreateRequest[]
        cards_updated?: CardUpdateRequest[]
        cards_deleted?: number[]
    }

    interface SetUpdateResponse {
        name?: string
        description?: string
        cards_created?: CardCreateResponse[]
        cards_updated?: CardUpdateResponse[]
        cards_deleted?: number[]
    }
    
    /**
     * Determines what set data has changed, sends relevant information to the API,
     * updates setRemote with new changes as reported by the API,
     * updates id's of any newly created cards in setLocal
     */
    async function update() {
        if (!setLocal || !setRemote) {
            return
        }
        var u: SetUpdateRequest = {}
        if (nameUpdate) {
            // add name to update
            u.name = setLocal.name
        }
        if (cardsToUpdate.length !== 0) {
            // Add cards to create or update
            for (const cardID of cardsToUpdate) {
                if (setLocal.cards) {
                    const cardLocal = setLocal.cards.find((card) => card.id == cardID)
                    if (cardLocal && cardLocal.id < 0) {
                        // Create card
                        const newCard: CardCreateRequest = {
                            id: cardLocal.id,
                            front: cardLocal.front || "",
                            back: cardLocal.back || ""
                        }
                        if (!u.cards_created) {
                            u.cards_created = []
                        }
                        u.cards_created.push(newCard)
                    } else {
                        // Update card
                        if (cardLocal) {
                            if (!u.cards_updated) {
                                u.cards_updated = []
                            }
                            u.cards_updated.push({
                                id: cardLocal.id,
                                front: cardLocal.front || "",
                                back: cardLocal.back || ""
                            })
                        } 
                    }
                }
            } 
        }
        // Add cards to delete
        if (cardsToDelete.length !== 0) {
            for (const id of cardsToDelete) {
                if (!u.cards_deleted) {
                    u.cards_deleted = []
                }
                u.cards_deleted.push(id)
            }
        }
        // If anything needs to be updated, send the request
        if (u.name || u.description || u.cards_created || u.cards_updated || u.cards_deleted) {
            try {
                isLoading = true
                const res = await fetch(`${API}/sets/${setRemote?.id}`, {
                    method: "PATCH",
                    credentials: "include",
                    body: JSON.stringify(u)
                })
                if (!res.ok) {
                    console.log(await res.text())
                    isLoading = false
                    return
                }
                // Invalidate cached set data, forcing a reload when returning to set page
                await invalidate((url) => url.pathname === `/sets/${data.set?.id}`)
                const updateRes = await res.json() as SetUpdateResponse
                // If name returned, update setRemote's name
                if (updateRes.name) {
                    setRemote.name = updateRes.name
                }
                // If description returned, update setRemote's description
                if (updateRes.description) {
                    setRemote.description = updateRes.description
                }
                // If cards have been created, add them to setRemote, 
                // update the id's of the cards in setLocal,
                // and remove the temporary id's from cardsToUpdate
                if (updateRes.cards_created) {
                    // Update card's id in setLocal
                    updateRes.cards_created.forEach((cardRes) => {
                        var cardLocal = setLocal?.cards?.find(cardLocal => cardLocal.id == cardRes.old_id)
                        if (cardLocal) {
                            const oldID = cardLocal.id
                            cardLocal.id = cardRes.new_id
                            // Remove temporary id from cardsToUpdate
                            cardsToUpdate = cardsToUpdate.filter((id) => id !== oldID)
                            // If not synced, add new id to cardsToUpdate
                            if (cardLocal.front !== cardRes.front || cardLocal.back !== cardRes.back || cardLocal.id !== cardRes.new_id) {
                                cardsToUpdate.push(cardLocal.id)
                            }
                        }
                        // Add new card to setRemote TODO date, set_id sent from API
                        setRemote?.cards?.push({
                            id: cardRes.new_id,
                            front: cardRes.front,
                            back: cardRes.back,
                            created: new Date(),
                            set_id: data?.set?.id || 0
                        })
                    })
                }
                // Flag name for update only if the local copy does not match the new remote copy
                nameUpdate = !(setLocal.name == setRemote.name)
                // TODO descriptionUpdate
                
                // Update cards in setRemote to match server response
                u.cards_updated?.forEach((c) => {
                    const cardRemote = setRemote?.cards?.find((card) => card.id === c.id)
                    if (cardRemote) {
                        cardRemote.front = c.front
                        cardRemote.back = c.back
                    }
                    const cardLocal = setLocal?.cards?.find((card) => card.id === c.id)
                    if (cardLocal?.front === cardRemote?.front &&
                        cardLocal?.back === cardRemote?.back) {
                            cardsToUpdate = cardsToUpdate.filter((id) => id !== c.id)
                    }
                })
                // Remove any deleted cards from setRemote to match server response
                u.cards_deleted?.forEach((id) => {
                    if (setRemote) {
                        setRemote.cards = setRemote?.cards?.filter((card) => card.id !== id)   
                        cardsToDelete = cardsToDelete.filter((i) => i !== id)
                    }
                    
                })
                isLoading = false
            } catch (e) {
                console.log(e)
                isLoading = false
            }
        }
    }

    const debouncedUpdate = debounce(update, 900)

    onMount(() => {
        setInterval(debouncedUpdate, 1000)
    })

    let dialogElement = $state<HTMLDialogElement>()

    function showDialog() {
        dialogElement?.showModal()
    }

    function closeDialog() {
        dialogElement?.close()
    }

    async function deleteSet() {
        try {
            const res = await fetch(`${API}/sets/${data.set?.id}`, {
                method: "DELETE",
                credentials: "include",
            })
            if (!res.ok) {
                console.log(await res.text())
                return
            }
            await invalidate((url) => url.pathname === `/study`)
            goto("/study")
        } catch (e) {
            console.log(e)
        }
    }

</script>

{#if data.set}
    <dialog bind:this={dialogElement}>
        <p>are you sure you want to delete this set? this can't be undone!</p>
        <div>
            <button onclick={deleteSet} class="delete">yes</button>
            <button onclick={closeDialog}>no</button>
        </div>
    </dialog>
    <div id="back_delete_container">
        <a href={`/sets/${data.set.id}`}>back to set</a>
        <button class="delete" onclick={showDialog}>delete set</button>
    </div>
    <div id="title">
        <h2>{setLocal?.name}</h2>
        {#if isLoading}
            <Loader />
        {/if}
        {#if synced}
            <span>synced</span>
        {/if}
    </div>

    <div id="create_frame">
        {#if setLocal}
        <form>
            <label>
                name <br />
                <input type="text" name="name" placeholder="name" bind:value={setLocal.name} oninput={updateName}>
            </label>
            
            <br />
            {#if setLocal.cards}
                <label for="card">cards</label>
                {#each setLocal.cards as card, index}
                <div class="card" id={`card_${card.id}`} role="listitem">
                        <div contenteditable="true" class="front" placeholder="front" bind:innerText={card.front} oninput={() => updateCard(card.id)}></div>
                        <div contenteditable="true" class="back" placeholder="back" bind:innerText={card.back} oninput={() => updateCard(card.id)}></div>
                        <button class="delete" onclick={(e) => {e.preventDefault(); deleteCard(card.id);}}>del</button>
                </div>
                {/each}
            {/if}
            <button onclick={addCard} id="add_card_button">new card</button>
        </form>
        {/if}
    </div>
{:else}
    {#if data.error}
        <p>there was en error loading the set: {data.error}</p>
    {:else}
        <Loader />
    {/if}
{/if}

<style>
    dialog {
        position: absolute;
        background-color: var(--col-darkblue);
        backdrop-filter: blur(5px);
        color: var(--col-lightpink);
        border: 2px solid var(--col-msg-error);
        border-radius: 1rem;
        box-shadow: 0px 0px 10px var(--col-msg-error);
        font-weight: 600;
    }

    dialog>p {
        color: var(--col-msg-error);
    }

    dialog>div>button {
        margin-left: 1rem;
    }

    ::backdrop {
        backdrop-filter: blur(7px);
        background-color: rgba(0,0,0,0.6);
    }

    #title {
        display: flex;
        flex-direction: row;
        align-items: center;
    }

    #title>h2 {
        margin-right: 1rem;
    }

    #add_card_button {
        display: block;
        margin-top: 0.5rem;
    }

    #back_delete_container {
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        max-width: 27rem;
    }

    .card {
        display: flex;
        flex-direction: row;
        align-items: start;
    }

    #create_frame {
        display: block;
        width: 100%;
    }

    .front, .back {
        height: fit-content;
        margin-bottom: 1rem;
    }

    .front {
        width: 40%;
        max-width: 25rem;
    }

    .back {
        width: 50%;
        max-width: 40rem;
    }
</style>