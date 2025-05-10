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
                        cardsToUpdate.filter((i) => {i !== id})
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

    interface CardUpdate {
        type: "create" | "update" | "delete"
        id?: number
        front?: string
        back?: string
    }

    interface SetUpdate {
        name?: string
        description?: string
        cards?: CardUpdate[]
    }
    
    async function update() {
        if (setLocal && setRemote) {
            var u: SetUpdate = {}
            if (nameUpdate) {
                // add name to update
                u.name = setLocal.name
            }
            u.cards = []
            if (cardsToUpdate.length !== 0) {
                // add cards to update
                for (const cardID of cardsToUpdate) {
                    if (setLocal.cards) {
                        const cardLocal = setLocal.cards.find((card) => card.id == cardID)
                        if (cardLocal && cardLocal.id < 0) {
                            // new card
                            const newCard: CardUpdate = {
                                type: "create",
                                front: cardLocal.front || "",
                                back: cardLocal.back || ""
                            }
                            u.cards.push(newCard)
                        } else {
                            // existing card
                            if (cardLocal) {
                                u.cards.push({
                                    type: "update",
                                    id: cardLocal.id,
                                    front: cardLocal.front || "",
                                    back: cardLocal.back || ""
                            })
                            }
                            
                        }
                    }
                } 
            }

            if (cardsToDelete.length !== 0) {
                for (const id of cardsToDelete) {
                    u.cards?.push({
                        type: "delete",
                        id: id,
                    })
                }
            }
                    
            if (u.name || u.description || u.cards.length > 0) {
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
                    await invalidate((url) => url.pathname === `/sets/${data.set?.id}`)
                    const newRemote = await res.json() as Set
                    nameUpdate = false
                    cardsToUpdate = []
                    cardsToDelete = []
                    setRemote = newRemote
                    if (setRemote.cards && setLocal.cards) {
                        if (setLocal.cards.length == setRemote.cards.length) {
                            for (let i = 0; i < setLocal.cards.length; i++) {
                                if (setLocal.cards[i].id < 0) {
                                    setLocal.cards[i].id = setRemote.cards[i].id
                                }
                            }
                        }
                    }
                    isLoading = false
                } catch (e) {
                    console.log(e)
                    isLoading = false
                }
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
        <p>are you sure you want to delete this set?</p>
        <button onclick={deleteSet}>yes</button>
        <button onclick={closeDialog}>no</button>
    </dialog>

    <a href={`/sets/${data.set.id}`}>back</a>
    <button onclick={showDialog}>delete set</button>
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
                title <br />
                <input type="text" placeholder="name" bind:value={setLocal.name} oninput={updateName}>
            </label>
            
            <br />
            {#if setLocal.cards}
                <label for="card">cards</label>
                {#each setLocal.cards as card, index}
                <div class="card" role="listitem">
                        <input type="text" placeholder="front" bind:value={card.front} oninput={() => updateCard(card.id)}>
                        <input type="text" placeholder="back" bind:value={card.back} oninput={() => updateCard(card.id)}>
                        <button onclick={() => deleteCard(card.id)}>del</button>
                </div>
                {/each}
            {/if}
            <button onclick={addCard}>new</button>
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
        background-color: var(--col-purplegrey);
        color: var(--col-lightpink)
    }

    ::backdrop {
        background-color: black;
        opacity: 0.5;
    }
    #title {
        display: flex;
        flex-direction: row;
        align-items: center;
    }

    #title>h2 {
        margin-right: 1rem;
    }
</style>