<script lang="ts">
    import Loader from "$lib/components/Loader.svelte";
    import type { Card, Set } from "$lib/types/types";
    import { onMount } from "svelte";

    let { data } = $props()

    let isLoading = $state(false)
    let queue = $state<Update[]>([])
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
            console.log("cards")
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

    interface Update {
        type: string
    }

    interface CardUpdate extends Update {
        type: "card"
        id: number
        newFront?: string
        newBack?: string
    }

    interface NameUpdate extends Update {
        type: "name"
        newTitle: string
    }

    var blankCard: Card = {
        id: -1,
        set_id: -1,
        created: new Date(),
        front: "",
        back: ""
    }

    

    async function processQueue() {
        isProcessingQueue = true
        
        for (var i = 0; i < queue.length; i++) {
            if (queue[i].type == "title") {

            }
        }
    }

    onMount(async () => {
        if (data.set) {
            setRemote = data.set
            setLocal = data.set
        }
    })

    function addCardLocal() {
        if (setLocal?.cards) {
            setLocal.cards.push(blankCard)
        }
    }

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
                <input type="text" placeholder="title" bind:value={setLocal.name}>
            </label>
            
            <br />
            {#if setLocal.cards}
                {#each setLocal.cards as card, index}
                <div class="card" role="listitem">
                        <span>{`${index + 1}. `}</span>
                        <input type="text" bind:value={card.front} placeholder="front">
                        <input type="text" bind:value={card.back} placeholder="back">
                        <button>del</button>
                </div>
                {/each}
                <button onclick={addCardLocal}>new</button>
            {/if}
        </form>
        {/if}
    </div>
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