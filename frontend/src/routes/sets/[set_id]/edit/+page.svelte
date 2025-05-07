<script lang="ts">
    import Loader from "$lib/components/Loader.svelte";
    import type { Card, Set } from "$lib/types/types";
    import { onMount } from "svelte";

    let isLoading = $state(false)

    let { data } = $props()

    let setLocal = $state<Set|undefined>(data.set)
    let setRemote = $state<Set|undefined>(data.set)
    let synced = $derived(setLocal == setRemote)

    $effect(() => {
        console.log(setLocal)
    })

    interface Update {
        handle(): Promise<Response>
    }

    class TitleUpdate implements Update {
        private newTitle: string

        constructor(newTitle: string) {
            this.newTitle = newTitle
        }

        public async handle(): Promise<Response> {
            return await fetch(`http://localhost:8080/sets/${setLocal?.id}`, {
                method: "PATCH",
                credentials: "include",
            })
        }

    }

    class CardUpdate implements Update {
        private newFront: string
        private newBack: string

        constructor(newFront: string, newBack: string) {
            this.newFront = newFront
            this.newBack = newBack
        }

        public async handle(): Promise<Response> {
            return await fetch(`http://localhost:8080/cards`)
        }

    }

    let queue = $state<Update[]>([])
    let isProcessingQueue = $state(false)

    async function processQueue() {
        
    }

    onMount(async () => {
        if (data.set) {
            setRemote = setLocal = data.set
        }
    })

    async function handleUpdateTitle(e: Event) {
        const inputElement = e.target as HTMLInputElement
        const newTitle = inputElement.value
    }

</script>

<div id="title">
    <h2>edit set</h2>
    {#if isLoading}
        <Loader />
    {/if}
</div>

<div id="create_frame">
    {#if setLocal}
    <form>
        <label>
            title <br />
            <input type="text" placeholder="title" bind:value={setLocal.name} onchange={handleUpdateTitle}>
        </label>
        
        <br />
        {#if setLocal.cards}
            {#each setLocal.cards as card, index}
            <div class="card" draggable="true"
                role="listitem">
                    <span>{`${index + 1}. `}</span>
                    <input type="text" bind:value={card.front} placeholder="front">
                    <input type="text" bind:value={card.back} placeholder="back">
                    <button>-</button>
            </div>
            {/each}
        {/if}
    </form>
    {/if}
</div>

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