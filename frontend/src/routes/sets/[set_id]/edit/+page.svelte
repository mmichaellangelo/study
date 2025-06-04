<svelte:head>
    <title>disco - edit: {data.set?.name}</title>
</svelte:head>

<script lang="ts">
    import { goto, invalidate } from "$app/navigation";
    import { API } from "$lib/api.js";
    import Loader from "$lib/components/Loader.svelte";
    import type { Card, Set } from "$lib/types/types";
    import { onMount } from "svelte";

    let {data} = $props()

    /** Local copy of set */
    let setLocal = $state<Set|undefined>()
    /** Most recent copy of the set as it exists on the server */
    let setRemote = $state<Set|undefined>(undefined)
    /** If true, show loading spinner */
    let isLoading = $state(true)
    /** Set true to prevent update() from running */
    let isUpdating = $state(false)

    /** Keeps track of sync status; true if setLocal === setRemote, otherwise false */
    let synced = $derived.by(() => {
        if (setLocal === undefined || setRemote === undefined) {
            return false
        }
        // Not synced if names are not equal
        if (setLocal.name !== setRemote.name) {
            return false
        }
        // Not synced if don't have the same number of cards
        if (setLocal.cards?.length !== setRemote.cards?.length) {
            return false
        }
        if (setLocal.cards && setRemote.cards) {
            const localCards = setLocal.cards
            const remoteCards = setRemote.cards
            for (let i = 0; i < localCards.length; i++) {
                // Check front and back of each card to ensure they are synced
                if (localCards[i].front !== remoteCards[i].front ||
                    localCards[i].back !== remoteCards[i].back
                ) {
                    return false
                }
            }
        }
        return true
    })

    onMount(async () => {
        // After set loaded, hide spinner
        isLoading = false
        // If sets loaded, initialize setRemote and setLocal
        if (data.set) {
            setRemote = JSON.parse(JSON.stringify(data.set))
            setLocal = JSON.parse(JSON.stringify(data.set))
        }
    })

    /** Index to keep track of new card element id's */
    let localNewCardIndex = $state(-1)

    /**
     * Adds a new card to local set, adds to update queue,
     * focuses front input element of newly created card
     */
    function addCard() {
        if (setLocal) {
            // New Card object to add to setLocal
            const newCard: Card = {
                id: localNewCardIndex,
                front: "",
                back: "",
                created: new Date(),
                set_id: setLocal.id
            }
            if (setLocal.cards) {
                // Add empty card to setLocal
                setLocal.cards.push(newCard)
            cardsToUpdate.push(newCard.id)
            } else {
                setLocal.cards = [newCard]
                cardsToUpdate.push(newCard.id)
            }
            // Timeout pushes execution to after DOM is updated
            setTimeout(() => {
                // Focus and scroll to front input of newly created card
                const newCardElement = document.querySelector(`#card_${newCard.id}`);
                const frontInput = newCardElement?.querySelector<HTMLInputElement>(".front");
                frontInput?.focus();
                frontInput?.scrollIntoView()
            }, 0);
            // Decrement localNewCardIndex so the next card created will have a unique index
            localNewCardIndex--
        }
    }

    /**
     * Returns a debounced function that can only run every (delay) ms
     * @param func function to debounce
     * @param delay debounce delay in ms
     * @return debounced function
     */
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

    /** List of cards to create or update to be sent to API */
    let cardsToUpdate = $state<number[]>([])
    /** List of cards to delete to be sent to API */
    let cardsToDelete = $state<number[]>([])
    /** If true, updated name must be sent to API */
    let nameUpdate = $state(false)

    /**
     * Handles card updates, adds to cardsToUpdate if the card has been
     * changed from the copy in setRemote
     * @param id id of local card to update
     */
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

    /**
     * Delete a card locally and add id to deletion queue 
     * if it exists on the server
     * @param id id of the card to delete
     */
    function deleteCard(id: number) {
        if (setLocal && setRemote) {
            // Delete from local and cardsToUpdate
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

    /**
     * Sets nameUpdate to true, signalling the system that 
     * the new set name should be sent to the API
     */
    function updateName() {
        nameUpdate = true
    }

    /**
     * Data structure for sending a card creation request
     * to the API
     */
    interface CardCreateRequest {
        id: number
        front: string
        back: string
    }

    /**
     * Data structure for a successful card creation response
     * from the API
     */
    interface CardCreateResponse {
        old_id: number
        new_id: number
        front: string
        back: string
    }

    /**
     * Data structure for sending a card update request
     * to the API
     */
    interface CardUpdateRequest {
        id: number
        front?: string
        back?: string
    }

    /**
     * Data structure for a successful card update response
     * from the API
     */
    interface CardUpdateResponse {
        id: number
        front: string
        back: string
    }

    /**
     * Data structure for a set update request
     * to the API
     */
    interface SetUpdateRequest {
        name?: string
        description?: string
        cards_created?: CardCreateRequest[]
        cards_updated?: CardUpdateRequest[]
        cards_deleted?: number[]
    }

    /**
     * Data structure for a successful set update response
     * from the API
     */
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
        if (isUpdating) {
            return
        }
        isUpdating = true
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
                        if (setRemote && !setRemote.cards) {
                            setRemote.cards = []
                        }
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
            } catch (e) {
                console.log(e)
            } finally {
                isUpdating = false
                isLoading = false
            }
        }
        isUpdating = false
        isLoading = false
    }

    /** Update function debounced to only run at quickest every 900 ms */
    const debouncedUpdate = debounce(update, 900)

    /** On load, run debounced update function every second to keep the set data synced */
    onMount(() => {
        setInterval(debouncedUpdate, 1000)
    })

    /** Dialog element for set deletion */
    let dialogElement = $state<HTMLDialogElement>()

    /** Shows the set delete dialog */
    function showDialog() {
        dialogElement?.showModal()
    }

    /** Closes the set delete dialog */
    function closeDialog() {
        dialogElement?.close()
    }

    /**
     * Sends API request to delete the set
    */
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