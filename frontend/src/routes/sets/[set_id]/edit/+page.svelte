<script lang="ts">
    import Loader from "$lib/components/Loader.svelte";
    import type { Card, Set } from "$lib/types/types";
    import { onMount } from "svelte";

    let {data} = $props()

    let isLoading = $state(false)
    let isProcessingQueue = $state(false)
    let setLocal = $state<Set|undefined>(undefined)
    let setRemote = $state<Set|undefined>(undefined)

    let synced = $derived.by(() => {
        if (isProcessingQueue) {
            return false
        }
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
        if (data.set) {
            setRemote = JSON.parse(JSON.stringify(data.set))
            setLocal = JSON.parse(JSON.stringify(data.set))
        }
    })

    let localNewCardIndex = $state(-1)

    function addCardLocal() {
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
            } else {
                setLocal.cards = [newCard]
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
    let nameUpdate = $state(false)

    function updateCard(id: number) {
        console.log("update: ", id)
        if (setLocal && setRemote && setLocal.cards && setRemote.cards) {
            if (!cardsToUpdate.includes(id)) {
                const cardLocal = setLocal.cards.find(card => card.id == id)
                const cardRemote = setRemote.cards.find(card => card.id == id)
                if (!cardLocal) {
                    // BAD
                    console.log("card not found")
                    return
                } else if (!cardRemote) {
                    // Created!
                    console.log("new card!")
                    cardsToUpdate.push(id)
                    console.log($state.snapshot(cardsToUpdate))
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

    function updateName() {
        nameUpdate = true
    }

    interface CardUpdate {
        id?: number
        front: string
        back: string
    }

    interface SetUpdate {
        name?: string
        description?: string
        cards?: CardUpdate[]
    }
    
    async function update() {
        if (setLocal && setRemote) {
            console.log("update")
            var u: SetUpdate = {}
            if (nameUpdate) {
                // add name to update
                u.name = setLocal.name
            }
            if (cardsToUpdate.length !== 0) {
                u.cards = []
                // add cards to update
                for (let i = 0; i <= cardsToUpdate.length; i++) {
                    if (setLocal.cards) {
                        const cardLocal = setLocal.cards.find((card) => card.id == cardsToUpdate[i])
                        console.log(cardLocal)
                        if (cardLocal && cardLocal.id < 0) {
                            // new card
                            const newCard = {
                                front: cardLocal.front || "",
                                back: cardLocal.back || ""
                            }
                            console.log("adding new card: ", newCard)
                            u.cards.push(newCard)
                        } else {
                            // existing card
                            u.cards.push({
                                id: setLocal.cards[i].id,
                                front: setLocal.cards[i].front || "",
                                back: setLocal.cards[i].back || ""
                            })
                        }
                    }
                } 
            }
                    
            if (u.name || u.description || u.cards) {
                try {
                    console.log("SENDING BODY:")
                    console.log(u)
                    const res = await fetch(`http://localhost:8080/sets/${setRemote?.id}`, {
                        method: "PATCH",
                        credentials: "include",
                        body: JSON.stringify(u)
                    })
                    if (!res.ok) {
                        console.log(await res.text())
                        return
                    }
                    const newRemote = await res.json() as Set
                    nameUpdate = false
                    cardsToUpdate = []
                    console.log(newRemote)
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
                } catch (e) {
                    console.log(e)
                }
            }
        }
    }

    const debouncedUpdate = debounce(update, 900)

    onMount(() => {
        setInterval(debouncedUpdate, 1000)
    })

</script>

{#if data.set}
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
                {#each setLocal.cards as card, index}
                <div class="card" role="listitem">
                        <span>{`${card.id}. `}</span>
                        <input type="text" placeholder="front" bind:value={card.front} oninput={() => updateCard(card.id)}>
                        <input type="text" placeholder="back" bind:value={card.back} oninput={() => updateCard(card.id)}>
                        <button>del</button>
                </div>
                {/each}
            {/if}
            <button onclick={addCardLocal}>new</button>
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
    #title {
        display: flex;
        flex-direction: row;
        align-items: center;
    }

    #title>h2 {
        margin-right: 1rem;
    }
</style>